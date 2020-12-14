package checks

import (
	"context"

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

// DefaultConfiguration contains the variables for the configuration file and the default values
// can be nil if no configuration is required
func (c *CheckMem) DefaultConfiguration() interface{} {
	return nil
}

// Configure should verify the configuration and set it
// will be run after every reload
// if DefaultConfiguration returns nil, the parameter will also be nil
func (c *CheckMem) Configure(_ interface{}) error {
	return nil
}
