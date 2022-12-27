package main

import (
        "net/http"
        "log"
        "context"
        "time"
        "fmt"

        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promhttp"
        "github.com/oracle/oci-go-sdk/v65/common"
        "github.com/oracle/oci-go-sdk/v65/example/helpers"
        "github.com/oracle/oci-go-sdk/v65/monitoring"
)

// var namespaces = []string{
//         "oci_computeagent",
// }

var metricname = []string{
        "oci_compute_instance_cpu_utilization",
        "oci_compute_instance_memory_utilization",
        "oci_compute_instance_diskbytes_read",
        "oci_compute_instance_diskbytes_written",
        "oci_compute_instance_diskiops_read",
        "oci_compute_instance_diskiops_written",
        "oci_compute_instance_loadaverage",
        "oci_compute_instance_networkbytes_in",
        "oci_compute_instance_networkbytes_out",
}

var queries = []string{
        "CpuUtilization[1m].avg()",
        "MemoryUtilization[1m].avg()",
        "DiskBytesRead[1m].rate()",
        "DiskBytesWritten[1m].rate()",
        "DiskIopsRead[1m].rate()",
        "DiskIopsWritten[1m].rate()",
        "LoadAverage[1m].avg()",
        "NetworksBytesIn[1m].rate()",
        "NetworksBytesOut[1m].rate()",
}

var metrichelp = []string{
        "OCI Compute Instance CPU Utilization(percent)",
        "OCI Compute Instance Memory Utilization(percent)",
        "OCI Compute Instance Disk Read Throughput(bytes)",
        "OCI Compute Instance Disk Written Throughput(bytes)",
        "OCI Compute Instance Disk Read IOPS(count)",
        "OCI Compute Instance Disk Written IOPS(count)",
        "OCI Compute Instance LoadAverage(process)",
        "OCI Compute Instance Network Get Throughput(bytes)",
        "OCI Compute Instance Network Send Throughput(bytes)",
}

var labels = []string{
        "resourceId",
        "availabilityDomain",
        "region",
        "resourceDisplayName",
        "faultDomain",
}

var gaugevec = []*prometheus.GaugeVec{}

func GetGagueVec(metricname string, metrichelp string, labels []string) *prometheus.GaugeVec {
        return prometheus.NewGaugeVec(prometheus.GaugeOpts{
                Name: metricname,
                Help: metrichelp,
            },labels,
            )
}

func init() {
        // Metrics have to be registered to be exposed:
        for i, metricname := range metricname {
                mtr := GetGagueVec(metricname, metrichelp[i], labels)
                prometheus.MustRegister(mtr)
                gaugevec = append(gaugevec, mtr)
        } 
}

func GetMetric(namespace string, query string, gaugevec *prometheus.GaugeVec) {
        // Send the request using the service client 
        client, err := monitoring.NewMonitoringClientWithConfigurationProvider(common.DefaultConfigProvider())
        helpers.FatalIfError(err)

	req := monitoring.SummarizeMetricsDataRequest{
		// CompartmentIdInSubtree: common.Bool(true),
		// OpcRequestId: common.String("2FAIPMGNBUMX9TB067XB<unique_ID>"),
		SummarizeMetricsDataDetails: monitoring.SummarizeMetricsDataDetails{
			// Resolution: common.String("EXAMPLE-resolution-Value"),
			// ResourceGroup: common.String("EXAMPLE-resourceGroup-Value"),
			StartTime:     &common.SDKTime{Time: time.Now().Add(-time.Minute * 2)},
			EndTime:       &common.SDKTime{Time: time.Now()},
			Namespace:     common.String(namespace),
			Query:         common.String(query)},
		CompartmentId: common.String("ocid1.tenancy.oc1..aaaaaaaapscp6dkyqn52z2j5u2qrz3yvhl7lw225maur75ptlollu4npaeza")}

        fmt.Println("req: ",req)

        resp, err := client.SummarizeMetricsData(context.Background(), req)
        fmt.Println("resp: ",resp)

        helpers.FatalIfError(err)

        for i, _ := range resp.Items {
                resourcedisplayname := resp.Items[i].Dimensions["resourceDisplayName"]
                region := resp.Items[i].Dimensions["region"]
                availabilitydomain := resp.Items[i].Dimensions["availabilityDomain"]
                faultdomain := resp.Items[i].Dimensions["faultDomain"]
                resourceid := resp.Items[i].Dimensions["resourceId"]
                metric := *resp.Items[i].AggregatedDatapoints[len(resp.Items[i].AggregatedDatapoints)-1].Value

                gaugevec.With(prometheus.Labels{
                        "resourceId": resourceid,
                        "availabilityDomain": availabilitydomain,
                        "region": region,
                        "resourceDisplayName": resourcedisplayname,
                        "faultDomain": faultdomain,
                        }).Set(metric)
                }
}

func GetMetrics() {
        for {
                // for _, ns := range namespaces {
                //         for i, _ := range metricname {
                //                 GetMetric(ns,queries[i],gaugevec[i])
                //         }
                // }        
                for i, _ := range metricname {
                        GetMetric("oci_computeagent",queries[i],gaugevec[i])
                }
                time.Sleep(60 * time.Second)
        }
}

func main() {
        go GetMetrics()

        http.Handle("/metrics", promhttp.Handler())
        log.Fatal(http.ListenAndServe(":8080", nil))
}