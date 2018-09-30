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
