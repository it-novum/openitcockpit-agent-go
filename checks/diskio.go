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
	//Meta data
	Timestamp int64  // Timestamp of the last check evaluation
	Device    string // Name of the disk

	// From iostat
	ReadBytes  uint64 // Number of bytes read from disk (Counter)
	WriteBytes uint64 // Number of bytes written to disk (Counter)
	ReadCount  uint64 // Number of read iops (Counter)
	WriteCount uint64 // Number of write iops (Counter)
	ReadTime   uint64 // Number of io_wait for read iops in milliseconds
	WriteTime  uint64 // Number of io_wait for write iops in milliseconds
	IoTime     uint64 // Time spent doing actual I/Os (in milliseconds) (busy_time in python psutil) (Counter)

	// Gets calculated
	ReadIopsPerSecond   uint64  // Number of read iops per second
	WriteIopsPerSecond  uint64  // Number of write iops per second
	TotalIopsPerSecond  uint64  // Number of read and write iops per second
	ReadBytesPerSecond  uint64  // Number of bytes read from disk per second
	WriteBytesPerSecond uint64  // Number of bytes written to disk per second
	TotalAvgWait        float64 // Total io_wait in milliseconds
	ReadAvgWait         float64 // Average io_wait for read iops in milliseconds
	WriteAvgWait        float64 // Average io_wait for write iops in milliseconds
	ReadAvgSize         float64 // Average request size of reads in bytes
	WriteAvgSize        float64 // Average request size of writes in bytes
	LoadPercent         float64 // Disk load as percentage

}

// Configure the command or return false if the command was disabled
func (c *CheckDiskIo) Configure(config *config.Configuration) (bool, error) {
	return config.DiskIo, nil
}
