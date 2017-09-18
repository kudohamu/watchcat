package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "watchcat"
	app.Version = "v0.1.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "conf",
			Usage: "path of watchcat's configuration",
		},
		cli.StringFlag{
			Name:  "notifiers",
			Usage: "notification parties (std, slack)",
		},
		cli.StringFlag{
			Name:  "slack_webhook_url",
			Usage: "webhook url for notifying to slack",
		},
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name:    "watch",
			Aliases: []string{"w"},
			Action:  Watch,
		},
	}
	app.Run(os.Args)
}
