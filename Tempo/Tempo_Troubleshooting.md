## metrics-generatorを有効にしたらTempoが起動しなくなった件
### 事象
- Tempo v2.1.1でmonolithic modeで動かして問題なかったけど、`metrics_generator`設定を追加したらpanicが起きてTempoが起動しなくなった
  - Tempo Logs  
    ~~~
    github.com/grafana/tempo/modules/generator/registry/registry.go:122 +0x9db
    created by github.com/grafana/tempo/modules/generator/registry.New
    github.com/grafana/tempo/modules/generator/registry/job.go:11 +0x4c
    github.com/grafana/tempo/modules/generator/registry.job({0x29a8920, 0xc00016e140}, 0xc001e46450, 0xc001e46460)
    time/tick.go:24 +0x10f
    time.NewTicker(0x29a8920?)
    goroutine 2376 [running]:
    panic: non-positive interval for NewTicker
    ~~~
  - Configuration  
    ~~~
    server:
      http_listen_port: 3200
      
    distributor:
      receivers:
          otlp:
            protocols:
              http:
              grpc:
      
    compactor:
      compaction:
        block_retention: 744h                # configure total trace retention here
      
    storage:
      trace:
        backend: s3
        s3:
          endpoint: s3.ap-northeast-1.amazonaws.com
          bucket: <S3 Bucket Name>
          forcepathstyle: true
          #set to true if endpoint is https
          insecure: true
        wal:
          path: /tmp/tempo/wal         # where to store the the wal locally
        local:
          path: /tmp/tempo/blocks
      
    overrides:
      metrics_generator_processors:
        - span-metrics
      
    metrics_generator:
      ring:
        kvstore:
      processor:
        service_graphs:
        span_metrics:
          intrinsic_dimensions:
          dimensions:
            - "db.statement"
      registry:
      storage:
        path: /opt/tempo/wal
        wal:
        remote_write:
          - url: <Remote Write URL>      
    ~~~

### 原因
- `metrics_generator.registry`がemptyだとその配下の項目(e.g. `collection_interval`)の設定値がdefault値になるのではなく、全部`0`が設定されるとのこと  
    ![](img/registry_trouble.jpg)

### 対処
- `registry` blockを削除するか、以下のように明示的に`registry`配下の項目を設定する  
  ~~~yaml
  metrics_generator:
      registry:
          collection_interval: 15s
          stale_duration: 15m
          max_label_name_length: 1024
          max_label_value_length: 2048
  ~~~

---

## GrafanaのDrildownでTrace （Tempo） を選択するとエラーが発生する
### 事象
- GrafanaのDrilldonwでTrace (Tempo) を選択すると以下のエラーが発生する
  ```
  level=warn ts=2026-02-03T03:13:25.758261042Z caller=server.go:2294 msg="GET /querier/api/metrics/query_range?start=1770086605755779972&end=1770088432000000000&step=28s&mode=recent&blockID=&startPage=0&pagesToSearch=0&version=&encoding=&size=0&footerSize=0&q=%7BnestedSetParent%3C0+%26%26+true+%26%26+status%3Derror%7D+%7C+rate%28%29++with%28sample%3Dtrue%29&exemplars=99&maxSeries=0 (500) 79.245µs"
  ```

![tempo_drilldown_emptyring](img/tempo_drilldown_emptyring.png)

### 原因
- 上記のログにも出ているように、DrilldownではTempoのメトリクスクエリを実行している。しかし、デフォルトの状態ではmetricsGeneratorが無効になっており、メトリクスクエリを実行できないためエラーとなっている。**なので、metricsGeneratorを有効にする必要がある**。(metrics-generatorのPodが作成される)
- 参考URL
  - https://grafana.com/docs/tempo/latest/configuration/
  - https://grafana.com/docs/tempo/latest/metrics-from-traces/metrics-queries/configure-traceql-metrics/
  - https://github.com/grafana/tempo/issues/5491
  - https://github.com/grafana/helm-charts/tree/main/charts/tempo-distributed  
    > ## Activate metrics generator
    > Metrics-generator is disabled by default and can be activated by configuring the following values:
    > ```yaml
    > metricsGenerator:
    >   enabled: true
    >   replicas: 2
    >   config:
    >     storage:
    >       remote_write:
    >       - url: http://cortex/api/v1/push
    >         send_exemplars: true
    >    #    headers:
    >    #      x-scope-orgid: operations
    > # Global overrides
    > overrides:
    >   defaults:
    >     metrics_generator:
    >       processors:
    >         - service-graphs
    >         - span-metrics

### 対処
- `metricsGenerator`を有効にし、`registry`と`storage`の設定を追加する。以下はCortex-tenantにRemote Writeする例。あと、`overrides`でprocessorsを指定するのも忘れずに。  
  ```yaml
  tempo:
    iamge:
      ・・・

  metricsGenerator:
    enabled: true
    config:
      registry:
        collection_interval: 30s
        external_labels:
          source: tempo
        inject_tenant_id_as: tenant # Tenant IDが「tenant」ラベルに設定される
      storage:
        path: /var/tempo/wal
        remote_write:
          - url: http://cortex-tenant.monitoring.svc:8080/push # Cortex-tenant側で
            send_exemplars: true
        remote_write_add_org_id_header: true

  overrides:
    defaults:
      metrics_generator:
        processors:
          - service-graphs
          - span-metrics
          - local-blocks # Grafana DrillDownが内部で使う TraceQL metrics クエリ（rate(), count_over_time() 等）には local-blocks が必須
  ```

> [!IMPORTANT]  
> - Tempoがマルチテナントモードの場合、`remote_write_add_org_id_header: true`を設定すると、Headerに`X-Scope-OrgID`が追加される
> - ただ、Remote Write先がThanosなど、テナントを識別するHeader名が異なる場合は、`inject_tenant_id_as`でテナント識別用のラベルを指定した上で、Remote Write先としてcortex-tenantを指定して、cortex-tenant側でそのラベルを元にThanos用のHeaderに変換するように設定する必要がある

## Tempoのクエリで`rpc error: code = ResourceExhausted desc = grpc: received message larger than max (6698514 vs. 4194304)`エラーが出る
- 事象
  - GrafanaでTempoのクエリを実行したら以下の添付のようなエラーが出る  
  ![grpc_error](./img/grpc_error.png)

  - あと、TempoのQuerierからは以下のようなエラーが出ていた  
    ```shell
    level=error ts=2026-02-18T10:36:16.534624984Z caller=frontend_processor.go:84 msg="error processing requests" address=10.23.13.124:9095 err="rpc error: code = ResourceExhausted desc = grpc: received message after decompression larger than max 4194304"
    ```
- 原因
  - レスポンスサイズがgRPCのデフォルトの最大値の4MB（4194304 bytes）を超えていたため、gRPCのResourceExhaustedエラーが出ていた

> [!NOTE]  
> https://grafana.com/docs/tempo/latest/configuration/#server
> - Tempo自体の `server`ブロックの`grpc_server_max_recv_msg_size`と`grpc_server_max_send_msg_size`の値は `16777216`になっているけど、なぜかHelm側（https://github.com/grafana-community/helm-charts/tree/main/charts/tempo-distributed） はデフォルト値として、`4194304`を設定している。なので、Helmのvalues.ymlで明示的に大きくする必要がある。

- 対処
  - `server`ブロックの`grpc_server_max_recv_msg_size`と`grpc_server_max_send_msg_size`の値をクエリ結果のサイズに合わせて大きくする（以下は`values.yml`の例）  
    ```yaml
    server:
      grpc_server_max_recv_msg_size: 16777216 # 16MB
      grpc_server_max_send_msg_size: 16777216 # 16MB
    ```

---

## Tempoの`tempodb_blocklist_tenant_index_errors_total`メトリクスの件数（error）が増加していて、QuerierとQuery Frontendでエラーが出ている

### 事象
- Tempoの`tempodb_blocklist_tenant_index_errors_total`メトリクスの件数（error）が増加していて、QuerierとQuery Frontendで以下のようなエラーが出ている  
  ```shell
  level=error ts=2026-02-24T09:57:04.325960337Z caller=poller.go:291 msg="failed to pull bucket index for tenant. falling back to polling" tenant=plat err="does not exist"
  ```

### 原因
- 途中でテナント名（`X-Scope-OrgID`）を変更したため、S3上に古いテナント名のディレクトリが残っている。Querier/Query Frontendのポーリング処理がそのディレクトリを発見し、テナントインデックスを取得しようとするが存在しないため`tempodb_blocklist_tenant_index_errors_total`が増加し続けている
- `poller.go:149` の `Tenants()`がバックエンドストレージのルート直下の全ディレクトリ名をテナントとして列挙する（`tempo_cluster_seed.json`と`work.json` のみ除外）。特定テナントをスキップする設定は存在しない。

### 対処
- `storage.trace.empty_tenant_deletion_enabled`を`true`に設定して、空のテナントを削除するようにする。
  ```yaml
  storage:
    trace:
      empty_tenant_deletion_enabled: true
      empty_tenant_deletion_age: 12h # default: 12h
  ```

> [!NOTE]  
> ### ブロック
> - テナントディレクトリ配下の `<tenant>/<block-uuid>/` ディレクトリで、中に`meta.json`（通常ブロック）または `meta.compacted.json`（コンパクション済みブロック）を持つもの。1ブロック = 一定期間のトレースデータをまとめた単位。`Blocks()` は `<tenant>/` 配下を走査し、`<uuid>/meta.json` または `<uuid>/meta.compacted.json` にマッチするものだけをブロックとして返す。
> ### オブジェクト
> - テナントディレクトリ配下の全ファイル。`Find()` はフィルタなしで再帰的に全ファイルを列挙する。ブロックのデータファイル、テナントインデックスファイル（`index.json.gz`,`index.pb.zst`）、フラグファイル（`nocompact.flg`）など全てが含まれる。

> [!IMPORTANT]  
> ## 削除判定のフロー
> 1. pollTenantBlocks() でテナント配下のブロック一覧を取得
>    - `<tenant>/<uuid>/meta.json` → 通常ブロック
>    - `<tenant>/<uuid>/meta.compacted.json` → compactedブロック
> 2. ブロックが1つでもある場合 → `deleteTenant()` は呼ばれない。終了。
>    - 「ブロックが1つでもある」= `<uuid>/meta.json` または `<uuid>/meta.compacted.json` を持つ UUID ディレクトリが1つでも存在する
> 3. ブロックが0個の場合 → `deleteTenant()` が呼ばれる
> 4. `Find()` でテナントディレクトリ配下の全オブジェクト（全ファイル）を再帰走査
> 5. 各オブジェクトの最終更新日時を `empty_tenant_deletion_age` と比較
> 6. `empty_tenant_deletion_age` より新しいオブジェクトが1つでもある → 何もしない。終了。
> 7. 全オブジェクトが `empty_tenant_deletion_age` より古い →
> テナントインデックスが存在しないことを再確認した上で、全オブジェクトを削除