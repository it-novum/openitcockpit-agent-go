package checks

import (
	"context"

	"github.com/shirou/gopsutil/v3/load"
)

// CheckLoad gathers information about system CPU load
type CheckLoad struct {
}

// Name will be used in the response as check name
func (c *CheckLoad) Name() string {
	return "system_load"
}

type resultLoad struct {
	Load1  float64 `json:"0"`
	Load5  float64 `json:"1"`
	Load15 float64 `json:"2"`
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckLoad) Run(ctx context.Context) (*CheckResult, error) {
	l, err := load.AvgWithContext(ctx)

	if err != nil {
		return nil, err
	}
	return &CheckResult{
		Result: &resultLoad{
			Load1:  l.Load1,
			Load5:  l.Load5,
			Load15: l.Load15,
		},
	}, nil
}

// DefaultConfiguration contains the variables for the configuration file and the default values
// can be nil if no configuration is required
func (c *CheckLoad) DefaultConfiguration() interface{} {
	return nil
}

// Configure should verify the configuration and set it
// will be run after every reload
// if DefaultConfiguration returns nil, the parameter will also be nil
func (c *CheckLoad) Configure(_ interface{}) error {
	return nil
}
