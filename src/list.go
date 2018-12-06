package main

import (
	"fmt"

	"github.com/rodaine/table"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/aws/awserr"
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

	svc, err := NewEcsClient(ps)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	input := &ecs.ListClustersInput{}

	result, err := svc.ListClusters(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ecs.ErrCodeServerException:
				fmt.Println(ecs.ErrCodeServerException, aerr.Error())
			case ecs.ErrCodeClientException:
				fmt.Println(ecs.ErrCodeClientException, aerr.Error())
			case ecs.ErrCodeInvalidParameterException:
				fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)
}

func listServices() {

}

func listTasks() {

}
