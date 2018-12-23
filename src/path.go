package main

import (
    "github.com/urfave/cli"
)

type PathStructure struct {
	Profile string
	Region string
	Cluster string
	Service string
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

	if c.String("service") != "" {
	    s.Service = c.String("service")
    }

	if c.String("container") != "" {
	    s.Container = c.String("container")
    }
}

func (s PathStructure) HasValidProfileName() bool {
    for _, profile := range profiles {
        if profile.Name == s.Profile {
            return true
        }
    }
    return false
}

func (s PathStructure) HasValidRegionName() bool {
    for _, region := range RegionsMock {
        if region == s.Region {
            return true
        }
    }
    return false
}

func (s PathStructure) HasValidClusterName() bool {
    clusters, err := GetClusterList(s)
    if err != nil {
        HandleAwsError(err)
        return false
    }

    for _, cluster := range clusters.Clusters {
        if *cluster.ClusterName == s.Cluster {
            return true
        }
    }
    return false
}

func (s PathStructure) HasValidServiceName() bool {

    services, err := GetServiceList(s)
    if err != nil {
        HandleAwsError(err)
        return false
    }

    for _, service := range services {
        if service.ServiceName == s.Service {
            return true
        }
    }
    return false
}

func (s PathStructure) HasValidContainerName() bool {

    noInstanceData := make(map[string]InstanceDigest)
    containers, err := GetContainerList(s, noInstanceData)
    if err != nil {
        HandleAwsError(err)
        return false
    }

    for _, container := range containers {
        if container.UUID == s.Container {
            return true
        }
    }
    return false
}