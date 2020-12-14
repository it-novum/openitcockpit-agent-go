package checks

import (
	"context"

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

// DefaultConfiguration contains the variables for the configuration file and the default values
// can be nil if no configuration is required
func (c *CheckDisk) DefaultConfiguration() interface{} {
	return nil
}

// Configure should verify the configuration and set it
// will be run after every reload
// if DefaultConfiguration returns nil, the parameter will also be nil
func (c *CheckDisk) Configure(_ interface{}) error {
	return nil
}
