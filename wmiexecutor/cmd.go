package wmiexecutor

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

type RootCmd struct {
	cmd *cobra.Command

	// options
	command string
	verbose bool
	debug   bool

	// Maybe we need this one day
	shutdown chan struct{}
	wg       sync.WaitGroup
}

type Configuration struct {
	verbose bool
	debug   bool
}

func (r *RootCmd) preRun(cmd *cobra.Command, args []string) error {
	if r.command == "" {
		msg := "No command was given"
		return fmt.Errorf(msg)
	}

	// A map should be mutch faster than an array
	availableCommands := map[string]bool{
		"cpu":      true,
		"diskio":   true,
		"memory":   true,
		"process":  true,
		"services": true,
		"swap":     true,
		"eventlog": true,
		"net":      true,
	}

	if _, exists := availableCommands[r.command]; !exists {

		keys := make([]string, 0, len(availableCommands))
		for k := range availableCommands {
			keys = append(keys, k)
		}

		msg := fmt.Sprintf("No command '%v' exists. Valid commands are: %v", r.command, strings.Join(keys, ", "))
		return fmt.Errorf(msg)
	}

	return nil
}

func New() *RootCmd {
	r := &RootCmd{
		shutdown: make(chan struct{}),
	}
	r.cmd = &cobra.Command{
		Use:     "wmiexecutor",
		Short:   "wmiexecutor collects system metrics on Windows for openitcockpit-agent",
		Long:    `wmiexecutor collects system metrics on Windows for openitcockpit-agent`,
		Args:    cobra.NoArgs,
		PreRunE: r.preRun,
		Run:     r.run,
	}
	r.cmd.PersistentFlags().StringVarP(&r.command, "command", "c", "", "Command to execute")
	r.cmd.PersistentFlags().BoolVarP(&r.verbose, "verbose", "v", false, "Enable info output")
	r.cmd.PersistentFlags().BoolVarP(&r.debug, "debug", "d", false, "Enable debug output")

	return r
}

func (r *RootCmd) Execute() error {
	r.wg.Add(1)
	defer r.wg.Done()

	return r.cmd.Execute()
}

func (r *RootCmd) Shutdown() {
	close(r.shutdown)
	r.wg.Wait()
}

func (r *RootCmd) run(cmd *cobra.Command, args []string) {

	conf := &Configuration{
		verbose: r.verbose,
		debug:   r.debug,
	}

	var result string
	var err error
	switch r.command {
	case "cpu":
		cpu := &CheckCpu{}
		_ = cpu.Configure(conf)
		result, err = cpu.RunQuery()

	case "diskio":
		diskio := &CheckDiskIo{}
		diskio.Configure(conf)
		result, err = diskio.RunQuery()

	case "memory":
		mem := &CheckMem{}
		_ = mem.Configure(conf)
		result, err = mem.RunQuery()

	case "process":
		proc := &CheckProcess{}
		_ = proc.Configure(conf)
		result, err = proc.RunQuery()

	case "services":
		services := &CheckWinService{}
		_ = services.Configure(conf)
		result, err = services.RunCheck()

	case "swap":
		swap := &CheckSwap{}
		_ = swap.Configure(conf)
		result, err = swap.RunQuery()

	case "eventlog":
		eventlog := &CheckWindowsEventLog{}
		_ = eventlog.Configure(conf)
		result, err = eventlog.RunQuery()

	case "net":
		netstat := &CheckNet{}
		_ = netstat.Configure(conf)
		result, err = netstat.RunCheck()

	}

	if err != nil {
		// Print error messages as JSON
		js, _ := json.Marshal(err)
		fmt.Println(string(js))
		os.Exit(1)
	}

	if result == "" {
		fmt.Println("Error: Got empty result from check")
		os.Exit(2)
	}

	// Print results as JSON
	fmt.Println(result)
	//os.Exit(0)

}
