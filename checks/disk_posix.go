// +build linux darwin

package checks

import (
	"context"

	"github.com/shirou/gopsutil/v3/disk"
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckDisk) Run(ctx context.Context) (interface{}, error) {

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

	return diskResults, nil
}
