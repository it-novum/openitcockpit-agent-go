// +build nobuild

package checks

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/utils"
	"github.com/leoluk/perflib_exporter/perflib"
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckDiskIo) Run(ctx context.Context) (interface{}, error) {
	//Todo can we cache this?
	nametable := perflib.QueryNameTable("Counter 009")
	query := strconv.FormatUint(uint64(nametable.LookupIndex("LogicalDisk")), 10)

	objects, err := perflib.QueryPerformanceData(query)
	diskResults := make(map[string]*resultDiskIo)
	if err != nil {
		return nil, err
	}

	for _, obj := range objects {
		if obj.Name != "LogicalDisk" {
			continue
		}

		var dst []Perf_LogicalDisk
		err = utils.UnmarshalObject(obj, &dst)
		if err != nil {
			// todo add logging
			fmt.Println(err)
			continue
		}

		for _, disk := range dst {
			// https://docs.microsoft.com/en-us/archive/blogs/askcore/windows-performance-monitor-disk-counters-explained
			//fmt.Println(disk)
			if lastCheckResults, ok := c.lastResults[disk.Name]; ok {
				ReadBytesPerSecond, _ := Wrapdiff(float64(lastCheckResults.ReadBytes), float64(disk.DiskReadBytesPerSec))
				WriteBytesPerSecond, _ := Wrapdiff(float64(lastCheckResults.WriteBytes), float64(disk.DiskWriteBytesPerSec))
				DiskReadsPerSeccond, _ := Wrapdiff(float64(lastCheckResults.ReadCount), float64(disk.DiskReadsPerSec))
				DiskWritesPerSec, _ := Wrapdiff(float64(lastCheckResults.WriteCount), float64(disk.DiskWritesPerSec))
				PercentDiskReadTime, _ := Wrapdiff(float64(lastCheckResults.ReadTime), float64(disk.PercentDiskReadTime))
				PercentDiskWriteTime, _ := Wrapdiff(float64(lastCheckResults.WriteTime), float64(disk.PercentDiskWriteTime))
				tmpIoTime := uint64(disk.PercentDiskReadTime) + uint64(disk.PercentDiskWriteTime)
				IoTime, _ := Wrapdiff(float64(lastCheckResults.IoTime), float64(tmpIoTime))

				//AvgDiskQueueLength, _ := Wrapdiff(float64(lastCheckResults.AvgDiskQueueLength), float64(disk.AvgDiskQueueLength))

				DiskTime_Base, _ := Wrapdiff(float64(lastCheckResults.DiskTime_Base), float64(disk.DiskTime_Base))
				DiskTime, _ := Wrapdiff(float64(lastCheckResults.DiskTime), float64(disk.DiskTime))
				//AvgDiskSecPerRead, _ := Wrapdiff(float64(lastCheckResults.AvgDiskSecPerRead), float64(disk.AvgDiskSecPerRead))
				//AvgDiskSecPerWrite, _ := Wrapdiff(float64(lastCheckResults.AvgDiskSecPerWrite), float64(disk.AvgDiskSecPerWrite))
				AvgDiskSecPerTransfer, _ := Wrapdiff(float64(lastCheckResults.AvgDiskSecPerTransfer), float64(disk.AvgDiskSecPerTransfer))
				//IdleTime, _ := Wrapdiff(float64(lastCheckResults.IdleTime), disk.IdleTime)

				Interval, _ := Wrapdiff(float64(lastCheckResults.Timestamp), float64(time.Now().Unix())) // Time between current and last check (in seconds)

				loadPercent := DiskTime_Base / DiskTime * 100.0

				//loadPercent := AvgDiskQueueLength //DiskTime / Interval * 100.0
				//fmt.Printf("\n***\nLoad of %v is percentage: %v\n", disk.Name, loadPercent)

				readIopsPerSecond := DiskReadsPerSeccond / Interval
				readBytesPerSecond := ReadBytesPerSecond / Interval
				readAvgWait := PercentDiskReadTime / DiskReadsPerSeccond
				readAvgSize := ReadBytesPerSecond / DiskReadsPerSeccond

				writeIopsPerSecond := DiskWritesPerSec / Interval
				writeBytesPerSecond := WriteBytesPerSecond / Interval
				writeAvgWait := PercentDiskWriteTime / DiskWritesPerSec
				writeAvgSize := WriteBytesPerSecond / DiskWritesPerSec

				totIops := DiskReadsPerSeccond + DiskWritesPerSec
				totIopsPerSecond := totIops / Interval
				totalAvgWait := (PercentDiskReadTime + PercentDiskWriteTime) / totIops

				fmt.Printf("\n***\n Total IO WAIT for %v is: %v ms\n", disk.Name, AvgDiskSecPerTransfer*utils.WINDOWS_TICKS_PER_SECONDS)
				//fmt.Printf("\n***\nRead IO WAIT for %v is: %v ms\n", disk.Name, disk.AvgDiskSecPerRead*utils.WINDOWS_TICKS_PER_SECONDS)
				//fmt.Printf("\n***\nWrite IO WAIT for %v is: %v ms\n", disk.Name, disk.AvgDiskSecPerWrite*utils.WINDOWS_TICKS_PER_SECONDS)

				if loadPercent <= 101 {
					// Just in case this this has the same bug as Python psutil has^^
					diskstats := &resultDiskIo{
						// Store counter values for next check evaluation
						Timestamp:             time.Now().Unix(),
						Device:                disk.Name,
						ReadBytes:             uint64(disk.DiskReadBytesPerSec),
						WriteBytes:            uint64(disk.DiskWriteBytesPerSec),
						ReadCount:             uint64(disk.DiskReadsPerSec),
						WriteCount:            uint64(disk.DiskWritesPerSec),
						ReadTime:              uint64(disk.PercentDiskReadTime),
						WriteTime:             uint64(disk.PercentDiskWriteTime),
						IoTime:                uint64(IoTime),
						AvgDiskQueueLength:    disk.AvgDiskQueueLength,
						IdleTime:              disk.IdleTime,
						AvgDiskSecPerRead:     disk.AvgDiskSecPerRead,
						AvgDiskSecPerWrite:    disk.AvgDiskSecPerWrite,
						AvgDiskSecPerTransfer: disk.AvgDiskSecPerTransfer,

						// Store calculated values
						ReadIopsPerSecond:   uint64(readIopsPerSecond),
						WriteIopsPerSecond:  uint64(writeIopsPerSecond),
						TotalIopsPerSecond:  uint64(totIopsPerSecond),
						ReadBytesPerSecond:  uint64(readBytesPerSecond),
						WriteBytesPerSecond: uint64(writeBytesPerSecond),
						TotalAvgWait:        totalAvgWait,
						ReadAvgWait:         readAvgWait,
						WriteAvgWait:        writeAvgWait,
						ReadAvgSize:         readAvgSize,
						WriteAvgSize:        writeAvgSize,
						LoadPercent:         loadPercent,
					}

					diskResults[disk.Name] = diskstats
				}

			} else {
				//No previous check results for calculations... wait until check runs again
				diskstats := &resultDiskIo{
					// Store counter values for next check evaluation
					Timestamp:             time.Now().Unix(),
					Device:                disk.Name,
					ReadBytes:             uint64(disk.DiskReadBytesPerSec),
					WriteBytes:            uint64(disk.DiskWriteBytesPerSec),
					ReadCount:             uint64(disk.DiskReadsPerSec),
					WriteCount:            uint64(disk.DiskWritesPerSec),
					ReadTime:              uint64(disk.PercentDiskReadTime),
					WriteTime:             uint64(disk.PercentDiskWriteTime),
					IoTime:                uint64(disk.PercentDiskReadTime) + uint64(disk.PercentDiskWriteTime),
					AvgDiskQueueLength:    disk.AvgDiskQueueLength,
					AvgDiskSecPerRead:     disk.AvgDiskSecPerRead,
					AvgDiskSecPerWrite:    disk.AvgDiskSecPerWrite,
					AvgDiskSecPerTransfer: disk.AvgDiskSecPerTransfer,
				}

				//Store result for next check run
				diskResults[disk.Name] = diskstats
			}
		}
	}

	c.lastResults = diskResults
	return diskResults, nil
}
