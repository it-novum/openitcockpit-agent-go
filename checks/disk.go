package checks

import (
	"context"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/shirou/gopsutil/v3/disk"
)

// CheckDisk gathers information about system disks
type CheckDisk struct {
}

// Name will be used in the response as check name
func (c *CheckDisk) Name() string {
	return "disks"
}

type resultDisk struct {
	Disk struct {
		Device     string   `json:"device"`
		Mountpoint string   `json:"mountpoint"`
		Fstype     string   `json:"fstype"`
		Opts       []string `json:"opts"`
	}
	Usage struct {
		Total   uint64  `json:"total"`
		Used    uint64  `json:"used"`
		Free    uint64  `json:"free"`
		Percent float64 `json:"percent"`
	}
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckDisk) Run(ctx context.Context) (*CheckResult, error) {

	disks, err := disk.PartitionsWithContext(ctx, true)
	if err != nil {
		return nil, err
	}
	diskResults := make([]*resultDisk, 0, len(disks))

	for _, device := range disks {
		usage, _ := disk.UsageWithContext(ctx, device.Mountpoint)

		result := &resultDisk{}

		result.Disk.Device = device.Device
		result.Disk.Mountpoint = device.Mountpoint
		result.Disk.Fstype = device.Fstype
		result.Disk.Opts = device.Opts

		result.Usage.Total = usage.Total
		result.Usage.Used = usage.Used
		result.Usage.Free = usage.Free
		result.Usage.Percent = usage.UsedPercent

		diskResults = append(diskResults, result)
	}

	return &CheckResult{Result: diskResults}, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckDisk) Configure(config *config.Configuration) (bool, error) {
	return config.Diskstats, nil
}
