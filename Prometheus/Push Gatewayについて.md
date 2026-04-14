# Push Gatewayとは
- Push Gatewayは、Prometheusが直接スクレイピングできない短命のジョブやバッチジョブからメトリクスを収集するためのコンポーネント。短命のジョブからメトリクスをプッシュするためのエンドポイントを提供し、Prometheusは定期的にPush Gatewayからこれらのメトリクスをスクレイピングする。

# Push Gatewayの設定について
- https://github.com/prometheus/pushgateway
- defaultではPush Gatewayはメトリクスを永続化しないため、Push Gatewayを再起動するとすべてのメトリクスが失われる。永続化を有効にするには、`--persistence.file`フラグを使用して永続化ファイルのパスを指定する必要がある。  
  > By default, Pushgateway does not persist metrics. However, the --persistence.file flag allows you to specify a file in which the pushed metrics will be persisted (so that they survive restarts of the Pushgateway).

## Push Gatewayのマニフェストファイル例
```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pushgateway
  namespace: monitoring
  labels:
    app: pushgateway
spec:
  replicas: 1
  selector:
    matchLabels:
      app: pushgateway
  template:
    metadata:
      labels:
        app: pushgateway
    spec:
      containers:
      - name: pushgateway
        image: prom/pushgateway:v1.11.2
        args:
          - --persistence.file=/data/pushgateway.data
          - --persistence.interval=5m
          - --web.enable-admin-api  # DELETE API など、Admin APIを使う場合に必要
        ports:
        - name: http
          containerPort: 9091
        volumeMounts:
        - name: data
          mountPath: /data
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
          limits:
            memory: 256Mi
        livenessProbe:
          httpGet:
            path: /-/healthy
            port: http
        readinessProbe:
          httpGet:
            path: /-/ready
            port: http
      volumes:
      - name: data
        emptyDir: {}  # 永続化したい場合は PVC に変更
---
apiVersion: v1
kind: Service
metadata:
  name: pushgateway
  namespace: monitoring
  labels:
    app: pushgateway
spec:
  type: ClusterIP
  ports:
  - name: http
    port: 9091
    targetPort: 9091
  selector:
    app: pushgateway
```

> [!NOTE]
> - `--persistence.interval` は、Pushgatewayがメモリ上のメトリクスをディスク（`--persistence.file` で指定したファイル）に書き出す間隔を指定するフラグ。デフォルトは5分（5m）。
> - Pushgatewayは受け取ったメトリクスを基本的にメモリ上に保持する。そのままだとPod再起動やクラッシュですべてのメトリクスが失われるため、定期的にディスクにスナップショットを書き出し、起動時にそのファイルから復元するようになっている。

## Push GatewayのメトリクスをPrometheusでスクレイピングする方法
- Prometheusのscrape config例  
  ```yaml
  scrape_configs:
    - job_name: 'pushgateway'
      honor_labels: true  # 必須：push側のjob/instanceを保持
      static_configs:
        - targets: ['pushgateway.monitoring.svc.cluster.local:9091']
  ```

> [!IMPORTANT]
> - `honor_labels: true` を設定しないと、Push Gateway側の `job`、`instance` ラベルが `exported_job`、`exported_instance` にリネームされてしまい、Push Gatewayにプッシュしたときの `job`、`instance` ラベルはPrometheus側のConfigで指定した `job`、`instance` ラベルに上書きされてしまうため、Push Gatewayにプッシュしたときの `job`、`instance` ラベルを保持したい場合は必ず `honor_labels: true` を設定する必要がある。
> - https://github.com/prometheus/pushgateway?tab=readme-ov-file#about-the-job-and-instance-labels  
>  > ### About the job and instance labels
>  > The Prometheus server will attach a `job` label and an `instance` label to each scraped metric. The value of the `job` label comes from the scrape configuration. When you configure the Pushgateway as a scrape target for your Prometheus server, you will probably pick a job name like `pushgateway`. The value of the `instance` label is automatically set to the host and port of the target scraped. Hence, all the metrics scraped from the Pushgateway will have the host and port of the Pushgateway as the `instance` label and a `job` label like `pushgateway`. The conflict with the `job` and `instance` labels you might have attached to the metrics pushed to the Pushgateway is solved by renaming those labels to `exported_job` and `exported_instance`.
>  >
>  > However, this behavior is usually undesired when scraping a Pushgateway. Generally, you would like to retain the `job` and `instance` labels of the metrics pushed to the Pushgateway. That's why you have to set `honor_labels: true` in the scrape config for the Pushgateway. It enables the desired behavior. See the [documentation](https://prometheus.io/docs/operating/configuration/#scrape_config) for details.
>  >
>  > This leaves us with the case where the metrics pushed to the Pushgateway do not feature an `instance` label. This case is quite common as the pushed metrics are often on a service level and therefore not related to a particular instance. Even with `honor_labels: true`, the Prometheus server will attach an `instance` label if no `instance` label has been set in the first place. Therefore, if a metric is pushed to the Pushgateway without an instance label (and without instance label in the grouping key, see below), the Pushgateway will export it with an empty instance label (`{instance=""}`), which is equivalent to having no `instance` label at all but prevents the server from attaching one.

# Push Gatewayへのメトリクスのプッシュ方法
## curlコマンドを使う方法
- 基本フォーマット  
  ```bash
  PUT/POST http://<pushgateway>:9091/metrics/job/<JOB>[/<LABEL>/<VALUE>]...
  ```
- 例  
  ```shell
  echo "metric_1 3.14" | curl --data-binary @- http://localhost:9091/metrics/job/my_job
  ```

  ```shell
  # POST: 同じ名前のメトリクスだけ上書き、他は残る
  echo "batch_job_duration_seconds 42.3" | curl --data-binary @- \
    http://pushgateway:9091/metrics/job/nightly_etl/instance/pod-xyz

  # PUT: このグルーピングキー配下を総入れ替え
  echo "batch_job_duration_seconds 42.3" | curl -X PUT --data-binary @- \
    http://pushgateway:9091/metrics/job/nightly_etl/instance/pod-xyz
  ```

- **POST**: 同じラベルのメトリクスのみ上書き（既存のグルーピングキーの他のメトリクスは残る）
- **PUT**: そのグルーピングキー配下の全メトリクスを置き換え
- **DELETE**: 削除

> [!IMPORTANT]
> ## POSTとPUTの違い
> - Pushgatewayでは「グルーピングキー」という単位でメトリクスが管理されている。グルーピングキーは、`job` ラベルと任意の数の追加ラベル（例: `instance`、`env` など）で構成される。
> 例えば、`job="nightly_etl", instance="pod-xyz"` というグルーピングキーがあり、Pushgatewayに既に以下が入っているとする:  
> ```
> [箱: job=nightly_etl, instance=pod-xyz]
>   ├─ batch_job_duration_seconds = 42.3
>   ├─ batch_job_processed_total  = 100
>   └─ batch_job_errors_total     = 2
> ```
> ここに `batch_job_duration_seconds = 50.0` だけを送った場合：
> #### POSTの場合（送ったメトリクスだけ上書き）
> ```
> [箱: job=nightly_etl, instance=pod-xyz]
>   ├─ batch_job_duration_seconds = 50.0   ← 上書きされた
>   ├─ batch_job_processed_total  = 100    ← そのまま残る
>   └─ batch_job_errors_total     = 2      ← そのまま残る
> ```
> #### PUTの場合（箱ごと総入れ替え）
> ```
> [箱: job=nightly_etl, instance=pod-xyz]
>   └─ batch_job_duration_seconds = 50.0   ← これだけ
>     (他のメトリクスは消える)
> ```
> ### 使い分け
> - **PUT (`push_to_gateway`)**: 「このジョブのメトリクスはこれで全部です。前回の情報はもう要りません」と言いたいとき。バッチジョブの最後にまとめて送信する場合はこちらが自然。
> - **POST (`pushadd_to_gateway`)**: 「既存のメトリクスはそのままで、今送ったものだけ追加・更新してください」と言いたいとき。複数の処理から別々のメトリクスを同じグルーピングキーに送る場合などに使う。

> [!IMPORTANT]
> ## グルーピングキーと通常ラベルの違い
> Pushgatewayには2種類の「ラベル的なもの」があり、混同しやすいので注意が必要。
>
> | 種類 | どこで決まる | 役割 |
> |---|---|---|
> | **グルーピングキー** | URLパスの `/job/xxx/instance/yyy/...`<br>(Pythonなら `grouping_key` 引数) | Pushgateway内部の「箱」の識別子。POST/PUT/DELETEの単位。 |
> | **通常のメトリクスラベル** | メトリクス本文の `{sid="...", type="..."}`<br>(Pythonなら `labelnames`) | Prometheusの時系列を分けるラベル。 |
>
> ### POST/PUTの上書き判定はグルーピングキー単位
> POST/PUTの上書き判定は**グルーピングキー**でのみ行われ、通常ラベル（`sid`, `type` など）は一切関係しない。
>
> 例えば、Pushgatewayに既に以下が入っているとする:
> ```
> [箱: job=myapp]
>   ├─ events_total{sid="A", type="click"} = 10
>   ├─ events_total{sid="A", type="view"}  = 50
>   └─ events_total{sid="B", type="click"} = 3
> ```
> ここに POST で `events_total{sid="A", type="click"} 99` を送ると、結果は:
> ```
> [箱: job=myapp]
>   └─ events_total{sid="A", type="click"} = 99   ← これだけ残る
>      (他は全部消える)
> ```
> つまりPOSTは「同じグルーピングキー + 同じメトリクス名(メトリクスファミリ)」の時系列を**まるごと**置き換える。ラベルの組み合わせごとに個別に上書きされるわけではない。
>
> | 操作 | 上書きされる範囲 |
> |---|---|
> | POST (`pushadd_to_gateway`) | 同じグルーピングキー配下の、**同じメトリクス名**の時系列すべて |
> | PUT (`push_to_gateway`) | 同じグルーピングキー配下の**全メトリクス** |
> | DELETE | 同じグルーピングキー配下の全メトリクス |

> [!NOTE]
> ## グルーピングキーの構成
> - **`job` は必須**。URLの `/metrics/job/<JOB>` 部分は省略できない。
> - **それ以外は任意**。`instance` は慣習的によく使われる例に過ぎず、`/key/value/key/value/...` の形で好きなラベルを好きな数だけ追加できる。
> - グルーピングキーの順序はPushgateway側でラベル名ソートされて正規化されるため、`/job/etl/env/prod/region/apne1` と `/job/etl/region/apne1/env/prod` は同じ箱として扱われる。
>
> ### グルーピングキーに入れるべきもの / 入れるべきでないもの
> グルーピングキーの組み合わせごとに別々の「箱」が作られ、明示的にDELETEしない限り残り続けるため、カーディナリティに注意する必要がある。
>
> - ✅ 向いているもの: `job`, `instance`, `pod`, `env`, `region`, `cluster`, `shard` など**固定的で値の種類が限られるもの**
> - ❌ 向かないもの: ユーザーID、セッションID、リクエストID、タイムスタンプなど**動的に無数に発生するもの** → これらは `labelnames` 側（通常ラベル）に入れる

## Prometheus Client Libraryを使う方法（Python例）
- https://github.com/prometheus/client_python
- コード例  
  ```python
  from prometheus_client import CollectorRegistry, Gauge, Counter
  from prometheus_client import push_to_gateway, pushadd_to_gateway, delete_from_gateway
  import time

  PUSHGATEWAY = 'pushgateway.monitoring.svc.cluster.local:9091'
  JOB = 'nightly_etl'
  GROUPING = {'instance': 'pod-xyz', 'env': 'prod'}

  # --- 1. バッチジョブ本体のメトリクス（PUTで総入れ替え） ---
  registry = CollectorRegistry()
  duration = Gauge('batch_job_duration_seconds', 'Job duration', registry=registry)
  processed = Counter('batch_job_processed_total', 'Processed items', registry=registry)
  last_success = Gauge('batch_job_last_success_timestamp_seconds','Last success timestamp', registry=registry)

  start = time.time()
  # ... 実処理 ...
  processed.inc(100)
  duration.set(time.time() - start)
  last_success.set_to_current_time()

  # PUT: このグルーピングキー配下を total 入れ替え
  push_to_gateway(PUSHGATEWAY, job=JOB, registry=registry, grouping_key=GROUPING)

  # --- 2. 後から追加のメトリクスだけ送る(POSTで追記) ---
  extra_registry = CollectorRegistry()
  rows_exported = Gauge('batch_job_rows_exported', 'Rows exported to S3',registry=extra_registry)
  rows_exported.set(12345)

  # POST: 上で送った duration/processed/last_success はそのまま残り、rows_exported だけが箱に追加される
  pushadd_to_gateway(PUSHGATEWAY, job=JOB, registry=extra_registry, grouping_key=GROUPING)

  # --- 3. ジョブが不要になったらグルーピングキーごと削除 ---
  # delete_from_gateway(PUSHGATEWAY, job=JOB, grouping_key=GROUPING)
  ```

- `push_to_gateway` は **PUT**、`pushadd_to_gateway` は **POST** に相当

> [!NOTE]
> `inc()` は Counter の値をインクリメントするメソッド。引数なしで呼ぶと1ずつ増える。引数に数値を渡すとその数だけ増える。`set()` は Gauge の値を指定した数値にセットするメソッド

> [!NOTE]
> ## `Counter`と`Gauge`の引数
> #### 第1引数: name（メトリクス名）
> - 必須。Prometheusのメトリクス名。  
>   ```python
>   Counter('http_requests_total', ...)
>   Gauge('queue_size', ...)
>   ```
> - Counterには慣習的に `_total` を付ける
> #### 第2引数: documentation（説明文）
> - 必須。Prometheusの exposition formatの `# HELP` 行になる。  
>   ```python
>   Counter('http_requests_total', 'Total number of HTTP requests')
>   Gauge('queue_size', 'Current size of the queue')
>   ```
> #### `labelnames`（ラベル名）
> - ラベル名のリスト/タプル。ここで宣言したラベルは、値を記録するときに `.labels()` で指定する。  
>   ```python
>   requests = Counter(
>     'http_requests_total',
>     'Total HTTP requests',
>     labelnames=['method', 'status', 'endpoint'],
>   )
>
>   # キーワード引数で指定
>   requests.labels(method='GET', status='200', endpoint='/api/users').inc()
>
>   # 位置引数でもOK（labelnamesで宣言した順）
>   requests.labels('POST', '500', '/api/login').inc()
>   ```

# 運用上の注意点
- **複数起動してもHA構成にならない**
  - 複数レプリカで動かしても同期しないので、基本的にレプリカ1で運用することが多い
- **メトリクスは自動削除されない**
  - 一度pushされたメトリクスは、明示的にDELETEするかPushgateway自体を再起動するまで残り続ける。ジョブ終了時に古いメトリクスを消す運用が必要なこともある。
- **タイムスタンプを持たない**
  - Pushgatewayにプッシュされたメトリクスはタイムスタンプを持たないため、Prometheusがスクレイピングした時点のタイムスタンプが付与される。ジョブの実行時間など、タイムスタンプを持たせたい場合は、メトリクスの値としてUnixタイムスタンプをプッシュするなどの工夫が必要。