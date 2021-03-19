// +build linux darwin

package checks

import (
	"context"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/shirou/gopsutil/v3/mem"
)

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
