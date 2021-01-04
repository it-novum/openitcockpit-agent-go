package cmd

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/platformpaths"
	"github.com/it-novum/openitcockpit-agent-go/webserver"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type agentInstance struct {
	webserver *webserver.Server
}

type RootCmd struct {
	cmd              *cobra.Command
	configPath       string
	verbose          bool
	logPath          string
	disableLog       bool
	disableLogRotate bool
	platformPath     platformpaths.PlatformPath
	initDone         bool
	initCCCDone      bool

	mtx sync.Mutex
}

func (r *RootCmd) preRun(cmd *cobra.Command, args []string) error {
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

// TODO find a better place
func (r *RootCmd) reloadCCConfiguration(ccc *config.CustomCheckConfiguration, err error) {
	go func() {
		r.mtx.Lock()
		defer r.mtx.Unlock()

		if err != nil {
			log.Errorln(err)
			if !r.initCCCDone {
				// on the initial load we want to exit the program if there is an error
				os.Exit(1)
			}
			return
		}
		r.initCCCDone = true
		// do reload for custom checks
	}()
}

func (r *RootCmd) reloadConfiguration(cfg *config.Configuration, err error) {
	go func() {
		r.mtx.Lock()
		defer r.mtx.Unlock()

		firstLoad := !r.initDone

		if err != nil {
			log.Errorln(err)
			if firstLoad {
				// on the initial load we want to exit the program if there is an error
				os.Exit(1)
			}
			return
		}
		if firstLoad && cfg.CustomchecksConfig != "" {
			config.LoadCustomChecks(cfg.CustomchecksConfig, r.reloadCCConfiguration)
		}
		r.initDone = true
		// do reload for everything else
		// TODO
	}()
}

func (r *RootCmd) run(cmd *cobra.Command, args []string) {
	// TODO configure logging
	config.Load(r.reloadConfiguration, &config.LoadConfigHint{
		SearchPath: r.configPath,
	})
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

	r.platformPath = platformpaths.Get()

	return r
}

func (r *RootCmd) Execute() error {
	return r.cmd.Execute()
}
