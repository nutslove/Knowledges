- ThanosのGithubリポジトリにGrafana dashboardとAlert設定のサンプルがあるので参考にすること
  - Dashboard(確認すべきメトリクス)
    - https://github.com/thanos-io/thanos/blob/main/examples/dashboards/
  - Alert
    - https://github.com/thanos-io/thanos/tree/main/examples/alerts

- Thanosが出しているメトリクス一覧の参考情報
  - https://github.com/thanos-io/thanos/issues/5758

- **Thanos-mixin**
  - **https://github.com/thanos-io/thanos/tree/main/mixin**

## 共通
| metric名 | metric概要 | metric type | 備考 |
|---|---|---|---|
|`http_requests_total`|リクエスト数| Counter |`code`ラベルでHTTP StatusCodeを確認できるので、5xx系を監視する。(routing receiver, querierで開示されるメトリクス)|
|`http_request_duration_seconds_bucket`|レイテンシー| Histogram | `code`ラベルでHTTP StatusCodeを確認できるので、5xx系を監視する。 (routing receiver, querierで開示されるメトリクス)|
|`thanos_objstore_bucket_operation_failures_total`|オブジェクトストレージとのやり取り(e.g. GET, PUTなど)で何かしらの理由で作業が失敗した回数|Counter| `operation`ラベルでS3に対してどういうオペレーションかが分かる (ingesting receiver, compactor, storeで開示されるメトリクス) |
|`thanos_objstore_bucket_operation_duration_seconds_bucket`|ThanosがS3などのObject Storageとの処理にかかった時間|Histogram|`operation`ラベルに処理の種類(e.g. get, uploadなど)が入る (ingesting receiver, compactor, storeで開示されるメトリクス)|

## Receiver
| metric名 | metric概要 | metric type | 備考 |
| --- | --- | --- | --- |
|`thanos_receive_replications_total`|`replication-factor`に設定した分、複数のReceiverにレプリケーションした回数| Counter | `result`ラベルに`error`か`success`が入り、`error`のメトリクスを監視する。PromQL例(アラートに設定): `thanos_receive_replication_factor > 1 and ((sum by (job) (rate(thanos_receive_replications_total{result="error", job=~".*thanos-receive.*"}[5m])) / sum by (job) (rate(thanos_receive_replications_total{job=~".*thanos-receive.*"}[5m]))) > (max by (job) (floor((thanos_receive_replication_factor{job=~".*thanos-receive.*"}+1) / 2)) / max by (job) (thanos_receive_hashring_nodes{job=~".*thanos-receive.*"}))) * 100` |
|`thanos_receive_forward_requests_total`|routing receiverが受信した書き込みリクエストをhashringの設定によって他のrouting receiverに転送(forward)した件数|Counter|`result`ラベルに`error`か`success`が入り、`error`のメトリクスを監視する。|
|`thanos_receive_head_series_limited_requests_total` |メモリ内のhead blockのactive series数の制限によって拒否された総書き込みリクエスト数|Counter|リミットはテナントごとの`series_limit`で変更可|

## Query Frontend
| metric名 | metric概要 | metric type |
| --- | --- | --- |
| `thanos_query_frontend_queries_total` | Total queries passing through query frontend | Counter |
| `thanos_frontend_split_queries_total` | Total number of underlying query requests after the split by interval is applied | Counter |
| `thanos_frontend_downsampled_extra_queries_total` | Total number of additional queries for downsampled data | Counter |

## Compactor
| metric名 | metric概要 | metric type | 備考 |
| --- | --- | --- | --- |
| `thanos_compact_halted` | will be set to 1 when halt happens | Gauge | |
| `thanos_compact_downsample_total` | Total number of downsampling attempts. | Counter | |
| `thanos_compact_downsample_failures_total` | Total number of failed downsampling attempts. | Counter | |
| `thanos_compact_downsample_duration_seconds` | Duration of downsample runs | Histogram | |
| `thanos_compact_iterations_total` | Total number of iterations that were executed successfully. | Counter | |
| `thanos_compact_group_compactions_failures_total` | compactions失敗数 | Counter | PromQL例: `(rate(thanos_compact_group_compactions_failures_total[5m]) / rate(thanos_compact_group_compactions_total[5m])) * 100`|
|`thanos_objstore_bucket_last_successful_upload_time`|ThanosがS3に最後に正常にデータをアップロードした時間 |Gauge|　PromQL例: `(time() - max by (job) (max_over_time(thanos_objstore_bucket_last_successful_upload_time{job=~".*thanos-compact.*"}[24h]))) / 60 / 60`|