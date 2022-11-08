// `go 関数()`の意味調べる。
// 1回目メトリクス取得後更新されない(値が変わらない)。理由調べる。

package main

import (
        "net/http"
        "log"
        "context"
        "time"

        "github.com/prometheus/client_golang/prometheus"
        // "github.com/prometheus/client_golang/prometheus/promauto"
        "github.com/prometheus/client_golang/prometheus/promhttp"
        "github.com/oracle/oci-go-sdk/v65/common"
        "github.com/oracle/oci-go-sdk/v65/example/helpers"
        "github.com/oracle/oci-go-sdk/v65/monitoring"
)

var (
        oci_instance_cpu_utilization = prometheus.NewGaugeVec(prometheus.GaugeOpts{
                Name: "oci_compute_instance_cpu_utilization",
                Help: "OCI Compute Instance CPU Utilization(percent)",
        },
                []string{
                        "resourceId",
                        "availabilityDomain",
                        "region",
                        "resourceDisplayName",
                        "faultDomain",
                },
        )
)

func init() {
        // Metrics have to be registered to be exposed:
        prometheus.MustRegister(oci_instance_cpu_utilization)
}

func GetMetrics() {
        client, err := monitoring.NewMonitoringClientWithConfigurationProvider(common.DefaultConfigProvider())
        helpers.FatalIfError(err)

	req := monitoring.SummarizeMetricsDataRequest{
		// CompartmentIdInSubtree: common.Bool(true),
		// OpcRequestId: common.String("2FAIPMGNBUMX9TB067XB<unique_ID>"),
		SummarizeMetricsDataDetails: monitoring.SummarizeMetricsDataDetails{
			// Resolution: common.String("EXAMPLE-resolution-Value"),
			// ResourceGroup: common.String("EXAMPLE-resourceGroup-Value"),
			// StartTime:     &common.SDKTime{Time: time.Now()},
			// EndTime:       &common.SDKTime{Time: time.Now()},
			Namespace:     common.String("oci_computeagent"),
			Query:         common.String("CpuUtilization[1m].avg()")},
		CompartmentId: common.String("ocid1.tenancy.oc1..aaaaaaaapscp6dkyqn52z2j5u2qrz3yvhl7lw225maur75ptlollu4npaeza")}

	// Send the request using the service client
	resp, err := client.SummarizeMetricsData(context.Background(), req)
	helpers.FatalIfError(err)

        for {
                for i, _ := range resp.Items {
                        resourcedisplayname := resp.Items[i].Dimensions["resourceDisplayName"]
                        region := resp.Items[i].Dimensions["region"]
                        availabilitydomain := resp.Items[i].Dimensions["availabilityDomain"]
                        faultdomain := resp.Items[i].Dimensions["faultDomain"]
                        resourceid := resp.Items[i].Dimensions["resourceId"]
                        metric := *resp.Items[i].AggregatedDatapoints[len(resp.Items[i].AggregatedDatapoints)-1].Value
        
                        oci_instance_cpu_utilization.With(prometheus.Labels{
                                "resourceId": resourceid,
                                "availabilityDomain": availabilitydomain,
                                "region": region,
                                "resourceDisplayName": resourcedisplayname,
                                "faultDomain": faultdomain,
                        }).Set(metric)
                }
                time.Sleep(10 * time.Second)        
        }
}

func main() {
        go GetMetrics()

        http.Handle("/metrics", promhttp.Handler())
        log.Fatal(http.ListenAndServe(":8080", nil))
}