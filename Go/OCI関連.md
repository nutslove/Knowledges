### SDK/CLIを使うためのConfiguration
- SDKを使うためにtenancy OCIDやuser credentials情報などを設定しておく必要がある
- 構成ファイル`~/.oci/config`にtenancyやregionなどを設定した上で`ClientWithConfigurationProvider(common.DefaultConfigProvider())`で読み込む
- 例
  ~~~go
  client, err := monitoring.NewMonitoringClientWithConfigurationProvider(common.DefaultConfigProvider())
  helpers.FatalIfError(err)

  req := monitoring.SummarizeMetricsDataRequest{
		  SummarizeMetricsDataDetails: monitoring.SummarizeMetricsDataDetails{
		    StartTime:     &common.SDKTime{Time: time.Now().Add(-time.Minute * 2)},
		    EndTime:       &common.SDKTime{Time: time.Now()},
		    Namespace:     common.String(namespace),
		    Query:         common.String(query)},
		  CompartmentId: common.String("ocid1.tenancy.oc1..**********")}

  resp, err := client.SummarizeMetricsData(context.Background(), req)
  ~~~
- 構成ファイルに設定が必要な項目
  - https://docs.oracle.com/ja-jp/iaas/Content/API/Concepts/sdkconfig.htm
- 参考URL
  - https://docs.oracle.com/en-us/iaas/Content/API/Concepts/sdkconfig.htm#SDK_and_CLI_Configuration_File
  - [Golangの設定](https://github.com/oracle/oci-go-sdk/blob/master/README.md#configuring)

### Monitoring API
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

  ##### **`SummarizeMetricsDataRequest`構造体(Struct)**
    - > SummarizeMetricsDataRequest wrapper for the SummarizeMetricsData operation
    - https://pkg.go.dev/github.com/oracle/oci-go-sdk/monitoring#SummarizeMetricsDataRequest

  ##### **`SummarizeMetricsData`function**
    - > SummarizeMetricsData Returns aggregated data that match the criteria specified in the request. Compartment OCID required. For information on metric queries, see Building Metric Queries (https://docs.cloud.oracle.com/iaas/Content/Monitoring/Tasks/buildingqueries.htm). For important limits information, see Limits on Monitoring (https://docs.cloud.oracle.com/iaas/Content/Monitoring/Concepts/monitoringoverview.htm#Limits). Transactions Per Second (TPS) per-tenancy limit for this operation: 10.
    - Format
      ~~~go
      func (client MonitoringClient) SummarizeMetricsData(ctx context.Context, request SummarizeMetricsDataRequest) (response SummarizeMetricsDataResponse, err error)
      ~~~
    - https://pkg.go.dev/github.com/oracle/oci-go-sdk/monitoring#MonitoringClient.SummarizeMetricsData

  ##### **`SummarizeMetricsDataDetails`構造体(Struct)**
    - > SummarizeMetricsDataDetails The request details for retrieving aggregated data. Use the query and optional properties to filter the returned results.
    - https://pkg.go.dev/github.com/oracle/oci-go-sdk/monitoring#SummarizeMetricsDataDetails

### __MQL(Monitoring Query Language)__
  - OCI Monitoring Metrics用のクエリー言語
  - https://docs.oracle.com/ja-jp/iaas/Content/Monitoring/Reference/mql.htm