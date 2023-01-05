## LokiのRateLimitによるpromtailからのError
- 事象
  - promtailのlogに以下のようなErrorが出る
    > level=warn ts=2022-11-30T06:59:03.853947325Z caller=client.go:369 component=client host=LOKI-NLB-******.elb.ap-northeast-1.amazonaws.com msg="error sending batch, will retry" status=429 error="server returned HTTP status 429 Too Many Requests (429): Ingestion rate limit exceeded for user fake (limit: 1310720 bytes/sec) while attempting to ingest '1154' lines totaling '1048548' bytes, reduce log volume or contact your Loki administrator to see if the limit can be increased"
- 原因
  - logに書いてある通りLokiの`limits_config.ingestion_rate_mb`と`limits_config.ingestion_burst_size_mb`で設定するRateLimitに引っかかっている
- 対処
  - Lokiの`limits_config.ingestion_rate_mb`と`limits_config.ingestion_burst_size_mb`の値を大きくする

## S3 ThrottlingによるIngesterからのError
- 事象
  - Ingesterのlogから以下のようなErrorが出る
    > "level=error ts=2023-01-04T16:35:24.514624173Z caller=flush.go:146 org_id=fake msg="failed to flush user" err="store put chunk: SlowDown: Please reduce your request rate.\n\tstatus code: 503, request id: K2HMP2637CPR01G5, host id: HNJwYGlLYO7AspzFdK8ik1UtmQrsgn43liboEzbyqIWcznqdOnZglxJ2SqOJvTAG+B04Mq/w+S4="\n"
- 原因
  - S3の

- 対処

- 参考URL
  - 