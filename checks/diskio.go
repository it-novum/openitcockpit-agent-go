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

	// From iostat (Linux and macOS)
	ReadBytes  uint64 // Number of bytes read from disk (Counter)
	WriteBytes uint64 // Number of bytes written to disk (Counter)
	ReadCount  uint64 // Number of read iops (Counter)
	WriteCount uint64 // Number of write iops (Counter)
	ReadTime   uint64 // Number of io_wait for read iops in milliseconds
	WriteTime  uint64 // Number of io_wait for write iops in milliseconds
	IoTime     uint64 // Time spent doing actual I/Os (in milliseconds) (busy_time in python psutil) (Counter)

	// Microsoft Windows WMI
	Frequency_PerfTime        uint64 // Frequency, in ticks per second, of Timestamp_Perftime
	Timestamp_Sys100NS        uint64 // Frequency, in ticks per second, of Timestamp_Sys100NS (10000000)
	AvgDiskSecPerRead         uint32 // Average time, in seconds, of a read operation of data from the disk.
	AvgDiskSecPerRead_Base    uint32 // Base value for AvgDiskSecPerRead. This value represents the accumulated number of operations that occurred.
	AvgDiskSecPerWrite        uint32 // Average time, in seconds, of a write operation of data to the disk.
	AvgDiskSecPerWrite_Base   uint32 // Base value for AvgDiskSecPerWrite. This value represents the accumulated number of operations that occurred.
	PercentDiskTime           uint64 // Percentage of elapsed time that the selected disk drive is busy servicing read or write requests.
	PercentDiskTime_Base      uint64 // Base value for PercentDiskTime.
	DiskReadsPerSec           uint32 // Rate of read operations on the disk.
	DiskWritesPerSec          uint32 // Rate of write operations on the disk.
	DiskReadBytesPerSec       uint64 // Rate at which bytes are transferred from the disk during read operations.
	DiskWriteBytesPerSec      uint64 // Rate at which bytes are transferred to the disk during write operations.
	AvgDiskBytesPerRead       uint64 // Average number of bytes transferred from the disk during read operations.
	AvgDiskBytesPerRead_Base  uint32 // Base value for AvgDiskBytesPerRead. This value represents the accumulated number of operations that occurred.
	AvgDiskBytesPerWrite      uint64 // Average number of bytes transferred to the disk during write operations.
	AvgDiskBytesPerWrite_Base uint64 // Base value for AvgDiskBytesPerWrite. This value represents the accumulated number of operations that occurred.

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
