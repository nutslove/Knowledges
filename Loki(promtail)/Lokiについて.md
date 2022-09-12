## 参考URL
- https://taisho6339.hatenablog.com/entry/2021/05/26/104449
- https://taisho6339.gitbook.io/grafana-loki-deep-dive/
- https://speakerdeck.com/line_developers/grafana-loki-deep-dive

## Architecture
![Loki_Architecture](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/Loki_Architecture.jpg)  
出典: https://grafana.com/blog/2021/08/11/a-guide-to-deploying-grafana-loki-and-grafana-tempo-without-kubernetes-on-aws-fargate/

## Components
#### Write path
- __Distributor__
- __Ingester__

![Write_Path_summarize](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/Write_Path_summarize.jpg)  

#### Read path
- __Query Frontend__
  - Grafana等からのクエリーを最初に受け付ける
  - 広い範囲のデカいクエリーを小さく分割して複数のQuerierにパラレルに実行させてQuerierから帰ってきた結果をaggregation(+deduplication)する
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
  - QuerierがStateful？
    - indexを保持するため、Stateful
    - `index-gateway`という
    - https://grafana.com/docs/loki/latest/operations/storage/boltdb-shipper/
  - 参考URL
    - https://grafana.com/docs/loki/latest/fundamentals/architecture/components/#querier

![Read_Path](https://github.com/nutslove/Knowledges/blob/main/Loki(promtail)/image/Read_Path.jpg)  

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
- 参考URL
  - https://grafana.com/docs/loki/latest/operations/storage/boltdb-shipper/
  - https://grafana.com/docs/loki/latest/fundamentals/architecture/

## LogがDropされるのを防ぐための仕組み
- __Replication factor__
- __WAL (Write Ahead Log)__
  - ingesterが書き込みを受け付ける前にまず先にログをDiskに全部書き込んでからメモリに書き込む。  
  そして、ingesterが何らかの理由で落ちたら、起動時にメモリにあったすべてのログを読み込んで修復する。
    > This is a new feature, available starting in the 2.2 release, which helps ensure Loki doesn’t drop logs by writing all incoming data to disk before acknowledging the write. If an ingester dies for some reason, it will replay this log upon startup, safely ensuring all the data it previously had in memory has been recovered.
- WALから読み込む(replay)時にWALサイズがingesterが利用可能な(割り当てられている)メモリサイズより大きい場合、幸いにメモリが制限されている状態でもbackpressureの形でreplayが実行されるけど、`replay_memory_ceiling`の設定でreplayデータ量が設定値に達したらreplayを一旦止めてflushさせてから再開させることができる。
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
#### ingester
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