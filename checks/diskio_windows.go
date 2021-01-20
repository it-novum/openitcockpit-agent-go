package checks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/it-novum/openitcockpit-agent-go/safemaths"
)

// WMI Structs
// https://docs.microsoft.com/en-us/previous-versions/aa394262(v=vs.85)
/*
type Win32_PerfFormattedData_PerfDisk_PhysicalDisk struct {
	AvgDiskBytesPerRead     uint64
	AvgDiskBytesPerTransfer uint64
	AvgDiskBytesPerWrite    uint64
	AvgDiskQueueLength      uint64
	AvgDiskReadQueueLength  uint64
	AvgDiskSecPerRead       uint32
	AvgDiskSecPerTransfer   uint32
	AvgDiskSecPerWrite      uint32
	AvgDiskWriteQueueLength uint64
	Caption                 string
	CurrentDiskQueueLength  uint32
	Description             string
	DiskBytesPerSec         uint64
	DiskReadBytesPerSec     uint64
	DiskReadsPerSec         uint32
	DiskTransfersPerSec     uint32
	DiskWriteBytesPerSec    uint64
	DiskWritesPerSec        uint32
	Frequency_Object        uint64
	Frequency_PerfTime      uint64
	Frequency_Sys100NS      uint64
	Name                    string
	PercentDiskReadTime     uint64
	PercentDiskTime         uint64
	PercentDiskWriteTime    uint64
	PercentIdleTime         uint64
	SplitIOPerSec           uint32
	Timestamp_Object        uint64
	Timestamp_PerfTime      uint64
	Timestamp_Sys100NS      uint64
}
*/

// https://docs.microsoft.com/en-us/previous-versions/aa394308(v=vs.85)
type Win32_PerfRawData_PerfDisk_PhysicalDisk struct {
	AvgDiskBytesPerRead          uint64
	AvgDiskBytesPerRead_Base     uint32
	AvgDiskBytesPerTransfer      uint64
	AvgDiskBytesPerTransfer_Base uint64
	AvgDiskBytesPerWrite         uint64
	AvgDiskBytesPerWrite_Base    uint64
	AvgDiskQueueLength           uint64
	AvgDiskReadQueueLength       uint64
	AvgDiskSecPerRead            uint32
	AvgDiskSecPerRead_Base       uint32
	AvgDiskSecPerTransfer        uint32
	AvgDiskSecPerTransfer_Base   uint32
	AvgDiskSecPerWrite           uint32
	AvgDiskSecPerWrite_Base      uint32
	AvgDiskWriteQueueLength      uint64
	Caption                      string
	CurrentDiskQueueLength       uint32
	Description                  string
	DiskBytesPerSec              uint64
	DiskReadBytesPerSec          uint64
	DiskReadsPerSec              uint32
	DiskTransfersPerSec          uint32
	DiskWriteBytesPerSec         uint64
	DiskWritesPerSec             uint32
	Frequency_Object             uint64
	Frequency_PerfTime           uint64
	Frequency_Sys100NS           uint64
	Name                         string
	PercentDiskReadTime          uint64
	PercentDiskReadTime_Base     uint64
	PercentDiskTime              uint64
	PercentDiskTime_Base         uint64
	PercentDiskWriteTime         uint64
	PercentDiskWriteTime_Base    uint64
	PercentIdleTime              uint64
	PercentIdleTime_Base         uint64
	SplitIOPerSec                uint32
	Timestamp_Object             uint64
	Timestamp_PerfTime           uint64
	Timestamp_Sys100NS           uint64
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckDiskIo) Run(ctx context.Context) (interface{}, error) {

	// Need some help on Windows Performance Counters? We all need - it's a nightmare
	// This is gold!
	// http://toncigrgin.blogspot.com/2015/11/windows-perf-counters-blog2.html?m=1
	// Counter types: https://docs.microsoft.com/de-de/windows/win32/wmisdk/countertype-qualifier?redirectedfrom=MSDN
	// Formulas: https://docs.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2003/cc785636(v=ws.10)?redirectedfrom=MSDN
	// https://docs.microsoft.com/en-us/archive/blogs/askcore/windows-performance-monitor-disk-counters-explained
	//
	// Many thanks to the_d3f4ult https://www.reddit.com/r/golang/comments/kxyhfo/get_windows_disk_io_latency_from_wmi_in/gjp5og4?utm_source=share&utm_medium=web2x&context=3
	// for making this happen!

	var dst []Win32_PerfRawData_PerfDisk_PhysicalDisk
	err := wmi.Query("SELECT * FROM Win32_PerfRawData_PerfDisk_PhysicalDisk WHERE Name <> '_Total'", &dst)
	if err != nil {
		return nil, err
	}

	diskResults := make(map[string]*resultDiskIo)

	for i, disk := range dst {
		diskName := disk.Name
		if diskName != "_Total" {
			//Remove the index from drive letter "0 C:" => "C:"
			diskName = strings.Replace(diskName, fmt.Sprintf("%v ", i), "", 1)
		}

		if lastCheckResults, ok := c.lastResults[diskName]; ok {
			// Disk latency in ms
			// Many thanks to the_d3f4ult https://www.reddit.com/r/golang/comments/kxyhfo/get_windows_disk_io_latency_from_wmi_in/gjp5og4?utm_source=share&utm_medium=web2x&context=3
			//
			// PERF_AVERAGE_TIMER
			// ((N1 - N0) / F) / (D1 - D0), where the numerator (N) represents the number of ticks counted during the last
			// sample interval, the variable F represents the frequency of the ticks, and the denominator (D) represents the number of operations completed during the last sample interval.
			//
			// N1 - N0 = AvgDiskSecPerRead F = Frequency_PerfTime D1 - D0 = AvgDiskSecPerWrite_Base
			// Then you multiply by the default scale which is 10 cubed (and is not included in the type formula).
			AvgDiskSecPerReadDiff := WrapDiffUint32(lastCheckResults.AvgDiskSecPerRead, disk.AvgDiskSecPerRead)               // N1 - N0 (Reads)
			AvgDiskSecPerReadBaseDiff := WrapDiffUint32(lastCheckResults.AvgDiskSecPerRead_Base, disk.AvgDiskSecPerRead_Base) // D1 - D0 (Reads)

			AvgDiskSecPerWriteDiff := WrapDiffUint32(lastCheckResults.AvgDiskSecPerWrite, disk.AvgDiskSecPerWrite)               // N1 - N0 (Writes)
			AvgDiskSecPerWriteBaseDiff := WrapDiffUint32(lastCheckResults.AvgDiskSecPerWrite_Base, disk.AvgDiskSecPerWrite_Base) // D1 - D0 (Writes)

			// nolint:ineffassign
			var AvgDiskSecPerRead_Us, AvgDiskSecPerRead_Ms float64 = 0.0, 0.0
			if AvgDiskSecPerReadDiff > 0 {
				AvgDiskSecPerRead_Us = safemaths.DivideFloat64(safemaths.DivideFloat64(float64(AvgDiskSecPerReadDiff), float64(disk.Frequency_PerfTime)), float64(AvgDiskSecPerReadBaseDiff))
				AvgDiskSecPerRead_Ms = AvgDiskSecPerRead_Us * 1000.0 // DefaultScale 3 -> DefaultScale is power of 10 -> 10 * 10 * 10 = 1000
			}

			// nolint:ineffassign
			var AvgDiskSecPerWrite_Us, AvgDiskSecPerWrite_Ms float64 = 0.0, 0.0
			if AvgDiskSecPerWriteDiff > 0 {
				AvgDiskSecPerWrite_Us = safemaths.DivideFloat64(safemaths.DivideFloat64(float64(AvgDiskSecPerWriteDiff), float64(disk.Frequency_PerfTime)), float64(AvgDiskSecPerWriteBaseDiff))
				AvgDiskSecPerWrite_Ms = AvgDiskSecPerWrite_Us * 1000.0 // DefaultScale 3 -> DefaultScale is power of 10 -> 10 * 10 * 10 = 1000
			}

			// Disk load % (PhysicalDisk\% Disk Time)
			// PERF_PRECISION_100NS_TIMER
			// AVG: N1 - N0 / D1 - D0
			// N1 - N0 / D1 - D0, where the numerator (N) represents the counter value, and the denominator (D) is the value of the private timer. The private timer has the same frequency as the 100 nanosecond timer.
			// The PercentDiskTime_Base property represents the private timer for the PercentDiskTime used in calculations FormattedData_PerfDisk from RawData_PerfDisk
			// https://stackoverflow.com/a/59228730
			PercentDiskTimeDiff := WrapDiffUint64(lastCheckResults.PercentDiskTime, disk.PercentDiskTime)               // N1 - N0
			PercentDiskTimeBaseDiff := WrapDiffUint64(lastCheckResults.PercentDiskTime_Base, disk.PercentDiskTime_Base) // D1 - D0

			loadPercentage := safemaths.DivideFloat64(float64(PercentDiskTimeDiff), float64(PercentDiskTimeBaseDiff)) * 100.0
			if loadPercentage >= 100.0 {
				// In windows disk % time can be gt 100 (yes)
				// https://docs.microsoft.com/en-us/archive/blogs/askcore/windows-performance-monitor-disk-counters-explained
				loadPercentage = 100.0
			}

			// IOPS / Sec (System\File Read Operations/sec)
			// PERF_COUNTER_COUNTER
			// AVG: (Nx - N0) / ((Dx - D0) / F)
			// (N1- N0) / ( (D1-D0) / F), where the numerator (N) represents the number of operations performed during the last sample interval,
			// the denominator (D) represents the number of ticks elapsed during the last sample interval, and F is the frequency of the ticks.

			// Read iops/sec
			DiskReadsPerSecDiff := WrapDiffUint32(lastCheckResults.DiskReadsPerSec, disk.DiskReadsPerSec)          // N1 - N0
			Timestamp_Sys100NSDiff := WrapDiffUint64(lastCheckResults.Timestamp_Sys100NS, disk.Timestamp_Sys100NS) // D1 - D0

			ReadIops := safemaths.DivideUint64(uint64(DiskReadsPerSecDiff), safemaths.DivideUint64(Timestamp_Sys100NSDiff, disk.Frequency_PerfTime))

			// Write iops/sec
			DiskWritesPerSecDiff := WrapDiffUint32(lastCheckResults.DiskWritesPerSec, disk.DiskWritesPerSec) // N1 - N0

			WriteIops := safemaths.DivideUint64(uint64(DiskWritesPerSecDiff), safemaths.DivideUint64(Timestamp_Sys100NSDiff, disk.Frequency_PerfTime))

			// Bytes/second read/write
			// PERF_COUNTER_BULK_COUNT
			// AVG: (Nx - N0) / ((Dx - D0)/ F)
			// (N1 - N0) / ( (D1 - D0) / F, where the numerator (N) represents the number of operations performed during the last sample interval,
			// the denominator (D) represent the number of ticks elapsed during the last sample interval, and the variable F is the frequency of the ticks.
			DiskReadBytesPerSecDiff := WrapDiffUint64(lastCheckResults.DiskReadBytesPerSec, disk.DiskReadBytesPerSec) // D1 - D0

			ReadBytesPerSecond := safemaths.DivideUint64(uint64(DiskReadBytesPerSecDiff), safemaths.DivideUint64(Timestamp_Sys100NSDiff, disk.Frequency_PerfTime))
			//var ReadGigabytesPerSecond float64 = float64(ReadBytesPerSecond) / 1024.0 / 1024.0 / 1024.0

			DiskWriteBytesPerSecDiff := WrapDiffUint64(lastCheckResults.DiskWriteBytesPerSec, disk.DiskWriteBytesPerSec) // D1 - D0

			WriteBytesPerSecond := safemaths.DivideUint64(uint64(DiskWriteBytesPerSecDiff), safemaths.DivideUint64(Timestamp_Sys100NSDiff, disk.Frequency_PerfTime))
			//var WriteGigabytesPerSecond float64 = float64(WriteBytesPerSecond) / 1024.0 / 1024.0 / 1024.0

			// Average request size of reads in bytes
			// PERF_AVERAGE_BULK
			// (Nx - N0) / (Dx - D0)
			// (N1 - N0) / (D1 - D0), where the numerator (N) represents the number of items processed during the last sample interval,
			// and the denominator (D) represents the number of operations completed during the last two sample intervals.
			AvgDiskBytesPerReadDiff := WrapDiffUint64(lastCheckResults.AvgDiskBytesPerRead, disk.AvgDiskBytesPerRead)               // N1 - N0
			AvgDiskBytesPerReadBaseDiff := WrapDiffUint32(lastCheckResults.AvgDiskBytesPerRead_Base, disk.AvgDiskBytesPerRead_Base) // D1 - D0

			var AvgReadRequestSizeInBytes float64 = 0.0
			if AvgDiskBytesPerReadBaseDiff > 0.0 {
				AvgReadRequestSizeInBytes = safemaths.DivideFloat64(float64(AvgDiskBytesPerReadDiff), float64(AvgDiskBytesPerReadBaseDiff))
			}

			// Writes
			AvgDiskBytesPerWriteDiff := WrapDiffUint64(lastCheckResults.AvgDiskBytesPerWrite, disk.AvgDiskBytesPerWrite)               // N1 - N0
			AvgDiskBytesPerWriteBaseDiff := WrapDiffUint64(lastCheckResults.AvgDiskBytesPerWrite_Base, disk.AvgDiskBytesPerWrite_Base) // D1 - D0

			var AvgWriteRequestSizeInBytes float64 = 0.0
			if AvgDiskBytesPerReadBaseDiff > 0.0 {
				AvgWriteRequestSizeInBytes = safemaths.DivideFloat64(float64(AvgDiskBytesPerWriteDiff), float64(AvgDiskBytesPerWriteBaseDiff))
			}

			diskstats := &resultDiskIo{
				// Store counter values for next check evaluation
				Timestamp:                 time.Now().Unix(),
				Device:                    diskName,
				Frequency_PerfTime:        disk.Frequency_PerfTime,
				Timestamp_Sys100NS:        disk.Timestamp_Sys100NS,
				AvgDiskSecPerRead:         disk.AvgDiskSecPerRead,
				AvgDiskSecPerRead_Base:    disk.AvgDiskSecPerRead_Base,
				AvgDiskSecPerWrite:        disk.AvgDiskSecPerWrite,
				AvgDiskSecPerWrite_Base:   disk.AvgDiskSecPerWrite_Base,
				PercentDiskTime:           disk.PercentDiskTime,
				PercentDiskTime_Base:      disk.PercentDiskTime_Base,
				DiskReadsPerSec:           disk.DiskReadsPerSec,
				DiskWritesPerSec:          disk.DiskWritesPerSec,
				DiskReadBytesPerSec:       disk.DiskReadBytesPerSec,
				DiskWriteBytesPerSec:      disk.DiskWriteBytesPerSec,
				AvgDiskBytesPerRead:       disk.AvgDiskBytesPerRead,
				AvgDiskBytesPerRead_Base:  disk.AvgDiskBytesPerRead_Base,
				AvgDiskBytesPerWrite:      disk.AvgDiskBytesPerWrite,
				AvgDiskBytesPerWrite_Base: disk.AvgDiskBytesPerWrite_Base,

				// Store calculated values
				ReadIopsPerSecond:   ReadIops,
				WriteIopsPerSecond:  WriteIops,
				TotalIopsPerSecond:  (ReadIops + WriteIops),
				ReadBytesPerSecond:  ReadBytesPerSecond,
				WriteBytesPerSecond: WriteBytesPerSecond,
				TotalAvgWait:        (AvgDiskSecPerRead_Ms + AvgDiskSecPerWrite_Ms),
				ReadAvgWait:         AvgDiskSecPerRead_Ms,
				WriteAvgWait:        AvgDiskSecPerWrite_Ms,
				ReadAvgSize:         AvgReadRequestSizeInBytes,
				WriteAvgSize:        AvgWriteRequestSizeInBytes,
				LoadPercent:         loadPercentage,
			}

			diskResults[diskName] = diskstats

		} else {
			//No previous check results for calculations... wait until check runs again
			diskstats := &resultDiskIo{
				// Store counter values for next check evaluation
				Timestamp: time.Now().Unix(),
				Device:    diskName,

				// Store calculated values
				Frequency_PerfTime:        disk.Frequency_PerfTime,
				Timestamp_Sys100NS:        disk.Timestamp_Sys100NS,
				AvgDiskSecPerRead:         disk.AvgDiskSecPerRead,
				AvgDiskSecPerRead_Base:    disk.AvgDiskSecPerRead_Base,
				AvgDiskSecPerWrite:        disk.AvgDiskSecPerWrite,
				AvgDiskSecPerWrite_Base:   disk.AvgDiskSecPerWrite_Base,
				PercentDiskTime:           disk.PercentDiskTime,
				PercentDiskTime_Base:      disk.PercentDiskTime_Base,
				DiskReadsPerSec:           disk.DiskReadsPerSec,
				DiskWritesPerSec:          disk.DiskWritesPerSec,
				DiskReadBytesPerSec:       disk.DiskReadBytesPerSec,
				DiskWriteBytesPerSec:      disk.DiskWriteBytesPerSec,
				AvgDiskBytesPerRead:       disk.AvgDiskBytesPerRead,
				AvgDiskBytesPerRead_Base:  disk.AvgDiskBytesPerRead_Base,
				AvgDiskBytesPerWrite:      disk.AvgDiskBytesPerWrite,
				AvgDiskBytesPerWrite_Base: disk.AvgDiskBytesPerWrite_Base,
			}

			//Store result for next check run
			diskResults[diskName] = diskstats
		}

	}

	c.lastResults = diskResults
	return diskResults, nil
}
