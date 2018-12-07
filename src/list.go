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

	if !ps.HasValidProfileName() {
		fmt.Println(fmt.Sprintf("The requested profile %s does not exist in .aws/credentials. Valid profiles:", ps.Profile))
		listProfiles()
		return
	}

	if ps.Region == "" {
		listRegions()
		return
	}

	if !ps.HasValidRegionName() {
		fmt.Println(fmt.Sprintf("The requested region name %s is not available. Valid regions:", ps.Region))
		listRegions()
		return
	}

	if ps.Cluster == "" {
		listClusters(ps)
		return
	}

	if !ps.HasValidClusterName() {
		fmt.Println(fmt.Sprintf("The requested cluster name %s was not found. Valid regions:", ps.Region))
		listClusters(ps)
		return
	}

	listServices(ps)

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

	// TODO: Sort clusters A-Z

	// Output the results in a nice table
	tbl := table.New("ID", "Name", "Running tasks", "Pending tasks", "Instances")
	tbl.WithHeaderFormatter(tblHeaderFmt).WithFirstColumnFormatter(tblColumnFmt)

	for id, c := range clusters.Clusters {
		tbl.AddRow(id, *c.ClusterName, *c.RunningTasksCount, *c.PendingTasksCount, *c.RegisteredContainerInstancesCount)
	}

	tbl.Print()
}

func listServices(ps PathStructure) {
	fmt.Println(fmt.Sprintf("List: %s > %s > %s > Services", ps.Profile, ps.Region, ps.Cluster))

	services, err := GetServiceList(ps)
	if err != nil {
		HandleAwsError(err)
		return
	}

	tbl := table.New("ID", "Name", "Task definition", "Running", "Pending", "Desired", "Role", "Status")
	tbl.WithHeaderFormatter(tblHeaderFmt).WithFirstColumnFormatter(tblColumnFmt)

	for i, s := range services {
		tbl.AddRow(i, s.ServiceName, s.TaskDefinition, s.RunningCount, s.PendingCount, s.DesiredCount, s.Role, s.Status)
	}

	tbl.Print()

	return
}

func listTasks() {

}
