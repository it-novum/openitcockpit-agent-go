package checks

import (
	"context"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/shirou/gopsutil/v3/mem"
)

// CheckMem gathers information about system memory
type CheckMem struct {
}

// Name will be used in the response as check name
func (c *CheckMem) Name() string {
	return "memory"
}

type resultMemory struct {
	Total     uint64  `json:"total"`
	Available uint64  `json:"available"`
	Percent   float64 `json:"percent"`
	Used      uint64  `json:"used"`
	Free      uint64  `json:"free"`
	Active    uint64  `json:"active"`
	Inactive  uint64  `json:"inactive"`
	Wired     uint64  `json:"wired"`
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckMem) Run(ctx context.Context) (*CheckResult, error) {
	v, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return &CheckResult{
		Result: &resultMemory{
			Total:     v.Total,
			Available: v.Available,
			Percent:   v.UsedPercent,
			Used:      v.Used,
			Free:      v.Free,
			Inactive:  v.Inactive,
			Wired:     v.Wired,
		},
	}, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckMem) Configure(config *config.Configuration) (bool, error) {
	return true, nil
}
