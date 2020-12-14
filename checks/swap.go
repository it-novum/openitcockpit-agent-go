package checks

import (
	"context"

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
	Total   uint64  `json:"total"`
	Percent float64 `json:"percent"`
	Used    uint64  `json:"used"`
	Free    uint64  `json:"free"`
	Sin     uint64  `json:"sin"`
	Sout    uint64  `json:"sout"`
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckSwap) Run(ctx context.Context) (*CheckResult, error) {
	s, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return &CheckResult{
		Result: &resultSwap{
			Total:   s.Total,
			Percent: s.UsedPercent,
			Used:    s.Used,
			Free:    s.Free,
			Sin:     s.Sin,
			Sout:    s.Sout,
		},
	}, nil
}

// DefaultConfiguration contains the variables for the configuration file and the default values
// can be nil if no configuration is required
func (c *CheckSwap) DefaultConfiguration() interface{} {
	return nil
}

// Configure should verify the configuration and set it
// will be run after every reload
// if DefaultConfiguration returns nil, the parameter will also be nil
func (c *CheckSwap) Configure(_ interface{}) error {
	return nil
}
