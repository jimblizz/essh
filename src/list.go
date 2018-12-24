package main

import (
    "bufio"
    "fmt"
    "github.com/rodaine/table"
    "os"
    "os/exec"
    "os/signal"
    "strconv"
    "strings"
    "syscall"
)

func list(ps PathStructure) {
    cmd := exec.Command("clear") // TODO: Support Windows
    cmd.Stdout = os.Stdout
    cmd.Run()

    // Listen for Ctrl-C events, and disconnect cleanly
    // Listen for sigterm
    // Watch for kill services
    var gracefulStop = make(chan os.Signal)
    signal.Notify(gracefulStop, syscall.SIGTERM)
    signal.Notify(gracefulStop, syscall.SIGINT)

    go func() {
        <-gracefulStop
        os.Exit(0)
    }()

	if ps.Profile == "" {
		listProfiles(ps)
		return
	}

	if !ps.HasValidProfileName() {
		fmt.Println(fmt.Sprintf("The requested profile %s does not exist in .aws/credentials. Valid profiles:", ps.Profile))
		listProfiles(ps)
		return
	}

	if ps.Region == "" {
		listRegions(ps)
		return
	}

	if !ps.HasValidRegionName() {
		fmt.Println(fmt.Sprintf("The requested region name %s is not available. Valid regions:", ps.Region))
		listRegions(ps)
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

func getUserSelection() int {
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Select ID > ")
    text, _ := reader.ReadString('\n')
    value := strings.TrimSuffix(text, "\n")

    if value == "exit" {
        os.Exit(0)
    }

    integer, err := strconv.Atoi(value)
    if err != nil {
        fmt.Println("That's not an integer!")
        os.Exit(0)
    }

    return integer
}

func listProfiles(ps PathStructure) {
    fmt.Println(fmt.Sprintf("List: Profiles"))

	tbl := table.New("ID", "Name", "Access Key")
	tbl.WithHeaderFormatter(tblHeaderFmt).WithFirstColumnFormatter(tblColumnFmt)

	for id, p := range profiles {
		tbl.AddRow(id, p.Name, p.AccessKeyId)
	}

	tbl.Print()

	i := getUserSelection()
	if len(profiles) > i {
        ps.Profile = profiles[i].Name
        list(ps)
    }
	return
}

func listRegions(ps PathStructure) {
    fmt.Println(fmt.Sprintf("List: %s > Regions", ps.Profile))

	// TODO: Pull the current list of regions from AWS
	// TODO: We can do this using EC2/ECS endpoint to get supported regions
	// TODO: We might want to cache that somehow?

	tbl := table.New("ID", "Name")
	tbl.WithHeaderFormatter(tblHeaderFmt).WithFirstColumnFormatter(tblColumnFmt)

	for id, r := range RegionsMock {
		tbl.AddRow(id, r)
	}

	tbl.Print()

    i := getUserSelection()
    if len(RegionsMock) > i {
        ps.Region = RegionsMock[i]
        list(ps)
    }
    return
}

func listClusters(ps PathStructure) {
	fmt.Println(fmt.Sprintf("List: %s > %s > Clusters", ps.Profile, ps.Region))

	clustersOutput, err := GetClusterList(ps)
	if err != nil {
		HandleAwsError(err)
		return
	}

	clusters := clustersOutput.Clusters

	// Output the results in a nice table
	tbl := table.New("ID", "Name", "Running tasks", "Pending tasks", "Instances")
	tbl.WithHeaderFormatter(tblHeaderFmt).WithFirstColumnFormatter(tblColumnFmt)

	for id, c := range clusters {
		tbl.AddRow(id, *c.ClusterName, *c.RunningTasksCount, *c.PendingTasksCount, *c.RegisteredContainerInstancesCount)
	}

	tbl.Print()

    i := getUserSelection()
    if len(clusters) > i {
        for ind, c := range clusters {
            if ind == i {
                ps.Cluster = *c.ClusterName
                list(ps)
                return
            }
        }
    }
    return
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

    i := getUserSelection()
    if len(services) > i {
        ps.Service = services[i].ServiceName
        list(ps)
    }
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

    i := getUserSelection()
    if len(containers) > i {
        ps.Container = containers[i].UUID
        list(ps)
    }
	return
}