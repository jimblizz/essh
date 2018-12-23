package main

/**
	Long form:
		essh <action> --profile <profile name> --region <region> --cluster <cluster name> --task <task name> --container <container>
		essh <action> -p <profile name> -r <region> -c <cluster name> -task <task name> -co <container>

	Short form:
		essh connect -path myprofile>eu-west-1>mycluster>mytask>containerid

	Shortest form:
		At each level, can also use the ID number from the "list" views you get
		essh connect -p 2>3>mycluster>task>2


	Path format is always the structure:

		PROFILE > REGION > CLUSTER > TASK > CONTAINER


	Actions:
		man
		version
		list (default)
		connect (open ssh session)

 */

import (
	"os"
	"log"
	"os/exec"

	"github.com/urfave/cli"
	"github.com/fatih/color"
)

var tblHeaderFmt = color.New(color.FgGreen, color.Underline).SprintfFunc()
var tblColumnFmt = color.New(color.FgYellow).SprintfFunc()

func main()  {

	cmd := exec.Command("clear") // TODO: Support Windows
	cmd.Stdout = os.Stdout
	cmd.Run()

	app := cli.NewApp()
	app.Name = "essh"
	app.Usage = "Perform ssh based actions on AWS ECS task containers"
	app.Version = "0.0.1"
	app.Author = "James Blizzard - jim@acidhl.co.uk"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "profile",
			Usage: "The AWS profile name",
		},
		cli.StringFlag{
			Name: "region",
			Usage: "The AWS region name",
		},
		cli.StringFlag{
			Name: "cluster",
			Usage: "The ECS cluster name",
		},
        cli.StringFlag{
            Name: "service",
            Usage: "The ECS service name",
        },
		cli.StringFlag{
			Name: "container",
			Usage: "The running container to connect to",
		},
	}

	app.Action = func(c *cli.Context) error {

		// Before we can do anything, we need to load the profile data
		loadProfiles()

		action := c.Args().Get(0)
		if action == "" {
			action = "list"
		}

		ps := PathStructure{}
		ps.ParseFlags(c)

		// TODO: Switch through actions
		list(ps)

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}