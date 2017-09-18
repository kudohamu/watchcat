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
			Usage: "notification parties (std)",
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
