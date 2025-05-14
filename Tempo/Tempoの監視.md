- https://github.com/grafana/tempo/tree/main/operations/tempo-mixin/dashboards
- https://grafana.com/docs/tempo/latest/operations/monitor/

## メトリクス一覧
|Metrics名|メトリクス概要|Type|PromQL例|備考|
|---|---|---|---|---|
|`tempo_request_duration_seconds_count`|QPS|Histogram|`sum by(pod,status_code)(rate(tempo_request_duration_seconds_count[$__rate_interval]))`|`status_code`ラベルの5xx/4xx系を監視|
|`tempo_request_duration_seconds`|レイテンシー|Histogram|`sum by(pod)(rate(tempo_request_duration_seconds_sum[$__rate_interval]) / rate(tempo_request_duration_seconds_count[$__rate_interval]))`、`sum by(pod)(histogram_quantile(0.95, rate(tempo_request_duration_seconds_bucket[$__rate_interval])))`||
|`tempo_distributor_spans_received_total`|Receiverから受信Span数|Counter|`sum(rate(tempo_distributor_spans_received_total[$__rate_interval]))`||
|`tempo_receiver_accepted_spans`|Receiverから正常にパイプラインにプッシュされたSpan数。ReceiverでSpanが正常に処理されたことを示す。|||`transport`ラベル(e.g. `http`)でプロトコルを確認できる|
|`tempo_receiver_refused_spans`|rejectedされたSpan数|Counter|`rate(tempo_receiver_refused_spans[$__rate_interval])`|**`tempo_discarded_spans_total`は受信はされた後に　内部のパイプラインで何らかの理由で破棄されたSpanで、`tempo_receiver_refused_spans`はReceiverの段階で何らなの理由でパイプラインにプッシュできず拒否されたSpan**|
|`tempo_discarded_spans_total`|破棄されたSpan数|Counter|`rate(tempo_discarded_spans_total[$__rate_interval])`| `reason`（.e.g `rate_limited`, `trace_too_large`など）と`tenant`ラベルを持っていてどのような理由でどのテナントのspanが破棄されたか分かる。|
|`tempo_distributor_bytes_received_total`|distributor受信bytes|Counter|`rate(tempo_distributor_bytes_received_total[$__rate_interval])`||
|`tempo_memcache_request_duration_seconds_count`|QPS(memcached)|Histogram|`sum by(pod,status_code)(rate(tempo_memcache_request_duration_seconds_count[$__rate_interval]))`||
|`tempo_memcache_request_duration_seconds`|レイテンシー(memcached)|Histogram|`sum by(pod)(rate(tempo_memcache_request_duration_seconds_sum[$__rate_interval]) / rate(tempo_memcache_request_duration_seconds_count[$__rate_interval]))`||
|`tempo_limits_defaults`、`tempo_limits_overrides`| テナントごとの各limit設定 | Gauge |`max(max by (tenant_id, limit_name) (tempo_limits_overrides) or max by (tenant_id, limit_name) (tempo_limits_defaults)) by (tenant_id, limit_name)`|
|`tempodb_blocklist_length`| テナント/PodごとのTotal blocks |Gauge|`sum by(pod,tenant)(tempodb_blocklist_length)`|
|`tempodb_blocklist_tenant_index_builder`| A gauge that has the value 1 if this compactor is attempting to build the tenant index and 0 if it is not. At least one compactor must have this value set to 1 for the system to be working. |Gauge|`sum by(tenant)(tempodb_blocklist_tenant_index_builder > 0)`|
|`tempodb_blocklist_tenant_index_errors_total`| A holistic metrics that indcrements for any error building the tenant index. Any increase in this metric should be reviewed. (**エラーが発生した場合のみ生成されるメトリクス**) | Counter | `tempodb_blocklist_tenant_index_errors_total or vector(0)`|
|`tempodb_blocklist_poll_errors_total`| A holistic metric that increments for any error with polling the blocklist. Any increase in this metric should be reviewed. (**エラーが発生した場合のみ生成されるメトリクス**) |Counter| `tempodb_blocklist_poll_errors_total or vector(0)` |
|`tempo_ingester_live_traces`| 現在Object Storageにフラッシュされずに、ingester（メモリ）内部にあるトレース数 | Gauge | | |
|`tempodb_compaction_outstanding_blocks`| まだcompaction(圧縮)されてないBlock数 | Gauge | | | 
|`tempo_query_frontend_queries_total`| query frontendで処理されたクエリー数 | Counter | | `op`ラベルには、`traces`: TraceIDをもとにしたクエリ、`search`: Searchから検索したもの、`metadata`: metadataクエリ、`metrics`: メトリクスクエリ が入る |
|`tempo_distributor_push_duration_seconds`| distributorがデータをingesterにプッシュ(転送)するのにかかった時間 | Histogram | | | 
|`tempodb_backend_request_duration_seconds_count`| backend(ストレージ)に対するリクエスト数（e.g. GET、DELETEなど）| Histogram | | `operation`ラベルでリクエストの種類が分かる |
|`tempodb_backend_request_duration_seconds`| backend(ストレージ)に対するリクエストでかかった時間 | Histogram | | |