//go:build linux || darwin
// +build linux darwin

package checks

import (
	"context"

	"github.com/shirou/gopsutil/v3/disk"
	log "github.com/sirupsen/logrus"
)

var devToIgnore = map[string]bool{
	"sysfs":       true,
	"proc":        true,
	"udev":        true,
	"devpts":      true,
	"devfs":       true,
	"tmpfs":       true,
	"securityfs":  true,
	"cgroup":      true,
	"cgroup2":     true,
	"pstore":      true,
	"debugfs":     true,
	"hugetlbfs":   true,
	"systemd-1":   true,
	"mqueue":      true,
	"none":        true,
	"sunrpc":      true,
	"nfsd":        true,
	"nsfs":        true,
	"fusectl":     true,
	"configfs":    true,
	"overlay":     true,
	"shm":         true,
	"tracefs":     true,
	"binfmt_misc": true,
}

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
		if devToIgnore[device.Device] {
			continue
		}

		// This will also ignore the main disk for LXC containers :(
		//if strings.HasPrefix(device.Device, "/dev/loop") {
		//	continue
		//}

		usage, err := disk.UsageWithContext(ctx, device.Mountpoint)

		if err != nil {
			log.Errorln("DiskCheck: Error for ", device.Mountpoint, err)
		}

		result := &resultDisk{}

		result.Disk.Device = device.Device
		result.Disk.Mountpoint = device.Mountpoint
		result.Disk.Fstype = device.Fstype
		result.Disk.Opts = device.Opts

		if usage != nil {
			result.Usage.Total = usage.Total
			result.Usage.Used = usage.Used
			result.Usage.Free = usage.Free
			result.Usage.Percent = usage.UsedPercent
		}

		diskResults = append(diskResults, result)
	}

	return diskResults, nil
}
