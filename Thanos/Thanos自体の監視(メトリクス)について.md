- ThanosのGithubリポジトリにGrafana dashboardとAlert設定のサンプルがあるので参考にすること
  - Dashboard(確認すべきメトリクス)
    - https://github.com/thanos-io/thanos/blob/main/examples/dashboards/
  - Alert
    - https://github.com/thanos-io/thanos/tree/main/examples/alerts

- Thanosが出しているメトリクス一覧の参考情報
  - https://github.com/thanos-io/thanos/issues/5758

## Query Frontend
| metric名 | metric概要 | metric type |
| --- | --- | --- |
| thanos_query_frontend_queries_total | Total queries passing through query frontend | Counter |
| thanos_frontend_split_queries_total | Total number of underlying query requests after the split by interval is applied | Counter |
| thanos_frontend_downsampled_extra_queries_total | Total number of additional queries for downsampled data | Counter |

## Compactor
| metric名 | metric概要 | metric type |
| --- | --- | --- |
| thanos_compact_halted | will be set to 1 when halt happens | Gauge |
| thanos_compact_downsample_total | Total number of downsampling attempts. | Counter |
| thanos_compact_downsample_failures_total | Total number of failed downsampling attempts. | Counter |
| thanos_compact_downsample_duration_seconds | Duration of downsample runs | Histogram |
| thanos_compact_iterations_total | Total number of iterations that were executed successfully. | Counter |