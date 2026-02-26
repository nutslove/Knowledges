- 参考ページ
  - https://thanos.io/tip/operating/troubleshooting.md/

# Receiver `Error on ingesting samples that are too old or are too far into the future`エラー

### 事象
- ingesting-receiverの方で以下のようなエラーが出て、メトリクスがdropされる  
  ```shell
  ts=2025-06-17T09:17:14.227933068Z caller=writer_errors.go:142 level=warn component=receive component=receive-writer tenant=plat msg="Error on ingesting samples that are too old or are too far into the future" numDropped=499
  ```
### 原因
- いくつか原因があり得るっぽい

#### 1. Remote WriteするPrometheusと受け取るThanos側でTimeがsyncされてない場合
- PrometheusのUIでTSDB StatusでMin/Max Timeがすごい過去/未来の時刻になってないか確認  
  ![](./image/prometheus_tsdb_time.png)
- PrometheusとThanosのTimeを同期させる必要がある

#### 2. PrometheusもしくはThanosが一定期間の障害から復旧された場合
- 復活したタイミングで、溜まっていた古いメトリクスを連携されて、Thanosが受け付けれるデータの範囲を超えてエラーになる
- 色々調べても明確な解決策はないように見える
  - Prometheus側で古いメトリクスを削除する？
- 一応ThanosのReceiver側で`--tsdb.out-of-order.time-window`、`--tsdb.out-of-order.time-window`フラグ（defaultでは無効になっている）で未来/過去のデータを受け付けるように設定することもできる
  - https://thanos.io/tip/components/receive.md/  
    > - `--tsdb.too-far-in-future.time-window=0s`  
    >   - Configures the allowed time window for ingesting samples too far in the future.  
    >   Disabled (0s) by default. Please note enable this flag will reject samples in the future of receive local NTP time + configured duration due to clock skew in remote write clients.
    > - `--tsdb.out-of-order.time-window=0s`
    >   - [EXPERIMENTAL] Configures the allowed time window for ingestion of out-of-order samples.  
    >     Disabled (0s) by default.  
    >     **Please note if you enable this option and you use compactor, make sure you have the `--compact.enable-vertical-compaction` flag enabled, otherwise you might risk compactor halt.**
  - https://groups.google.com/g/prometheus-users/c/vtmeo06pxiE?pli=1

---

# Ingesting ReceiverのPV(EBS)の空き容量がなくなった場合
- Ingesting Receiverから以下のようなログが出る。ただ、必要最低限のReceiverが正常に動いていればメトリクスの受信はできる。  
  ```shell
  ts=2025-06-12T05:43:12.633036106Z caller=receive.go:665 level=error component=receive err="compact head: persist head block: mkdir /tmp/thanos/receive/unknown/01JXHAKTZR7NKFSPMQCGD2AT13.tmp-for-creation: no space left on device" msg="failed to flush storage"
  ts=2025-06-12T05:43:12.633062461Z caller=multitsdb.go:432 level=info component=receive component=multi-tsdb msg="closing TSDB" tenant=unknown
  ts=2025-06-12T05:43:12.635902998Z caller=receive.go:673 level=info component=receive msg="storage is closed"
  ts=2025-06-12T05:43:12.635973381Z caller=http.go:92 level=info component=receive service=http/server component=receive msg="internal server is shutting down" err="opening storage: open /tmp/thanos/receive/s000128/wal/00000286: no space left on device"
  ts=2025-06-12T05:43:12.636208934Z caller=shipper.go:337 level=warn component=receive component=multi-tsdb tenant=unknown msg="updating meta file failed" err="write /tmp/thanos/receive/unknown/thanos.shipper.json.tmp: no space left on device"
  ts=2025-06-12T05:43:12.636291969Z caller=shipper.go:337 level=warn component=receive component=multi-tsdb tenant=unknown msg="updating meta file failed" err="write /tmp/thanos/receive/unknown/thanos.shipper.json.tmp: no space left on device"
  ```
> [!CAUTION]  
> 上記のエラーは空き容量がなくなったときに数回出て、その後は何もログが出ないのでご注意

- Thanosが上記の状態になってメトリクスを受け付けれない状態でも、Thanos側では何もログが出ず、Prometheus側では以下のような503エラーログしか確認できない  
  ```shell
  time=2025-06-17T09:15:18.698Z level=WARN source=queue_manager.go:2031 msg="Failed to send batch, retrying" component=remote remote_name=6c1ca9 url=http://thanos-routing-receiver.monitoring.svc:19291/api/v1/receive err="server returned HTTP status 503 Service Unavailable: 3 errors: forwarding request to endpoint {thanos-ingesting-receiver-0.thanos-ingesting-receiver.monitoring.svc.cluster.local:10901 thanos-ingesting-receiver-0.thanos-ingesting-receiver.monitoring.svc.cluster.local:19391 }: rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 10.1.24.111:10901: connect: connection refused\"; forwarding request to endpoint {thanos-ingesting-receiver-1.thanos-ingesting-receiver.monitoring.svc.cluster.local:10901 thanos-ingesting-receiver-1.thanos-ingesting-receiver.monitoring.svc.cluster.local:19391 }: rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 10.1.27.68:10901: connect: connection refused\"; forwarding request to endpoint {thanos-ingesting-receiver-3.thanos-ingesting-receiver.monitoring.svc.cluster.local:10901 thanos-ingesting-receiver-3.thanos-ingesting-receiver.monitoring.svc.cluster.local:19391 }: rpc error: code = Unavailable desc = connection error: desc = \"transpo"
  ```

- Pod,PVを削除してPVの容量を増やした後、Podを再デプロイする

> [!IMPORTANT]  
> **以下のPromQLでPVの使用率を監視すること！**  
> `(kubelet_volume_stats_used_bytes / kubelet_volume_stats_capacity_bytes) * 100`

---

# Compactorの`compactor halt`エラー
### 事象
- Compactorのログに以下のようなエラー(`critical error detected; halting`)が出て、`thanos_compact_halted`メトリクスも"1"になる  
  ```shell
  ts=2026-02-18T11:42:30.015761639Z caller=compact.go:559 level=error msg="critical error detected; halting" err="compaction: group 0@13457662937362014338: compact blocks [/tmp/thanos/compact/compact/0@13457662937362014338/01KAN0XB9C5Z67DM96WR9JGMSW /tmp/thanos/compact/compact/0@13457662937362014338/01KAT5PRPF8WESAD5FE77CXFHE /tmp/thanos/compact/compact/0@13457662937362014338/01KAZAG8695GAN40H1T6ERVJC8 /tmp/thanos/compact/compact/0@13457662937362014338/01KB4FA0B3DHZ8ZPSMKY9329QF /tmp/thanos/compact/compact/0@13457662937362014338/01KB9M3EB1KGRG7XWJFWNYKWYN /tmp/thanos/compact/compact/0@13457662937362014338/01KBERX2HNHAQMEE4DFGTNNTTW /tmp/thanos/compact/compact/0@13457662937362014338/01KBKXWRJWNG4W6RET9AAG0B80]: 2 errors: preallocate: no space left on device; sync /tmp/thanos/compact/compact/0@13457662937362014338/01KHR8WC0TY3V7CK0F8QE10FYE.tmp-for-creation/chunks/000008: file already closed"
  ```
- `no space left on device;`エラーが出てるけど、実際 `(kubelet_volume_stats_used_bytes / kubelet_volume_stats_capacity_bytes) * 100`で確認したPV使用率は52%程度で、空き容量は十分にあるように見える状態だった

### 原因
- compaction処理時に、圧縮対象のInputブロックだけではなく、Outputブロックも同じPVに作成されるため、InputブロックとOutputブロックの両方の容量が必要になり、outputブロックのための空き容量が十分じゃなかった

> [!IMPORTANT]  
> - compactorのPVの空き容量が十分にあるように見えても、compaction処理に必要な容量が足りてない（処理に失敗する）可能性があるので、空き容量ではなく、`thanos_compact_halted`メトリクスと`critical error detected; halting`エラーログで監視すること！

### 対策
- compactorのPVの容量を増やす

---

# CompactorのOverlapエラー
### 事象
- Compactorのログに以下のようなエラーが出てた  
  ```shell
  ts=2026-02-26T01:25:08.161800499Z caller=compact.go:559 level=error msg="critical error detected; halting" err="compaction: group 0@10476963415830505152: pre compaction overlap check: overlaps found while gathering blocks. [mint: 1772056261448, maxt: 1772056285308, range: 23s, blocks: 2]: <ulid: 01KJBCJ5JCVEEER7YVZ71VT9E1, mint: 1772049600067, maxt: 1772056285308, range: 1h51m25s>, <ulid: 01KJBQ0VKS66N9W7R3DW9H27DC, mint: 1772056261448, maxt: 1772056800000, range: 8m58s>"
  ```

### 原因
> [!IMPORTANT]  
> ### Overlap発生条件
> 1. ブロックは 外部ラベル（external labels） と ダウンサンプリング解像度（resolution） が完全一致するとき、同じグループに分類されます。
>    - グループキー = resolution@labelsのハッシュ値
> 2. 同グループ内で時間範囲が重複していること
>    - 同一グループ内のブロックが MinTime〜MaxTime の時間範囲で重なっている場合にオーバーラップと判定される。
>    - Prometheus TSDBの OverlappingBlocks は、MinTimeでソート済みのブロック列を走査し、あるブロックの MinTime が直前のブロックの MaxTime より前であれば「時間範囲が重複している」と判定

- `01KJBQ0VKS66N9W7R3DW9H27DC`ブロックのMinTime（1772056261448）が`01KJBCJ5JCVEEER7YVZ71VT9E1`ブロックのMaxTime（1772056285308）より前になっているため、同じグループ内で時間範囲が重複していると判定されてエラーになっていた
- 上記のoverlapが発生した原因は、EKS AutoModeにより、ワーカーノードが再配置され、Ingesting ReceiverのPodが再起動され、TSDBのHead blockがフラッシュされた直後に、新しいHeadに全seriesの最初の数サンプルだけが入った状態でブロック化されたパターン
- 流れ：
  1. receiver-3がHead blockにデータを蓄積中（2h window途中）
  2. Podが再起動 → graceful shutdownでHead blockがフラッシュされ Block 1 が生成（2h境界の手前で途切れる）
  3. Pod再起動後、remote writeクライアントがバッファしていたデータを再送
  4. 2h境界到達 → 残りのデータで Block 2 が生成
  5. 23秒のオーバーラップは remote write の再送バッファによるもの

### 対策
- thanosのPodに入って、`thanos tool`を使ってoverlapしているブロック（rangeが短い方）に削除マークをつける  
  - Format  
    ```shell
    thanos tools bucket mark \
      --objstore.config-file=<bucket-config> \
      --id=<対象ブロックULID> \
      --marker=deletion-mark.json \
      --details="理由を記載"
    ```

  - 例  
    ```shell
    thanos tools bucket mark --id "01KJBQ0VKS66N9W7R3DW9H27DC" --details "workernode restart" --marker "deletion-mark.json" --objstore.config-file "/etc/thanos/object-store.yaml"
    ```
- **ただ、compactorには`--delete-delay=48h`が設定されていて、削除マークをつけたブロックはすぐには削除されない。**  
  そのため、以下のいずれかの追加対策が必要！  
  1. `thanos tools bucket cleanup`コマンドで**削除マークがついている**ブロックをすぐに削除する（**`--delete-delay=0s`を指定**）  
     - Format  
       ```shell
       thanos tools bucket cleanup \
         --objstore.config-file=<bucket-config> \
         --delete-delay=0s
       ```
     - 例  
       ```shell
       thanos tools bucket cleanup --objstore.config-file="/etc/thanos/object-store.yaml" --delete-delay=0s
       ```
  2. compactorの`--delete-delay=0s`に変更して、削除マークがついているブロックをすぐに削除するようにする（**ただし、compactorの再起動が必要**）

> [!CAUTION]  
> - S3上からブロックを直接削除するとcompactorなどから予期しないエラーが出る可能性があるので注意が必要

> [!IMPORTANT]  
> `thanos tools bucket verify`コマンド（`--issues=overlapped_blocks`）でoverlapが発生しているブロックを確認できる  
> ```shell
> thanos tools bucket verify --objstore.config-file="/etc/thanos/object-store.yaml" --issues=overlapped_blocks
> ```
> そこで実際にoverlapしているブロックがある場合は以下のように表示される（`found overlapped blocks`の部分）  
> ```shell
> ts=2026-02-26T12:49:11.839762239Z caller=factory.go:54 level=info msg="loading bucket configuration"
> ts=2026-02-26T12:49:11.840347286Z caller=verify.go:138 level=info verifiers=overlapped_blocks msg="Starting verify task"
> ts=2026-02-26T12:49:11.840393398Z caller=overlapped_blocks.go:29 level=info verifiers=overlapped_blocks verifier=overlapped_blocks msg="started verifying issue"
> ts=2026-02-26T12:49:18.874098206Z caller=fetcher.go:627 level=info component=block.BaseFetcher msg="successfully synchronized block metadata" duration=7.033673769s duration_ms=7033 cached=4390 returned=4091 partial=0
> ts=2026-02-26T12:49:18.881519859Z caller=overlapped_blocks.go:42 level=warn verifiers=overlapped_blocks verifier=overlapped_blocks msg="found overlapped blocks" group=0@80629547281119333 overlap="[mint: 1772056542896, maxt: 1772056566081, range: 23s, blocks: 2]: <ulid: 01KJBCTR6JA781C9D49TKS38QQ, mint: 1772049600067, maxt: 1772056566081, range: 1h56m6s>, <ulid: 01KJBQ5XJAZ4J3RVVA6TFVEFRB, mint: 1772056542896, maxt: 1772056800000, range: 4m17s>"
> ts=2026-02-26T12:49:18.881582192Z caller=overlapped_blocks.go:42 level=warn verifiers=overlapped_blocks verifier=overlapped_blocks msg="found overlapped blocks" group=0@6733965995379593202 overlap="[mint: 1772056341239, maxt: 1772056362895, range: 21s, blocks: 2]: <ulid: 01KJBCMJ1P0BYEXK8DMW6WFWJM, mint: 1772049600067, maxt: 1772056362895, range: 1h52m42s>, <ulid: 01KJBQ5E05BK2BEAJ3HEMJPRJX, mint: 1772056341239, maxt: 1772056800000, range: 7m38s>"
> ts=2026-02-26T12:49:18.881602598Z caller=verify.go:157 level=info verifiers=overlapped_blocks msg="verify task completed"
> ts=2026-02-26T12:49:18.881802832Z caller=main.go:174 level=info msg=exiting
> ```

> [!NOTE]  
> ブロックを削除するとThanos Storeから以下のようなWarningログが出て、`thanos_objstore_bucket_operation_failures_total`メトリクス（`operation`ラベルは`attributes`で）が増加する可能性があるけど、Storeのメモリに残っている古いブロックを参照しているためで、`--sync-block-duration`パラメータに指定している時間後にS3からSyncするので問題ない
> ```shell
> ts=2026-02-26T11:22:35.20312404Z caller=bucket.go:861 level=warn msg="loading block failed" elapsed=80.029083ms id=01KAFVZD5ZXSCRBAWV3YEQW38A err="create index header reader: write index header: new index reader: get object attributes of 01KAFVZD5ZXSCRBAWV3YEQW38A/index: The specified key does not exist."
> ```

### `--delete-delay`が48hになっている理由
#### 1. Store Gatewayのクエリ保護
> deleting blocks immediately can cause query failures, if store gateway still has the block loaded, or compactor is ignoring the deletion because it's compacting the block at the same time.          

Store Gatewayはブロックのメタデータをキャッシュしており、即座に削除するとキャッシュとの不整合でクエリエラーが発生する。48hの猶予でStore Gatewayがsyncしてキャッシュを更新する時間を確保する。

#### 2. Compactorがギャップなく動作するため
> The delay of deleteDelay/2 is added to ensure we fetch blocks that are meant to be deleted but do not have a replacement yet. This is to make sure compactor will not accidentally perform compactions with gap instead.

Compaction後、ソースブロックに削除マークが付くが、新しいcompacted blockがまだアップロード完了していない場合がある。deleteDelay/2（=24h）の間はマーク付きブロックを残すことで、置き換えブロックが揃う前にソースが消えてギャップが生じるのを防ぐ。

#### 3. オブジェクトストレージの結果整合性への対応
> blocks are not deleted directly. Instead, blocks are marked for deletion by uploading deletion-mark.json file for the block that was chosen to be deleted. This adds a consistent lock-free way of dealing with Object Storage irrespective of the choice of object storage.

S3などのオブジェクトストレージは結果整合性（eventual consistency）のため、書き込み直後に全リーダーから見えるとは限らない。48hの猶予でこの伝播時間を吸収する。

#### 4. 中断されたPartial Uploadの誤削除防止
> If the Compactor process crashes during upload of a compacted block, the whole compaction starts from scratch and a new block ID is created. This means that partial upload will never be retried.
> To handle this case there is the --delete-delay=48h flag that starts deletion of directories inside object storage without meta.json only after a given time.

Compactorがアップロード中にクラッシュすると、meta.json のない不完全なブロックディレクトリが残る。48hの猶予があることで、正常にアップロード中のブロック（まだ meta.jsonが書き込まれていないだけ）を誤って消さないようにしている。

---

| # | 理由 | 保護対象 |
|:-:|:-----|:---------|
| 1 | Store Gatewayのキャッシュ更新猶予 | クエリの安定性 |
| 2 | 置き換えブロックのアップロード完了待ち | Compactionのギャップ防止 |
| 3 | オブジェクトストレージの結果整合性 | 全コンポーネント間の整合性 |
| 4 | Partial Uploadの誤削除防止 | アップロード中のブロックの保護 |

