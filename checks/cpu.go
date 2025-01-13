package checks

import (
	"github.com/it-novum/openitcockpit-agent-go/config"
)

// CheckCpu gathers information about system CPU load
type CheckCpu struct {
	checkInterval int64
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
	c.checkInterval = config.CheckInterval // check interval in seconds (default: 30)
	return config.CPU, nil
}
