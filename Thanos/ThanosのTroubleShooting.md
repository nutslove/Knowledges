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