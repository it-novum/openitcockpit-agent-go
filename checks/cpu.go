package checks

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/it-novum/openitcockpit-agent-go/config"
)

// CheckCpu gathers information about system CPU load
type CheckCpu struct {
	WmiExecutorPath string // Used for Windows
}

type cpuDetails struct {
	User   float64 // Linux, macOS, Windows  - Seconds
	Nice   float64 // Linux, macOS           - Seconds
	System float64 // Linux, macOS, Windows  - Seconds
	Idle   float64 // Linux, macOS, Windows  - Seconds
	Iowait float64 // Linux
}

type resultCpu struct {
	PercentageTotal   float64      `json:"cpu_total_percentage"`
	PercentagePerCore []float64    `json:"cpu_percentage"`
	DetailsTotal      *cpuDetails  `json:"cpu_total_percentage_detailed"`
	DetailsPerCore    []cpuDetails `json:"cpu_percentage_detailed"`
}

// Name will be used in the response as check name
func (c *CheckCpu) Name() string {
	return "cpu"
}

// Configure the command or return false if the command was disabled
func (c *CheckCpu) Configure(config *config.Configuration) (bool, error) {
	if config.CPU && runtime.GOOS == "windows" {
		// Check is enabled
		agentBinary, err := os.Executable()
		if err == nil {
			wmiPath := filepath.Dir(agentBinary) + string(os.PathSeparator) + "wmiexecutor.exe"
			c.WmiExecutorPath = wmiPath
		}
	}

	return config.CPU, nil
}
