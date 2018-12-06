package main

import (
	"fmt"

	"github.com/rodaine/table"
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
	fmt.Println(fmt.Sprintf("List: %s > %s > Clusters", ps.Profile, ps.Region))

	clusters, err := GetClusterList(ps)
	if err != nil {
		HandleAwsError(err)
		return
	}

	// Output the results in a nice table
	tbl := table.New("ID", "Name", "Running tasks", "Pending tasks", "Instances")
	tbl.WithHeaderFormatter(tblHeaderFmt).WithFirstColumnFormatter(tblColumnFmt)

	for id, c := range clusters.Clusters {
		tbl.AddRow(id, *c.ClusterName, *c.RunningTasksCount, *c.PendingTasksCount, *c.RegisteredContainerInstancesCount)
	}

	tbl.Print()
}

func listServices() {

}

func listTasks() {

}
