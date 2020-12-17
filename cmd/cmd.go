package cmd

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/spf13/cobra"
)

type platformPathInterface interface {
	Init() error
	LogPath() string
	ConfigPath() string
}

type RootCmd struct {
	cmd              *cobra.Command
	configPath       string
	verbose          bool
	logPath          string
	disableLog       bool
	disableLogRotate bool
	platformPath     platformPathInterface
}

func (r *RootCmd) preRun(cmd *cobra.Command, args []string) error {
	r.platformPath.Init()

	if r.configPath == "" {
		if platformConfigPath := r.platformPath.ConfigPath(); platformConfigPath != "" {
			r.configPath = platformConfigPath
		} else {
			msg := "No config.ini path given"
			if runtime.GOOS == "windows" {
				msg = msg + " (probably missing windows registry InstallLocation)"
			}
			return fmt.Errorf(msg)
		}
	}
	if !r.disableLog && r.logPath == "" {
		if platformLogPath := r.platformPath.LogPath(); platformLogPath != "" {
			r.logPath = platformLogPath
		} else {
			msg := "No log file path given"
			if runtime.GOOS == "windows" {
				msg = msg + " (probably missing windows registry InstallLocation)"
			}
			return fmt.Errorf(msg)
		}
	}

	if _, err := os.Stat(r.configPath); os.IsNotExist(err) {
		return fmt.Errorf("--config \"%s\" does not exist", r.configPath)
	}

	if !r.disableLog {
		fl, err := os.OpenFile(r.logPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return fmt.Errorf("Could not open/create log file: %s", err)
		}
		fl.Close()
		if !r.disableLogRotate {
			testPath := path.Join(path.Dir(r.logPath), "agent.log.test")
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
}

func (r *RootCmd) run(cmd *cobra.Command, args []string) {

}

func New() *RootCmd {
	r := &RootCmd{}
	r.cmd = &cobra.Command{
		Use:     "openitcockpit-agent",
		Short:   "openitcockpit-agent collects system metrics for openitcockpit",
		Long:    `openitcockpit-agent collects system metrics for openitcockpit`,
		Args:    cobra.NoArgs,
		PreRunE: r.preRun,
		Run:     r.run,
	}
	r.cmd.PersistentFlags().StringVarP(&r.configPath, "config", "c", "", "Path to configuration file")
	r.cmd.PersistentFlags().BoolVarP(&r.verbose, "verbose", "v", false, "Enable debug output")
	r.cmd.PersistentFlags().StringVarP(&r.logPath, "log", "l", "", "Set alternative path for log file output")
	r.cmd.PersistentFlags().BoolVar(&r.disableLog, "disable-logfile", false, "disable log file")
	r.cmd.PersistentFlags().BoolVar(&r.disableLogRotate, "disable-logrotate", false, "disable log file rotation")

	r.platformPath = getPlatformPath()

	return r
}

func (r *RootCmd) Execute() error {
	return r.cmd.Execute()
}