package checks

import (
	"context"

	"github.com/shirou/gopsutil/v3/process"
)

// CheckProcess gathers information about each process
type CheckProcess struct {
}

type resultProcess struct {
	Pid           int32
	Ppid          int32
	Username      string
	Name          string
	CPUPercent    float64
	MemoryPercent float32
}

// Name will be used in the response as check name
func (c *CheckProcess) Name() string {
	return "processes"
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckProcess) Run(ctx context.Context) (*CheckResult, error) {
	var err error

	pids, err := process.PidsWithContext(ctx)
	if err != nil {
		return nil, err
	}

	processResults := make([]*resultProcess, 0, len(pids))

	// TODO log errors
	for _, pid := range pids {
		p, err := process.NewProcessWithContext(ctx, pid)
		if err != nil {
			// We ignore errors, because we the process just might have stopped or
			// is inaccessible in some way
			continue
		}
		result := &resultProcess{
			Pid: p.Pid,
		}
		result.Username, _ = p.UsernameWithContext(ctx)
		result.CPUPercent, _ = p.CPUPercentWithContext(ctx)
		result.MemoryPercent, _ = p.MemoryPercentWithContext(ctx)
		if parent, err := p.ParentWithContext(ctx); err == nil {
			result.Ppid = parent.Pid
		}

		processResults = append(processResults, result)
	}
	return &CheckResult{Result: processResults}, nil
}

// DefaultConfiguration contains the variables for the configuration file and the default values
// can be nil if no configuration is required
func (c *CheckProcess) DefaultConfiguration() interface{} {
	return nil
}

// Configure should verify the configuration and set it
// will be run after every reload
// if DefaultConfiguration returns nil, the parameter will also be nil
func (c *CheckProcess) Configure(_ interface{}) error {
	return nil
}
