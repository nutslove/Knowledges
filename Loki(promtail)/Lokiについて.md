## 参考URL
- https://taisho6339.hatenablog.com/entry/2021/05/26/104449
- https://taisho6339.gitbook.io/grafana-loki-deep-dive/
- https://speakerdeck.com/line_developers/grafana-loki-deep-dive

## Architecture
![Loki_Architecture](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/Loki_Architecture.jpg)  
出典: https://grafana.com/blog/2021/08/11/a-guide-to-deploying-grafana-loki-and-grafana-tempo-without-kubernetes-on-aws-fargate/

## Components
### Write path
- __Distributor__
- __Ingester__

![Write_Path_summarize](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/Write_Path_summarize.jpg)  

### Read path
- __Query Frontend__
  - Grafana等からのクエリーを最初に受け付ける
  - 広い範囲のデカいクエリーを小さく分割して複数のQuerierにパラレルに実行させてQuerierから帰ってきた結果をaggregationする
    - Query frontendが内部でqueueを持っていてそこに分割したクエリーを入れて、Querierがそこからqueueを取り出してクエリーを実行して結果をQuery frontendに返す
    - どの単位でクエリーを分割するかは`split_queries_by_interval`(defaultは30m)で設定できる  
    → 例えばデフォルト(30m)で2h範囲のクエリーを実行したら、4つのクエリーに分割してパラレルにQuerierに実行させる
  - クエリー結果をResult cacheにキャッシュする
  - クエリーrequestが失敗したら`max_retries`に設定された回数(defaultは5回)リトライする
  - 参考URL
    - https://grafana.com/docs/loki/latest/fundamentals/architecture/components/#query-frontend
    - https://grafana.com/docs/loki/latest/configuration/query-frontend/
    - https://github.com/taisho6339/loki-book/tree/main/query-process
- __Querier__
  - Query Frontendから連携されたクエリーをIngesterとBackend(S3)に投げて処理する
  - [Querier](https://grafana.com/docs/loki/latest/fundamentals/architecture/components/#querier)はデータのdeduplicationを行う
    > Queriers query all ingesters for in-memory data before falling back to running the same query against the backend store. Because of the replication factor, it is possible that the querier may receive duplicate data. To resolve this, the querier internally deduplicates data that has the same nanosecond timestamp, label set, and log message.
  - QuerierがStateful？
    - 検索時に使うindexを保持するためStateful
    - [Querier](https://grafana.com/docs/loki/latest/operations/storage/boltdb-shipper/#queriers)はObject storageからBoltDBファイルを`cache_location`で設定したディレクトリに非同期でダウンロード(lazily loads)する。read requestを受けた時にcache(index memcache)やダウンロードしたBoltDBファイルに該当indexが存在しない場合はObject storageから同様に`cache_location`で設定したディレクトリにダウンロードし、cache(index memcache)にindexを保存する。
    - ただ、Read Performanceに影響するもので[Querier](https://grafana.com/docs/loki/latest/operations/storage/boltdb-shipper/#queriers)が落ちても再度indexをObject storageからダウンロードして終わりの話の気がする。。  
    → つまりQueriorがStatefulであることはそこまで気にしなくて良いのでは？  
      >  When a querier receives a read request, the query range from the request is resolved to period numbers and all the files for those period numbers are downloaded to cache_location, if not already.   

      しかもQueriorにも専用のEBSをProvisioningするので再起動を気にせず、QueriorにMemoryのLimitsを設定して良さそう
      > Within Kubernetes, if you are not using an Index Gateway, we recommend running Queriers as a StatefulSet with persistent storage for downloading and querying index files. This will obtain better read performance, and it will avoid using node disk.
    - [Index Gateway](https://grafana.com/docs/loki/latest/operations/storage/boltdb-shipper/#queriers)という別コンポーネントをデプロイすればQueriorをStatelessにすることができる。
  - 参考URL
    - https://grafana.com/docs/loki/latest/fundamentals/architecture/components/#querier
    - https://grafana.com/docs/loki/latest/operations/storage/boltdb-shipper/

![Read_Path](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/Read_Path.jpg)  

### その他
- [__Compactor__](https://grafana.com/docs/loki/latest/operations/storage/boltdb-shipper/#compactor)
  - 複数のindexを重複排除して1つのファイルとしてまとめる
  - query latencyの改善につながる
  - Compactorは1つだけ動かさないといけない
    - 複数動かすとdata lossにつながる問題を起こす恐れがある
      > Note: There should be only 1 compactor instance running at a time that otherwise could create problems and may lead to data loss.
  - compact前
    ![before_compact](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/before_compact_2.jpg)  
  - compact後
    ![after_compact](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/after_compact_2.jpg)  

## BoltDB Shipper
- __背景__
  - Lokiは`index`と`chunk`2種類のデータを保存する必要がある
    > Grafana Loki needs to store two different types of data: chunks and indexes.
    > Loki receives logs in separate streams, where each stream is uniquely identified by its tenant ID and its set of labels. As log entries from a stream arrive, they are compressed as “chunks” and saved in the chunks store. 
    - `index`  
      → labelとtenant IDの組合せから生成されるchunkを検索するためのindex
    - `chunk`  
      → logデータが圧縮されたもの
    ![Write_Path](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/Write_Path.jpg)  
  - v1.5前まではindex(ex. DynamoDB)とchunk(ex. S3)を別々のところに保存していた
  - v1.5からindexもchunkと同じObject Storageに保存できるようにするためにBoltDB Shipperが登場した
- __仕組み__
  - BoltDB[^1]という組み込み型KVSにindexを保存し、それをObject Storageに送信したり、逆にObject Storageから別のIngesterから生成されたindexを受信して同期する
    [^1]: https://grafana.com/docs/loki/latest/storage/#boltdb、https://github.com/boltdb/bolt
  - BoltDBとBoltDB Shipperが使われるのはIngesterとQuerior
    - [Ingesters](https://grafana.com/docs/loki/latest/operations/storage/boltdb-shipper/#ingesters)
      > Ingesters keep writing the index to BoltDB files in active_index_directory and BoltDB Shipper keeps looking for new and updated files in that directory every 1 Minutes to upload them to the shared object store. When running Loki in clustered mode there could be multiple ingesters serving write requests hence each of them generating BoltDB files locally.
      > 
      > Note: To avoid any loss of index when Ingester crashes it is recommended to run Ingesters as statefulset(when using k8s) with a persistent storage for storing index files.
      > 
      > Another important detail to note is when chunks are flushed they are available for reads in object store instantly while index is not since we only upload them every 15 Minutes with BoltDB shipper. Ingesters expose a new RPC for letting Queriers query the Ingester’s local index for chunks which were recently flushed but its index might not be available yet with Queriers. For all the queries which require chunks to be read from the store, Queriers also query Ingesters over RPC for IDs of chunks which were recently flushed which is to avoid missing any logs from queries.
    - [Querior](https://grafana.com/docs/loki/latest/operations/storage/boltdb-shipper/#queriers)については上の ***QuerierがStateful？*** を参照
- 参考URL
  - https://grafana.com/docs/loki/latest/operations/storage/boltdb-shipper/
  - https://grafana.com/docs/loki/latest/fundamentals/architecture/

## LogがDropされるのを防ぐための仕組み
- __Replication factor__
- __WAL (Write Ahead Log)__
  - ingesterが書き込みを受け付ける前にまず先にログをDiskに全部書き込んでからメモリに書き込む。  
  そして、ingesterが何らかの理由で落ちたら、起動時にメモリにあったすべてのログを読み込んで修復する。
    > This is a new feature, available starting in the 2.2 release, which helps ensure Loki doesn’t drop logs by writing all incoming data to disk before acknowledging the write. If an ingester dies for some reason, it will replay this log upon startup, safely ensuring all the data it previously had in memory has been recovered.
  - WALから読み込む(replay)時にWALサイズがingesterが利用可能な(割り当てられている)メモリサイズより大きい場合、幸いにメモリが制限されている状態でもbackpressureの形でreplayが実行されるけど、`replay_memory_ceiling`の設定でreplayデータ量が設定値に達したらreplayを一旦止めてflushさせてから再開させることができる
    > replay_memory_ceiling It’s possible that after an outage scenario, a WAL is larger than the available memory of an ingester. Fortunately, the WAL implements a form of backpressure, allowing large replays even when memory is constrained. This replay_memory_ceiling config is the threshold at which the WAL will pause replaying and signal the ingester to flush its data before continuing. Because this is a less efficient operation, we suggest setting this threshold to a high, but reasonable, bound, or about 75% of the ingester’s normal memory limits. 
- 参考URL
  - https://grafana.com/blog/2021/02/16/the-essential-config-settings-you-should-use-so-you-wont-drop-logs-in-loki/
  - https://grafana.com/docs/loki/latest/design-documents/2020-09-write-ahead-log/
  - https://grafana.com/docs/loki/latest/operations/storage/boltdb-shipper/#queriers

## chunkの圧縮
- 転送速度向上およびストレージコスト削減のため、ログはgzip[^3]で圧縮されてchunkとして保存される
  [^3]: defaultではgzipだけどingesterの設定`chunk_encoding`にてsnappy(圧縮率は低いけどその分検索が早い)などに変えることもできる
  > gzip is the default and has the best compression ratio, but we suggest snappy for its faster decompression rate, which results in better query speeds.
- 参考URL
  - https://grafana.com/docs/loki/latest/operations/storage/boltdb-shipper/#operational-details
  - https://grafana.com/docs/loki/latest/configuration/#ingester
  - https://grafana.com/blog/2021/02/16/the-essential-config-settings-you-should-use-so-you-wont-drop-logs-in-loki/

## Configuration
### ingester
- 参考URL
  - https://grafana.com/docs/loki/latest/configuration/#ingester
  - https://grafana.com/docs/loki/latest/best-practices/#use-chunk_target_size
- 以下の3つがingesterからBackend(S3等)にflushされるタイミングに影響する設定  
  → 個の3つの値を大きくするとメモリ使用量も上がるので要注意
  - `chunk_target_size`
    - chunkがここに設定したsizeに達したらingesterがBackend(S3)にchunkをflushする
  - `max_chunk_age`
    - ここに指定した時間が経過したchunkをflushする
  - `chunk_idle_period`
    - ここに指定した時間の間、chunkに更新がない場合flushする

## Observability
- Loki/promtailも自身に関するメトリクスを開示している
  - https://grafana.com/docs/loki/latest/operations/observability/
- 役に立つメトリクス[^2]
  [^2]: https://taisho6339.gitbook.io/grafana-loki-deep-dive/monitoring
  - __Distributor__
    - `loki_distributor_lines_received_total` (counter)  
      → Distributorが受け付けたログ数(per tanant)
    - `loki_distributor_ingester_append_failures_total` (counter)  
      → The total number of failed batch appends sent to ingesters.  
        > **Note**  
        > ingesterへのappendが失敗した場合再送されるのか、このメトリクスの影響を確認！  
        > replication_factorの中で一部失敗したけど過半数は成功したので問題なしとか？
  - __Ingester__
    - `loki_ingester_chunks_flushed_total` (counter)  
      → どの要因でflushされたか、以下の`reason`ごとにflushされた件数  
       ・`full` → `chunk_target_size`の条件を満たしてflushされたもの  
       ・`idle` → `chunk_idle_period`の条件を満たしてflushされたもの  
       ・`max_age` → `max_chunk_age`の条件を満たしてflushされたもの  
  - __promtail__
    - `promtail_sent_entries_total` (counter)  
      → promtailがingesterに送ったログ数
    - `promtail_dropped_entries_total` (counter)  
      → promtailが設定されているすべてのretry回数内にingesterへの送信に失敗した(dropされた)ログ数

## HelmによるMicroServices Modeのデプロイ
- githubリポジトリ
  - https://github.com/grafana/helm-charts/tree/main/charts/loki-distributed
- Volumesは`/var/loki`にマウントされるので、各設定上のdirectoryは`/var/loki`配下(e. g. `/var/loki/index`, `/var/loki/cache`)に設定すること

## migration between k8s clusters (for k8s cluster VerUp)
1. Cluster VerUp前  
![migration_1](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/loki_migration_1.jpg)
2. 新Ver Cluster作成  
![migration_2](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/loki_migration_2.jpg)
3. NLB Target Groupに新Ver Cluster上のLokiを登録
![migration_3](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/loki_migration_3.jpg)
4. NLB Target Groupから旧Ver Cluster上のLokiを削除
![migration_4](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/loki_migration_4.jpg)
5. 旧Ver Cluster上のLoki(ingester)に対してflushを実行
![migration_5](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/loki_migration_5.jpg)
6. 新クラスター上のLokiから旧Lokiにあったログがすべて見えることを確認[^4]
![migration_6](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/loki_migration_6.jpg)
[^4]: 数十分～1時間くらいかかる
7. 旧クラスター上のLokiを削除（EBSも明示的に削除）
![migration_7](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/loki_migration_7.jpg)