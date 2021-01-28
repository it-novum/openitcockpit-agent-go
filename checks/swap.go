package checks

import (
	"context"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/shirou/gopsutil/v3/mem"
)

// CheckSwap gathers information about system swap
type CheckSwap struct {
}

// Name will be used in the response as check name
func (c *CheckSwap) Name() string {
	return "swap"
}

type resultSwap struct {
	Total   uint64  `json:"total"`   // Total amount of swap space in bytes
	Percent float64 `json:"percent"` // Used swap space as percentage
	Used    uint64  `json:"used"`    // Used swap space in bytes
	Free    uint64  `json:"free"`    // Free swap space in bytes
	Sin     uint64  `json:"sin"`     // Linux only - Number of bytes the system has swapped in from disk
	Sout    uint64  `json:"sout"`    // Linux only - Number of bytes the system has swapped out to disk
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckSwap) Run(ctx context.Context) (interface{}, error) {
	s, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return &resultSwap{
		Total:   s.Total,
		Percent: s.UsedPercent,
		Used:    s.Used,
		Free:    s.Free,
		Sin:     s.Sin,
		Sout:    s.Sout,
	}, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckSwap) Configure(config *config.Configuration) (bool, error) {
	return config.Swap, nil
}
