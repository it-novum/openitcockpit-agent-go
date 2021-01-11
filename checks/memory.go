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
	Total     uint64  `json:"total"`     // Total amount of memory (RAM) in bytes
	Available uint64  `json:"available"` // Available memory in bytes (inactive_count + free_count)
	Percent   float64 `json:"percent"`   // Used memory as percentage
	Used      uint64  `json:"used"`      // Used memory in bytes (totalCount - availableCount)
	Free      uint64  `json:"free"`      // Free memory in bytes
	Active    uint64  `json:"active"`    // Active memory in bytes
	Inactive  uint64  `json:"inactive"`  // Inactive memory in bytes
	Wired     uint64  `json:"wired"`     // Wired memory in bytes - macOS and BSD only - memory that is marked to always stay in RAM. It is never moved to disk
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckMem) Run(ctx context.Context) (interface{}, error) {
	v, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return &resultMemory{
		Total:     v.Total,
		Available: v.Available,
		Percent:   v.UsedPercent,
		Used:      v.Used,
		Free:      v.Free,
		Inactive:  v.Inactive,
		Wired:     v.Wired,
	}, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckMem) Configure(config *config.Configuration) (bool, error) {
	return true, nil
}
