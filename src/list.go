package main

import (
	"fmt"

	"github.com/rodaine/table"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func list(ps PathStructure) {

	if ps.Profile == "" {
		listProfiles()
		return
	}

	if !IsValidProfileName(ps.Profile) {
		fmt.Println(fmt.Sprintf("The requested profile %s does not exist in .aws/credentials. Valid profiles:", ps.Profile))
		listProfiles()
		return
	}

	if ps.Region == "" {
		listRegions()
		return
	}

	if !IsValidRegionName(ps.Region) {
		fmt.Println(fmt.Sprintf("The requested region name %s is not available. Valid regions:", ps.Region))
		listRegions()
		return
	}

	if ps.Cluster == "" {
		listClusters(ps)
		return
	}

}

func listProfiles() {
	tbl := table.New("ID", "Name", "Access Key")
	tbl.WithHeaderFormatter(tblHeaderFmt).WithFirstColumnFormatter(tblColumnFmt)

	for id, p := range profiles {
		tbl.AddRow(id, p.Name, p.AccessKeyId)
	}

	tbl.Print()
}

func listRegions() {
	// TODO: Pull the current list of regions from AWS
	// TODO: We can do this using EC2/ECS endpoint to get supported regions
	// TODO: We might want to cache that somehow?

	tbl := table.New("ID", "Name")
	tbl.WithHeaderFormatter(tblHeaderFmt).WithFirstColumnFormatter(tblColumnFmt)

	for id, r := range RegionsMock {
		tbl.AddRow(id, r)
	}

	tbl.Print()
}

func listClusters(ps PathStructure) {
	fmt.Println("List clusters")

	// New ECS client
	svc, err := NewEcsClient(ps)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Get a list of clusters
	listInput := &ecs.ListClustersInput{}
	clusters, err := svc.ListClusters(listInput)
	if err != nil {
		HandleAwsError(err)
		return
	}

	// Get more data on these clusters
	describeInput := &ecs.DescribeClustersInput{
		Clusters: clusters.ClusterArns,
	}
	results, err := svc.DescribeClusters(describeInput)
	if err != nil {
		HandleAwsError(err)
		return
	}

	// Output the results in a nice table
	tbl := table.New("ID", "Name", "Running tasks", "Pending tasks", "Instances")
	tbl.WithHeaderFormatter(tblHeaderFmt).WithFirstColumnFormatter(tblColumnFmt)

	for id, c := range results.Clusters {
		tbl.AddRow(id, *c.ClusterName, *c.RunningTasksCount, *c.PendingTasksCount, *c.RegisteredContainerInstancesCount)
	}

	tbl.Print()
}

func listServices() {

}

func listTasks() {

}
