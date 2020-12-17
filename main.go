package main

import (
	"os"

	"github.com/it-novum/openitcockpit-agent-go/cmd"
)

func main() {
	if err := cmd.New().Execute(); err != nil {
		os.Exit(1)
	}
}
