## 参考URL
- https://taisho6339.hatenablog.com/entry/2021/05/26/104449
- https://taisho6339.gitbook.io/grafana-loki-deep-dive/
- https://speakerdeck.com/line_developers/grafana-loki-deep-dive

## Architecture

## Observability
- Loki/promtailも自身に関するメトリクスを開示している
  - https://grafana.com/docs/loki/latest/operations/observability/
- 監視した方が良いメトリクス[^1]
  [^1]: 参考URL: https://taisho6339.gitbook.io/grafana-loki-deep-dive/monitoring
  - `loki_distributor_lines_received_total`
  - 

## Configuration
#### ingester
- https://grafana.com/docs/loki/latest/configuration/#ingester
- 以下の3つがingesterからBackend(S3等)にflushされるタイミングに関連する設定
  - `chunk_target_size`
    - chunkがここに設定したsizeに達したらingesterがBackend(S3)にchunkをflushする
  - `max_chunk_age`
    - ここに指定した時間が経過したchunkをflushする
  - `chunk_idle_period`
    - ここに指定した時間の間、chunkに更新がない場合flushする