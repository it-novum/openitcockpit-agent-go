package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var (
	configPath       string
	verbose          bool
	logPath          string
	testLogPath      string
	disableLog       bool
	disableLogRotate bool
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
			if !disableLog {
				if logPath == "" {
					logPath = platformLogFile()
					if logPath == "" { // windows no registry
						return fmt.Errorf("No logfile given and no registry InstallLocation set")
					}
				}
				fl, err := os.OpenFile(logPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
				if err != nil {
					return fmt.Errorf("Could not open/create log file: %s", err)
				}
				fl.Close()
				if !disableLogRotate {
					testPath := path.Join(path.Dir(logPath), "agent.log.test")
					fl, err := os.OpenFile(testPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
					defer func() {
						fl.Close()
						os.Remove(testPath)
					}()
					if err != nil {
						return fmt.Errorf("Test for log file rotation was not successful: %s (please check directory permissions)", err)
					}
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.ini", "Path to configuration file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug output")
	rootCmd.PersistentFlags().StringVarP(&logPath, "log", "l", "", "Set alternative path for log file output")
	rootCmd.PersistentFlags().BoolVar(&disableLog, "disable-logfile", false, "disable log file")
	rootCmd.PersistentFlags().BoolVar(&disableLogRotate, "disable-logrotate", false, "disable log file rotation")
}

// Execute command line handling
func Execute() error {
	if rootCmd == nil {
		initCommand()
	}
	return rootCmd.Execute()
}
