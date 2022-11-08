// This is an automatically generated code sample.
// To make this code sample work in your Oracle Cloud tenancy,
// please replace the values for any parameters whose current values do not fit
// your use case (such as resource IDs, strings containing ‘EXAMPLE’ or ‘unique_id’, and
// boolean, number, and enum parameters with values not fitting your use case).

package main

import (
	"context"
	"fmt"
	// "time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
	"github.com/oracle/oci-go-sdk/v65/monitoring"
	// "github.com/oracle/oci-go-sdk/v65/identity"
)

func main() {
	// // Create a default authentication provider that uses the DEFAULT
	// // profile in the configuration file.
	// // Refer to <see href="https://docs.cloud.oracle.com/en-us/iaas/Content/API/Concepts/sdkconfig.htm#SDK_and_CLI_Configuration_File>the public documentation</see> on how to prepare a configuration file.
	// client, err := monitoring.NewMonitoringClientWithConfigurationProvider(common.DefaultConfigProvider())
	// helpers.FatalIfError(err)

	// // Create a request and dependent object(s).

	// req := monitoring.ListMetricsRequest{CompartmentId: common.String("ocid1.tenancy.oc1..aaaaaaaapscp6dkyqn52z2j5u2qrz3yvhl7lw225maur75ptlollu4npaeza"),
	// 	CompartmentIdInSubtree: common.Bool(true),
	// 	Limit:                  common.Int(1000),
	// 	ListMetricsDetails: monitoring.ListMetricsDetails{
	// 		GroupBy:       []string{"namespace"},
	// 		SortBy:        monitoring.ListMetricsDetailsSortByName,
	// 		SortOrder:     monitoring.ListMetricsDetailsSortOrderAsc,
	// 	},
	// 	}

	// // Send the request using the service client
	// resp, err := client.ListMetrics(context.Background(), req)
	// helpers.FatalIfError(err)

	// // Retrieve value from the response.
	// fmt.Println(resp)





	// Create a default authentication provider that uses the DEFAULT
	// profile in the configuration file.
	// Refer to <see href="https://docs.cloud.oracle.com/en-us/iaas/Content/API/Concepts/sdkconfig.htm#SDK_and_CLI_Configuration_File>the public documentation</see> on how to prepare a configuration file.
	client, err := monitoring.NewMonitoringClientWithConfigurationProvider(common.DefaultConfigProvider())
	helpers.FatalIfError(err)

	// Create a request and dependent object(s).

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

	// Retrieve value from the response.
	// fmt.Println(resp)
	// fmt.Printf("The type of SummarizeMetricsData's response: %T\n",resp)
	// fmt.Printf("The type of SummarizeMetricsData.Items's response: %T\n",resp.Items)
	// fmt.Println(resp.Items[0])
	// fmt.Printf("The type of SummarizeMetricsData.Items[0]'s response: %T\n",resp.Items[0])
	// fmt.Printf("The type of SummarizeMetricsData.Items[0].Dimensions's response: %T\n",resp.Items[0].Dimensions)
	// fmt.Println(resp.Items[0].Dimensions)

	for i, _ := range resp.Items {
		fmt.Println("Instance Name:",resp.Items[i].Dimensions["resourceDisplayName"])
		fmt.Println("Region:",resp.Items[i].Dimensions["region"])
		fmt.Println("AD:",resp.Items[i].Dimensions["availabilityDomain"])
		fmt.Println("Fault Domain:",resp.Items[i].Dimensions["faultDomain"])
		fmt.Println("Resource ID:",resp.Items[i].Dimensions["resourceId"])
		fmt.Println("Metric:",*resp.Items[i].AggregatedDatapoints[len(resp.Items[i].AggregatedDatapoints)-1].Value)
	}

	// for k, v := range resp.Items[0].Dimensions {
	// 	fmt.Println("Key:", k, "Value:", v)
	// }
}