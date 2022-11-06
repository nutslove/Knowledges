- 事前設定
  - [OCI CLIインストール/設定](https://docs.oracle.com/ja-jp/iaas/Content/API/SDKDocs/cliinstall.htm#InstallingCLI__PromptsInstall)

- SDK/CLIを使うためのConfiguration
  - SDKを使うためにtenancy OCIDやuser credentials情報などを設定しておく必要がある
  - https://docs.oracle.com/en-us/iaas/Content/API/Concepts/sdkconfig.htm#SDK_and_CLI_Configuration_File
  - [Golangの設定](https://github.com/oracle/oci-go-sdk/blob/master/README.md#configuring)
- Monitoring API
  - https://docs.oracle.com/en-us/iaas/api/#/en/monitoring
  - Monitoring APIのEndpoint
    - https://telemetry-ingestion.ap-tokyo-1.oraclecloud.com
    - https://telemetry-ingestion.ap-osaka-1.oraclecloud.com
  - __API to find metric names and dimensions__
    - [ListMetrics](https://docs.oracle.com/en-us/iaas/api/#/en/monitoring/20180401/Metric/ListMetrics)
  - __API to query metrics by name__
    - [SummarizeMetricsData](https://docs.oracle.com/en-us/iaas/api/#/en/monitoring/20180401/MetricData/SummarizeMetricsData)

- Building Metric Queries
  - https://docs.oracle.com/en-us/iaas/Content/Monitoring/Tasks/buildingqueries.htm#CreateQuery

- __MQL(Monitoring Query Language)__
  - OCI Monitoring Metrics用のクエリー言語
  - https://docs.oracle.com/ja-jp/iaas/Content/Monitoring/Reference/mql.htm