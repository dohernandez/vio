package main

import (
	"log"
	"os"

	"github.com/dohernandez/vio/internal/platform/cli"
)

func main() {
	appcli := cli.NewCliApp()

	err := appcli.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
