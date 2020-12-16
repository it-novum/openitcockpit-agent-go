package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configPath string
	verbose    bool
)

var rootCmd *cobra.Command = nil

func initCommand() {
	rootCmd = &cobra.Command{
		Use:   "openitcockpit-agent",
		Short: "openitcockpit-agent collects system metrics for openitcockpit",
		Long:  `openitcockpit-agent collects system metrics for openitcockpit`,
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				return fmt.Errorf("--config \"%s\" does not exist", configPath)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.ini", "Path to configuration file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug output")
}

// Execute command line handling
func Execute() error {
	if rootCmd == nil {
		initCommand()
	}
	return rootCmd.Execute()
}
