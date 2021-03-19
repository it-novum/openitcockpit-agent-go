package checks

import (
	"context"
	"fmt"

	"github.com/StackExchange/wmi"
)

// https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-pagefileusage
type Win32_PageFileUsage struct {
	Caption           string
	Description       string
	Status            string
	AllocatedBaseSize uint32
	CurrentUsage      uint32
	Name              string
	PeakUsage         uint32
	TempPageFile      bool
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckSwap) Run(ctx context.Context) (interface{}, error) {

	var dst []Win32_PageFileUsage
	err := wmi.Query("SELECT * FROM Win32_PageFileUsage", &dst)
	if err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		return nil, fmt.Errorf("Empty result from WMI")
	}

	var info Win32_PageFileUsage = dst[0]

	total := info.AllocatedBaseSize * 1024 * 1024
	used := info.CurrentUsage * 1024 * 1024
	var free uint32 = 0
	var percent float64 = 0.0

	if total > 0 {
		free = total - used
		percent = float64(used) / float64(total) * 100.0
	}

	return &resultSwap{
		Total:   uint64(total),
		Percent: percent,
		Used:    uint64(used),
		Free:    uint64(free),
	}, nil
}
