//go:build linux || darwin
// +build linux darwin

package checks

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

func saturatingSub(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return 0
}

func calcuateBusy(t cpu.TimesStat) (float64, float64, float64) {
	busy := t.User + t.Nice + t.System + t.Irq + t.Softirq + t.Steal
	idle := t.Idle + t.Iowait
	total := busy + idle
	return busy, idle, total
}

func calculateUsagePercentage(prev, current cpu.TimesStat) float64 {
	// This is highly inspired by
	// https://github.com/shirou/gopsutil/blob/master/cpu/cpu.go#L107
	// and htop
	// https://github.com/htop-dev/htop/blob/main/linux/LinuxProcessList.c#L1948-L2006
	// and yes - htop is reading /proc/stat as we do
	prevBusy, _, prevTotal := calcuateBusy(prev)
	currentBusy, _, currentTotal := calcuateBusy(current)

	if currentBusy <= prevBusy {
		return 0
	}

	if currentTotal <= prevTotal {
		return 100
	}
	return math.Min(100, math.Max(0, (currentBusy-prevBusy)/(currentTotal-prevTotal)*100))
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckCpu) Run(ctx context.Context) (interface{}, error) {
	result := &resultCpu{}

	prevTimeStats, err := cpu.TimesWithContext(ctx, true)
	if err != nil {
		return nil, err
	}

	if err := c.SleepWithContext(ctx, 1*time.Second); err != nil {
		return nil, err
	}

	timeStats, err := cpu.TimesWithContext(ctx, true)
	if err != nil {
		return nil, err
	}

	if len(prevTimeStats) != len(timeStats) {
		return nil, fmt.Errorf("number of CPU cores has changed %v != %v", len(prevTimeStats), len(timeStats))
	}

	// Get CPU usage per core
	var cpuPercentages []float64
	for i := range prevTimeStats {
		cpuPercentage := calculateUsagePercentage(prevTimeStats[i], timeStats[i])
		cpuPercentages = append(cpuPercentages, cpuPercentage)
	}
	result.PercentagePerCore = cpuPercentages

	// Get total CPU usage as percentage
	var totalCpuPercentage float64
	for i := range cpuPercentages {
		totalCpuPercentage = totalCpuPercentage + cpuPercentages[i]
	}
	result.PercentageTotal = totalCpuPercentage / float64(len(cpuPercentages))

	// Get CPU timing details per core
	var timings []cpuDetails
	var totalUser, totalNice, totalSystem, totalIoWait, totalIdle, totalTotal float64
	for i := range prevTimeStats {
		// Ignore t.Irq + t.Softirq + t.Steal to get 100% over all values we list in the json
		prevTotal := prevTimeStats[i].User + prevTimeStats[i].Nice + prevTimeStats[i].System + prevTimeStats[i].Idle + prevTimeStats[i].Iowait
		currentTotal := timeStats[i].User + timeStats[i].Nice + timeStats[i].System + timeStats[i].Idle + timeStats[i].Iowait

		userDelta := saturatingSub(timeStats[i].User, prevTimeStats[i].User)
		niceDelta := saturatingSub(timeStats[i].Nice, prevTimeStats[i].Nice)
		systemDelta := saturatingSub(timeStats[i].System, prevTimeStats[i].System)
		ioWaitDelta := saturatingSub(timeStats[i].Iowait, prevTimeStats[i].Iowait)
		idleDelta := saturatingSub(timeStats[i].Idle, prevTimeStats[i].Idle)
		totalDelta := saturatingSub(currentTotal, prevTotal)

		if totalDelta == 0 {
			timings = append(timings, cpuDetails{})
		} else {
			// Do not Division by zero (not sure if this could ever happen to be honest but better safe than sorry)
			timings = append(timings, cpuDetails{
				User:   userDelta / totalDelta * 100,
				Nice:   niceDelta / totalDelta * 100,
				System: systemDelta / totalDelta * 100,
				Idle:   idleDelta / totalDelta * 100,
				Iowait: ioWaitDelta / totalDelta * 100,
			})
		}

		totalUser = totalUser + userDelta
		totalNice = totalNice + niceDelta
		totalSystem = totalSystem + systemDelta
		totalIoWait = totalIoWait + ioWaitDelta
		totalIdle = totalIdle + idleDelta
		totalTotal = totalTotal + totalDelta
	}
	result.DetailsPerCore = timings

	// Get total CPU timing details
	result.DetailsTotal = &cpuDetails{
		User:   totalUser / totalTotal * 100,
		Nice:   totalNice / totalTotal * 100,
		System: totalSystem / totalTotal * 100,
		Idle:   totalIdle / totalTotal * 100,
		Iowait: totalIoWait / totalTotal * 100,
	}

	return result, nil

}

// Like sleep but can be canceled by context
func (c *CheckCpu) SleepWithContext(ctx context.Context, interval time.Duration) error {
	var timer = time.NewTimer(interval)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
