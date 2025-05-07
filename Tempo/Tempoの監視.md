- https://github.com/grafana/tempo/tree/main/operations/tempo-mixin/dashboards
- https://grafana.com/docs/tempo/latest/operations/monitor/

## メトリクス一覧
|Metrics名|メトリクス概要|Type|PromQL例|備考|
|---|---|---|---|---|
|`tempo_request_duration_seconds_count`|QPS|Histogram|`sum by(pod,status_code)(rate(tempo_request_duration_seconds_count[$__rate_interval]))`||
|`tempo_request_duration_seconds`|レイテンシー|Histogram|`sum by(pod)(rate(tempo_request_duration_seconds_sum[$__rate_interval]) / rate(tempo_request_duration_seconds_count[$__rate_interval]))`、`sum by(pod)(histogram_quantile(0.95, rate(tempo_request_duration_seconds_bucket[$__rate_interval])))`||
|`tempo_receiver_refused_spans`|受信Span数|Counter|`sum(rate(tempo_receiver_accepted_spans[$__rate_interval]))`||
|`tempo_receiver_refused_spans`|rejectedされたSpan数|Counter|`rate(tempo_receiver_refused_spans[$__rate_interval])`||
|`tempo_distributor_bytes_received_total`|distributor受信bytes|Counter|`rate(tempo_distributor_bytes_received_total[$__rate_interval])`||
|`tempo_memcache_request_duration_seconds_count`|QPS(memcached)|Histogram|`sum by(pod,status_code)(rate(tempo_memcache_request_duration_seconds_count[$__rate_interval]))`||
|`tempo_memcache_request_duration_seconds`|レイテンシー(memcached)|Histogram|`sum by(pod)(rate(tempo_memcache_request_duration_seconds_sum[$__rate_interval]) / rate(tempo_memcache_request_duration_seconds_count[$__rate_interval]))`||