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
- Store GatewayとObject Storageは１対１の設定で、複数のObject Storageがある場合はObject Storageの数の分Store Gatewayが必要  
  ![](./image/StoreGateway_1.jpg)
- StoreGatewayはローカルディスクをそこまで必要とせず、再起動などでデータが削除されても起動時間が増加するくらいでそこまで影響はない
  > It acts primarily as an API gateway and therefore does not need significant amounts of local disk space. It joins a Thanos cluster on startup and advertises the data it can access. It keeps a small amount of information about all remote blocks on local disk and keeps it in sync with the bucket. This data is generally safe to delete across restarts at the cost of increased startup times.
  > In general, an average of 6 MB of local disk space is required per TSDB block stored in the object storage bucket, but for high cardinality blocks with large label set it can even go up to 30MB and more. It is for the pre-computed index, which includes symbols and postings offsets as well as metadata JSON.

## Querier (Query)
- https://thanos.io/tip/components/query.md/
- Querier is fully **stateless** and horizontally scalable.
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
- **以下公式サイトから抜粋**
```
## Deduplication 
The query layer can deduplicate series that were collected from high-availability pairs of data sources such as Prometheus. 
A fixed single or multiple replica labels must be chosen for the entire cluster and can then be passed to query nodes on startup.

Two or more series that are only distinguished by the given replica label, will be merged into a single time series.
This also hides gaps in collection of a single data source.

## An example with a single replica labels: 
Prometheus + sidecar “A”: cluster=1,env=2,replica=A
Prometheus + sidecar “B”: cluster=1,env=2,replica=B
Prometheus + sidecar “A” in different cluster: cluster=2,env=2,replica=A
If we configure Querier like this:

thanos query \
    --http-address        "0.0.0.0:9090" \
    --query.replica-label "replica" \
    --endpoint            "<store-api>:<grpc-port>" \
    --endpoint            "<store-api2>:<grpc-port>" \
And we query for metric up{job="prometheus",env="2"} with this option we will get 2 results:

up{job="prometheus",env="2",cluster="1"} 1
up{job="prometheus",env="2",cluster="2"} 1
WITHOUT this replica flag (deduplication turned off), we will get 3 results:

up{job="prometheus",env="2",cluster="1",replica="A"} 1
up{job="prometheus",env="2",cluster="1",replica="B"} 1
up{job="prometheus",env="2",cluster="2",replica="A"} 1
```

## Query Frontend
- https://thanos.io/tip/components/query-frontend.md/
- Query Frontend is fully **stateless** and horizontally scalable.

## Compactor
- https://thanos.io/tip/components/compact.md/
- Down Sampling、Retention、Compactionを担当するコンポーネント
- defaultではobject storage上のデータの保持期間はない(無期限保持)
- データ削除は`--retention.resolution-raw`、`--retention.resolution-5m`、`--retention.resolution-1h`の３つのフラグで設定できる。この３つを設定しなかったり`0s`に設定するとデータは無期限保存される。  
  > You can configure retention by using `--retention.resolution-raw` `--retention.resolution-5m` and `--retention.resolution-1h` flag. Not setting them or setting to 0s means no retention.
- **Retention is applied right after Compaction and Downsampling loops. If those are failing, data will never be deleted.**
- *Compaction*
  - responsible for **compacting multiple blocks into one to reduce the number of blocks and compact index indices.** We can compact an index quite well in most cases, because series usually live longer than the duration of the smallest blocks (2 hours).
  - https://thanos.io/tip/components/compact.md/#compaction
- **Compactorは1つのObject Storageごとに1つのみ動かす必要がある**
  - https://thanos.io/tip/components/compact.md/#warning-only-one-instance-of-compactor-may-run-against-a-single-stream-of-blocks-in-a-single-object-storage
- HA構成のPrometheusからのメトリクスをCompactor側でもdedupすることができる
  - https://thanos.io/tip/components/compact.md/#vertical-compaction-use-cases
  - でもリスクがあるらしく、あまり使わない方が良さそう？
- defaultではCompactorはcronjobとして動かせるように処理が終わったらCompletedになってしまうため、継続的に実行させるためには`--wait`と`--wait-interval=5m`フラグを付ける必要がある

## Store API
- https://thanos.io/tip/thanos/integrations.md/#storeapi
- https://thanos.io/tip/thanos/quick-tutorial.md/#store-api