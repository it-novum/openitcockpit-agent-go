// +build linux darwin

package checks

import (
	"context"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckCpu) Run(ctx context.Context) (interface{}, error) {
	result := &resultCpu{}

	cpuPercentages, err := cpu.PercentWithContext(ctx, 1*time.Second, true)

	if err == nil {
		result.PercentagePerCore = cpuPercentages
	}

	//get total CPU percentage
	cpuPercentageTotal, err := cpu.PercentWithContext(ctx, 1*time.Second, false)

	if err == nil {
		result.PercentageTotal = cpuPercentageTotal[0]
	}

	if runtime.GOOS == "linux" {
		timeStats, err := cpu.TimesWithContext(ctx, false)

		if err == nil {
			total := timeStats[0].User + timeStats[0].Nice + timeStats[0].System + timeStats[0].Idle + timeStats[0].Iowait + timeStats[0].Irq + timeStats[0].Softirq
			result.DetailsTotal = &cpuDetails{
				User:   timeStats[0].User / total * 100,
				Nice:   timeStats[0].Nice / total * 100,
				System: timeStats[0].System / total * 100,
				Iowait: timeStats[0].Iowait / total * 100,
			}
		}

		//get timings per CPU
		var timings []cpuDetails
		timeStats, err = cpu.TimesWithContext(ctx, true)
		if err == nil {
			for _, timeStat := range timeStats {
				total := timeStat.User + timeStat.Nice + timeStat.System + timeStat.Idle + timeStat.Iowait + timeStat.Irq + timeStat.Softirq
				timings = append(timings, cpuDetails{
					User:   timeStat.User / total * 100,
					Nice:   timeStat.Nice / total * 100,
					System: timeStat.System / total * 100,
					Idle:   timeStat.Idle / total * 100,
					Iowait: timeStat.Iowait / total * 100,
				})
			}
			result.DetailsPerCore = timings
		}
	}

	if runtime.GOOS == "darwin" {
		timeStats, err := cpu.TimesWithContext(ctx, false)

		if err == nil {
			total := timeStats[0].User + timeStats[0].Nice + timeStats[0].System + timeStats[0].Idle
			result.DetailsTotal = &cpuDetails{
				User:   timeStats[0].User / total * 100,
				Nice:   timeStats[0].Nice / total * 100,
				System: timeStats[0].System / total * 100,
				Idle:   timeStats[0].Idle / total * 100,
			}
		}

		//get timings per CPU
		var timings []cpuDetails
		timeStats, err = cpu.TimesWithContext(ctx, true)
		if err == nil {
			for _, timeStat := range timeStats {
				total := timeStat.User + timeStat.Nice + timeStat.System + timeStat.Idle
				timings = append(timings, cpuDetails{
					User:   timeStat.User / total * 100,
					Nice:   timeStat.Nice / total * 100,
					System: timeStat.System / total * 100,
					Idle:   timeStat.Idle / total * 100,
				})
			}
			result.DetailsPerCore = timings
		}
	}

	return result, nil

}
