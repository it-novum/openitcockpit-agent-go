package checks

import (
	"github.com/it-novum/openitcockpit-agent-go/config"
)

// CheckCpu gathers information about system CPU load
type CheckCpu struct {
}

// Name will be used in the response as check name
func (c *CheckCpu) Name() string {
	return "cpu"
}

type resultCpu struct {
	PercentageTotal   float64      `json:"cpu_total_percentage"`
	PercentagePerCore []float64    `json:"cpu_percentage"`
	DetailsTotal      *cpuDetails  `json:"cpu_total_percentage_detailed"`
	DetailsPerCore    []cpuDetails `json:"cpu_percentage_detailed"`
}

// Configure the command or return false if the command was disabled
func (c *CheckCpu) Configure(config *config.Configuration) (bool, error) {
	return config.CPU, nil
}
