package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/utils"
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

// Unfortunately the WMI library is suffering from a memory leak
// especially on windows Server 2016 and Windows 10.
// For this reason all WMI queries have been moved to an external binary (fork -> exec) to avoid any memory issues.
//
// Hopefully the memory issues will be fixed one day.
// This check used to look like this: https://github.com/it-novum/openitcockpit-agent-go/blob/a8ec01146e419a2db246844ca95cbe4ea560d9e6/checks/swap_windows.go

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckSwap) Run(ctx context.Context) (interface{}, error) {
	// exec wmiexecutor.exe to avoid memory leak
	timeout := 10 * time.Second
	commandResult, err := utils.RunCommand(ctx, utils.CommandArgs{
		Command: c.WmiExecutorPath + " --command swap",
		Shell:   "",
		Timeout: timeout,
		Env: []string{
			"OITC_AGENT_WMI_EXECUTOR=1",
		},
	})

	if err != nil {
		return nil, err
	}

	if commandResult.RC > 0 {
		return nil, fmt.Errorf(commandResult.Stdout)
	}

	var dst []*Win32_PageFileUsage
	err = json.Unmarshal([]byte(commandResult.Stdout), &dst)

	if err != nil {
		return nil, err
	}

	var info *Win32_PageFileUsage = dst[0]

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
