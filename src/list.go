package main

import (
	"fmt"

	"github.com/rodaine/table"
	"github.com/davecgh/go-spew/spew"
)

func list(ps PathStructure) {
	fmt.Println("List")

	if ps.Profile == "" {
		listProfiles()
		return
	}

	if !IsValidProfileName(ps.Profile) {
		fmt.Println(fmt.Sprintf("The requested profile %s does not exist in .aws/credentials", ps.Profile))
		return
	}

	if ps.Region == "" {
		listRegions()
		return
	}

	if ps.Cluster == "" {
		listClusters(ps)
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

func listClusters(ps PathStructure) {
	fmt.Println("List clusters")

	// Load session from shared config
	spew.Dump(ps)

	//sess := session.Must(session.NewSessionWithOptions(session.Options{
	//	// TODO: Region should be changeable
	//	Config: aws.Config{Region: aws.String("eu-west-1")},
	//	Profile: "jim-tech",
	//}))
	//
	//svc := ecs.New(sess)
	//
	//input := &ecs.ListClustersInput{}
	//
	//result, err := svc.ListClusters(input)
	//if err != nil {
	//	if aerr, ok := err.(awserr.Error); ok {
	//		switch aerr.Code() {
	//		case ecs.ErrCodeServerException:
	//			fmt.Println(ecs.ErrCodeServerException, aerr.Error())
	//		case ecs.ErrCodeClientException:
	//			fmt.Println(ecs.ErrCodeClientException, aerr.Error())
	//		case ecs.ErrCodeInvalidParameterException:
	//			fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
	//		default:
	//			fmt.Println(aerr.Error())
	//		}
	//	} else {
	//		// Print the error, cast err to awserr.Error to get the Code and
	//		// Message from an error.
	//		fmt.Println(err.Error())
	//	}
	//	return
	//}
	//
	//fmt.Println(result)
}

func listServices() {

}

func listTasks() {

}
