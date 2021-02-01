package checks

import (
	"context"

	"github.com/it-novum/openitcockpit-agent-go/config"
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
func (c *CheckLoad) Run(ctx context.Context) (interface{}, error) {
	l, err := load.AvgWithContext(ctx)

	if err != nil {
		return nil, err
	}
	return &resultLoad{
		Load1:  l.Load1,
		Load5:  l.Load5,
		Load15: l.Load15,
	}, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckLoad) Configure(config *config.Configuration) (bool, error) {
	return config.Load, nil
}
