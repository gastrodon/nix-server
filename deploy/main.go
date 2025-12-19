package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "deploy",
		Usage: "Deploy NixOS configuration to remote host(s)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "user",
				Aliases: []string{"u"},
				Value:   "root",
				Usage:   "SSH user for deployment",
			},
		},
		Action: func(c *cli.Context) error {
			user := c.String("user")
			hosts := c.Args().Slice()

			if len(hosts) == 0 {
				return fmt.Errorf("no hosts specified for deployment")
			}

			for _, host := range hosts {
				if err := Deploy(host, user); err != nil {
					fmt.Fprintf(os.Stderr, "Error deploying to %s: %v\n", host, err)
					continue
				}
			}

			fmt.Println("========================================")
			fmt.Println("Deployment complete for all hosts.")
			fmt.Println("========================================")

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
