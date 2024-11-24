package testdata

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

func NewApp() *cli.App {
	return &cli.App{
		Name:  "test",
		Usage: "A simple test CLI app",
		Commands: []*cli.Command{
			{
				Name:  "greet",
				Usage: "Greet someone",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Name of the person to greet",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					fmt.Fprintf(c.App.Writer, "Hello %s", c.String("name")) //nolint:errcheck

					return nil
				},
			},
		},
	}
}
