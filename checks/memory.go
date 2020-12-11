package checks

import (
	"context"

	"github.com/shirou/gopsutil/v3/mem"
)

// CheckMem gathers information about system memory
type CheckMem struct {
}

// Name will be used in the response as check name
func (m *CheckMem) Name() string {
	return "memory"
}

type result struct {
	Total     uint64
	Available uint64
	Percent   float64
	Used      uint64
	Free      uint64
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (m *CheckMem) Run(ctx context.Context) (*CheckResult, error) {
	v, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return &CheckResult{
		Result: &result{
			Total:     v.Total,
			Available: v.Available,
			Percent:   v.UsedPercent,
			Used:      v.Used,
			Free:      v.Free,
		},
	}, nil
}

// DefaultConfiguration contains the variables for the configuration file and the default values
// can be nil if no configuration is required
func (m *CheckMem) DefaultConfiguration() interface{} {
	return nil
}

// Configure should verify the configuration and set it
// will be run after every reload
// if DefaultConfiguration returns nil, the parameter will also be nil
func (m *CheckMem) Configure(_ interface{}) error {
	return nil
}
