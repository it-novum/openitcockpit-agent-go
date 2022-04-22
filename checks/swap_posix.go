//go:build linux || darwin
// +build linux darwin

package checks

import (
	"context"

	"github.com/shirou/gopsutil/v3/mem"
)

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
