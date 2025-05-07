- https://github.com/grafana/tempo/tree/main/operations/tempo-mixin/dashboards
- https://grafana.com/docs/tempo/latest/operations/monitor/

## メトリクス一覧
|Metrics名|メトリクス概要|Type|PromQL例|備考|
|---|---|---|---|---|
|`tempo_request_duration_seconds_count`|QPS|Histogram|`sum by(pod,status_code)(rate(tempo_request_duration_seconds_count[$__rate_interval]))`||
|`tempo_request_duration_seconds`|レイテンシー|Histogram|`sum by(pod)(rate(tempo_request_duration_seconds_sum[$__rate_interval]) / rate(tempo_request_duration_seconds_count[$__rate_interval]))`、`sum by(pod)(histogram_quantile(0.95, rate(tempo_request_duration_seconds_bucket[$__rate_interval])))`||