package checks

import (
	"context"

	"github.com/yusufpapurcu/wmi"
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

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckCpu) Run(ctx context.Context) (interface{}, error) {

	var dst []Win32_PerfFormattedData_PerfOS_Processor
	err := wmi.Query("SELECT * FROM Win32_PerfFormattedData_PerfOS_Processor", &dst)
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
