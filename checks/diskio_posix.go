package checks

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/disk"
)

// CheckDiskIo gathers information about system disks IO
type CheckDiskIo struct {
	lastResult *resultDiskIo
}

// Name will be used in the response as check name
func (c *CheckDiskIo) Name() string {
	return "disk_io"
}

type resultDiskIo struct {
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
func (c *CheckDiskIo) Run(ctx context.Context) (*CheckResult, error) {

	disks, err := disk.PartitionsWithContext(ctx, true)
	if err != nil {
		return nil, err
	}
	diskResults := make([]*resultDiskIo, 0, len(disks))

	for _, device := range disks {
		usage, _ := disk.IOCountersWithContext(ctx)
		fmt.Println(usage)
		fmt.Println(device)

		result := &resultDiskIo{}

		diskResults = append(diskResults, result)
	}

	return &CheckResult{Result: diskResults}, nil
}

// DefaultConfiguration contains the variables for the configuration file and the default values
// can be nil if no configuration is required
func (c *CheckDiskIo) DefaultConfiguration() interface{} {
	return nil
}

// Configure should verify the configuration and set it
// will be run after every reload
// if DefaultConfiguration returns nil, the parameter will also be nil
func (c *CheckDiskIo) Configure(_ interface{}) error {
	return nil
}
