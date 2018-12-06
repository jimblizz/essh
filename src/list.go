package main

import (
	"fmt"

	"github.com/rodaine/table"
)

func list(ps PathStructure) {
	fmt.Println("List")

	if ps.Profile == "" {
		listProfiles()
		return
	}

	if ps.Region == "" {
		listRegions()
		return
	}

}

func listProfiles() {
	fmt.Println("List profiles")

	tbl := table.New("ID", "Name", "Access Key")
	tbl.WithHeaderFormatter(tblHeaderFmt).WithFirstColumnFormatter(tblColumnFmt)

	for id, p := range profiles {
		tbl.AddRow(id, p.Name, p.AccessKeyId)
	}

	tbl.Print()
}

func listRegions() {
	fmt.Println("List regions")
	// TODO: Pull the current list of regions from AWS
	// TODO: We can do this using EC2/ECS endpoint to get supported regions
	// TODO: We might want to cache that somehow?

	regionsMock := []string{
		"eu-west-1",
	}

	tbl := table.New("ID", "Name")
	tbl.WithHeaderFormatter(tblHeaderFmt).WithFirstColumnFormatter(tblColumnFmt)

	for id, r := range regionsMock {
		tbl.AddRow(id, r)
	}

	tbl.Print()

}

func listClusters() {
	fmt.Println("List clusters")
}

func listServices() {

}

func listTasks() {

}
