package main

import (
	"github.com/urfave/cli"
)

type PathStructure struct {
	Profile string
	Region string
	Cluster string
	Task string
	Container string
}

func (s *PathStructure) ParseFlags (c *cli.Context) {

	if c.String("profile") != "" {
		s.Profile = c.String("profile")
	}

	if c.String("region") != "" {
		s.Region = c.String("region")
	}

	// TODO: Region
	// TODO: Cluster
	// TODO: Task
	// TODO: Container
}