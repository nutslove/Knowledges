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