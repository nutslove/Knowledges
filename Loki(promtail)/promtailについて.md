## Rate Limit

## Position

## Retry
- Lokiへのログ送信に失敗した際、リトライを行う
- デフォルトでは下記の通り10回リトライを行い、すべて失敗したらログがDropされる
  > Default backoff schedule:
  > 0.5s, 1s, 2s, 4s, 8s, 16s, 32s, 64s, 128s, 256s(4.267m)
  > For a total time of 511.5s(8.5m) before logs are lost
- `clients.backoff_config`blockにて最大リトライ数(`max_retries`)等を設定(変更)できる
- 参考URL
  - https://grafana.com/docs/loki/latest/clients/promtail/configuration/#clients
  - https://grafana.com/docs/loki/latest/clients/promtail/troubleshooting/#loki-is-unavailable