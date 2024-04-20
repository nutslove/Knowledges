## アーキテクチャ（Sidecar方式）
![](./image/Thanos_sidecar.jpg)

## Multi Tenancy
- https://thanos.io/tip/operating/multi-tenancy.md/
- **Thanosは色んなところでPrometheusの`global.external_labels`を使ってPrometheusインスタンスを識別するため、Prometheus側の`global.external_labels`の設定は必須である**
  - https://thanos.io/tip/thanos/quick-tutorial.md/#external-labels

## Sidecar
- https://thanos.io/tip/components/sidecar.md/
- PrometheusコンテナのSidecarコンテナとして起動され、定期的に (defaultでは2時間ごと) PrometheusのメトリクスをObject Storageに送信する
  - **Object Storageに送られる前にPrometheusがcrashすると直近のメトリクスデータがなくなるので、PV付きでPrometheusを実行すること！**
- Store APIも実装されており、オブジェクトストレージに送信されてない直近のメトリクスについてPrometheusにクエリーを投げる
- **`--storage.tsdb.min-block-duration`と`--storage.tsdb.max-block-duration`は必ず同じ値を設定しなければいけない。デフォルトでは`2h`が設定されていて、これが推奨値でもある。**  
  > The `--storage.tsdb.min-block-duration` and `--storage.tsdb.max-block-duration` must be set to equal values to disable local compaction in order to use Thanos sidecar upload, otherwise leave local compaction on if sidecar just exposes StoreAPI and your retention is normal. The default of 2h is recommended. Mentioned parameters set to equal values disable the internal Prometheus compaction, which is needed to avoid the corruption of uploaded data when Thanos compactor does its job, this is critical for data consistency and should not be ignored if you plan to use Thanos compactor. Even though you set mentioned parameters equal, you might observe Prometheus internal metric `prometheus_tsdb_compactions_total` being incremented, don’t be confused by that: Prometheus writes initial head block to filesystem via its internal compaction mechanism, but if you have followed recommendations - data won’t be modified by Prometheus before the sidecar uploads it. Thanos sidecar will also check sanity of the flags set to Prometheus on the startup and log errors or warning if they have been configured improperly
- Prometheusのメトリクスデータ保持期間は`min-block-duration`の3倍以上(6h以上)にすることが推奨されている
  - Object Storageの障害に備えて
- Thanos Sidecarに強制的にObject StorageにflushするようなAPIはない
  - https://github.com/thanos-io/thanos/issues/1849

## Store (Store Gateway)
- https://thanos.io/tip/components/store.md/

## Querier (Query)
- https://thanos.io/tip/components/query.md/
- HA構成のPrometheusで収集された重複メトリクスのdeduplication(重複排除)もQuerierが行う
  - `--query.replica-label`フラグでdedupのためのラベルを指定  
    ```shell
    thanos query \
        --http-address        0.0.0.0:19192 \
        --endpoint            1.2.3.4:19090 \
        --endpoint            1.2.3.5:19090 \
        --query.replica-label replica          # Replica label for deduplication
        --query.replica-label replicaX         # Supports multiple replica labels for deduplication
    ```
    - 複数の`--query.replica-label`がある場合、**OR条件**になる
    - 複数の`--query.replica-label`が与えられた場合、Thanos Queryはこれらのラベルのうちどれか一つが一致するデータセットを同じソースからのものと見なし、それらの間でデータの重複を解消する
  - **`global.external_labels`をもとにdedupを行うので、Prometheus側で`global.external_labels`が設定されている必要がある**
  - https://thanos.io/tip/thanos/quick-tutorial.md/#deduplicating-data-from-prometheus-ha-pairs

## Compactor
- https://thanos.io/tip/components/compact.md/
- Down Sampling、Retention、Compactionを担当するコンポーネント
- *Compaction*
  - responsible for **compacting multiple blocks into one to reduce the number of blocks and compact index indices.** We can compact an index quite well in most cases, because series usually live longer than the duration of the smallest blocks (2 hours).
  - https://thanos.io/tip/components/compact.md/#compaction
- **Compactorは1つのObject Storageごとに1つのみ動かす必要がある**
  - https://thanos.io/tip/components/compact.md/#warning-only-one-instance-of-compactor-may-run-against-a-single-stream-of-blocks-in-a-single-object-storage
- HA構成のPrometheusからのメトリクスをCompactor側でもdedupすることができる
  - https://thanos.io/tip/components/compact.md/#vertical-compaction-use-cases
  - でもリスクがあるらしく、あまり使わない方が良さそう？

## Store API
- https://thanos.io/tip/thanos/integrations.md/#storeapi
- https://thanos.io/tip/thanos/quick-tutorial.md/#store-api