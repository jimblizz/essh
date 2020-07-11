package service

import (
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"os/user"
	"sort"
)

var profiles = make([]AwsProfile, 0)

type AwsProfile struct {
	Name            string
	AccessKeyId     string
	SecretAccessKey string
}

func loadProfiles() {
	usr, usrErr := user.Current()
	if usrErr != nil {
		fmt.Println(usrErr.Error())
		os.Exit(0)
	}

	cfg, err := ini.Load(usr.HomeDir + "/.aws/credentials")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(0)
	}

	for _, val := range cfg.Sections() {
		if val.Key("aws_secret_access_key").String() != "" {
			profiles = append(profiles, AwsProfile{
				Name:            val.Name(),
				AccessKeyId:     val.Key("aws_access_key_id").String(),
				SecretAccessKey: val.Key("aws_secret_access_key").String(),
			})
		}
	}

	sort.Slice(profiles[:], func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})
}