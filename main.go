package main

import (
	"os"

	"github.com/it-novum/openitcockpit-agent-go/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
