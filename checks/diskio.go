package checks

import (
	"github.com/it-novum/openitcockpit-agent-go/config"
)

// CheckDiskIo gathers information about system disks IO
type CheckDiskIo struct {
	lastResults map[string]*resultDiskIo
}

// Name will be used in the response as check name
func (c *CheckDiskIo) Name() string {
	return "disk_io"
}

type resultDiskIo struct {
	ReadBytes    uint64  // Number of bytes read from disk since last evaluation
	WriteBytes   uint64  // Number of bytes written to disk since last evaluation
	ReadIops     uint64  // Number of read iops since last evaluation
	WriteIops    uint64  // Number of write iops since last evaluation
	TotalIops    uint64  // Total number of write iops since last evaluation
	ReadCount    uint64  // Number of read iops since last evaluation (same as ReadIops)
	WriteCount   uint64  // Number of write iops since last evaluation (same as WriteIops)
	IoTime       uint64  // Time spent doing actual I/Os (in milliseconds) (busy_time in python psutil)
	ReadAvgWait  float64 // Average io_wait for read iops in milliseconds
	ReadTime     uint64  // Number of io_wait for read iops in milliseconds
	ReadAvgSize  float64 // Average request size of reads in bytes
	WriteAvgWait float64 // Average io_wait for write iops in milliseconds
	WriteAvgSize float64 // Average request size of writes in bytes
	WriteTime    uint64  // Number of io_wait for write iops in milliseconds
	TotalAvgWait float64 // Total io_wait in milliseconds
	LoadPercent  int64   // Disk load as percentage
	Timestamp    int64   // Timestamp of the last check evaluation
	Device       string  // Name of the disk
}

// Configure the command or return false if the command was disabled
func (c *CheckDiskIo) Configure(config *config.Configuration) (bool, error) {
	return config.DiskIo, nil
}
