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

	if ps.Service == "" {
		listServices(ps)
		return
	}

    if !ps.HasValidServiceName() {
        fmt.Println(fmt.Sprintf("The requested service %s was not found. Valid containers:", ps.Service))
        listServices(ps)
        return
    }

	if ps.Container == "" {
        listContainers(ps)
        return
    }

	// TODO: Check is valid container UUID

	// Redirect to SSH action!
    shellOpenSession(ps)

	return
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

func listContainers(ps PathStructure) {
	fmt.Println(fmt.Sprintf("List: %s > %s > %s > %s > Running", ps.Profile, ps.Region, ps.Cluster, ps.Service))

	instanceMap, err := GetClusterInstanceMap(ps)
	if err != nil {
		HandleAwsError(err)
		return
	}

	containers, err := GetContainerList(ps, instanceMap)
	if err != nil {
		HandleAwsError(err)
		return
	}

	tbl := table.New("ID", "UUID", "Revision", "Status", "Host instance", "Shortcut")
	tbl.WithHeaderFormatter(tblHeaderFmt).WithFirstColumnFormatter(tblColumnFmt)

	for i, c := range containers {
		shortcut := fmt.Sprintf("ssh ec2-user@%s -i %s.pem", c.Instance.PublicIpAddress, c.Instance.KeyName)
		tbl.AddRow(i, c.UUID, c.TaskRevision, c.Status, c.Instance.Ec2InstanceId, shortcut)
	}

	tbl.Print()

	return
}