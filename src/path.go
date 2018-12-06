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

	if c.String("cluster") != "" {
	    s.Cluster = c.String("cluster")
    }

	// TODO: Service
	// TODO: Task
}