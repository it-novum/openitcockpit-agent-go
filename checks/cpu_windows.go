package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/utils"
)

// https://wutils.com/wmi/root/cimv2/win32_perfformatteddata_perfos_processor/
type Win32_PerfFormattedData_PerfOS_Processor struct {
	PercentC1Time         uint64
	PercentC2Time         uint64
	PercentC3Time         uint64
	PercentDPCTime        uint64
	PercentIdleTime       uint64
	PercentInterruptTime  uint64
	PercentPrivilegedTime uint64 // system
	PercentProcessorTime  uint64 // total usage that taskmanager Shows
	PercentUserTime       uint64 // user
	Name                  string
}

// Unfortunately the WMI library is suffering from a memory leak
// especially on windows Server 2016 and Windows 10.
// For this reason all WMI queries have been moved to an external binary (fork -> exec) to avoid any memory issues.
//
// Hopefully the memory issues will be fixed one day.
// This check used to look like this: https://github.com/it-novum/openitcockpit-agent-go/blob/a8ec01146e419a2db246844ca95cbe4ea560d9e6/checks/cpu_windows.go

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckCpu) Run(ctx context.Context) (interface{}, error) {
	// exec wmiexecutor.exe to avoid memory leak
	timeout := 10 * time.Second
	commandResult, err := utils.RunCommand(ctx, utils.CommandArgs{
		Command: c.WmiExecutorPath + " --command cpu",
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

	var dst []*Win32_PerfFormattedData_PerfOS_Processor
	err = json.Unmarshal([]byte(commandResult.Stdout), &dst)

	if err != nil {
		return nil, err
	}

	result := &resultCpu{}
	var detailsPerCore []cpuDetails

	for _, cpu := range dst {
		if cpu.Name == "_Total" {
			result.PercentageTotal = float64(cpu.PercentProcessorTime)
			result.DetailsTotal = &cpuDetails{
				User:   float64(cpu.PercentUserTime),
				System: float64(cpu.PercentPrivilegedTime),
				Idle:   float64(cpu.PercentIdleTime),
			}
		} else {
			result.PercentagePerCore = append(result.PercentagePerCore, float64(cpu.PercentProcessorTime))
			detailsPerCore = append(detailsPerCore, cpuDetails{
				User:   float64(cpu.PercentUserTime),
				System: float64(cpu.PercentPrivilegedTime),
				Idle:   float64(cpu.PercentIdleTime),
			})
		}
	}

	result.DetailsPerCore = detailsPerCore

	return result, nil
}
