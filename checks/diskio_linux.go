package checks

import (
	"context"
	"fmt"
	"math"

	"github.com/it-novum/openitcockpit-agent-go/config"
)

// CheckDiskIo gathers information about system disks IO
type CheckDiskIo struct {
	lastResults []*resultDiskIo
}

// Name will be used in the response as check name
func (c *CheckDiskIo) Name() string {
	return "disk_io"
}

type resultDiskIo struct {
	ReadBytes    uint64
	WriteBytes   uint64
	ReadIops     uint64
	WriteIops    uint64
	TotalIops    uint64
	ReadCount    uint64
	WriteCount   uint64
	IoTime       uint64
	ReadAvgWait  float64
	ReadTime     uint64
	ReadAvgSize  float64
	WriteAvgWait float64
	WriteAvgSize float64
	WriteTime    uint64
	TotalAvgWait float64
	LoadPercent  int64
	Timestamp    int64
	Device       string
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckDiskIo) Run(ctx context.Context) (*CheckResult, error) {
	// https://www.kernel.org/doc/html/latest/admin-guide/abi-testing.html#symbols-under-proc-diskstats
	// https://github.com/giampaolo/psutil/blob/f18438d135c12f7eb186f49622e0f6683c37f7f5/psutil/_pslinux.py#L1093
}

//Wrapdiff calculate the difference between last and curr
//If last > curr, try to guess the boundary at which the value must have wrapped
//by trying the maximum values of 64, 32 and 16 bit signed and unsigned ints.
func (c *CheckDiskIo) Wrapdiff(last, curr float64) (float64, error) {
	if last <= curr {
		return curr - last, nil
	}

	boundaries := []float64{64, 63, 32, 31, 16, 15}
	var currBoundary float64
	for _, boundary := range boundaries {
		if last > math.Pow(2, boundary) {
			currBoundary = boundary
		}
	}

	if currBoundary == 0 {
		return 0, fmt.Errorf("Couldn't determine boundary")
	}

	return math.Pow(2, currBoundary) - last + curr, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckDiskIo) Configure(config *config.Configuration) (bool, error) {
	return config.DiskIo, nil
}
