package checks

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckDiskIo) Run(ctx context.Context) (interface{}, error) {

	disks, err := disk.IOCountersWithContext(ctx)
	if err != nil {
		return nil, err
	}
	diskResults := make(map[string]*resultDiskIo)

	for device, iostats := range disks {
		if lastCheckResults, ok := c.lastResults[iostats.Name]; ok {
			//All values are counters so we need to diff the current value with the last value
			ReadBytes := WrapDiffUint64(lastCheckResults.ReadBytes, iostats.ReadBytes)    // Number of bytes read from disk (Counter)
			WriteBytes := WrapDiffUint64(lastCheckResults.WriteBytes, iostats.WriteBytes) // Number of bytes written to disk (Counter)
			ReadCount := WrapDiffUint64(lastCheckResults.ReadCount, iostats.ReadCount)    // Number of read iops (Counter)
			WriteCount := WrapDiffUint64(lastCheckResults.WriteCount, iostats.WriteCount) // Number of write iops (Counter)
			ReadTime := WrapDiffUint64(lastCheckResults.ReadTime, iostats.ReadTime)       // Time spent reading from disk (in milliseconds)
			WriteTime := WrapDiffUint64(lastCheckResults.WriteTime, iostats.WriteTime)    // Time spent writing to disk (in milliseconds)
			IoTime := WrapDiffUint64(lastCheckResults.IoTime, iostats.IoTime)             // Time spent doing actual I/Os (in milliseconds)
			Interval := uint64(time.Now().Unix() - lastCheckResults.Timestamp)            // Time between current and last check (in seconds)

			loadPercent := float64(IoTime) / (float64(Interval) * 1000.0) * 100.0

			readIopsPerSecond := ReadCount / Interval
			readBytesPerSecond := float64(ReadBytes) / float64(Interval)
			readAvgWait := float64(ReadTime) / float64(ReadCount)
			readAvgSize := float64(ReadBytes) / float64(ReadCount)

			writeIopsPerSecond := WriteCount / Interval
			writeBytesPerSecond := float64(WriteBytes) / float64(Interval)
			writeAvgWait := float64(WriteTime) / float64(WriteCount)
			writeAvgSize := float64(WriteBytes) / float64(WriteCount)

			totIops := ReadCount + WriteCount

			totIopsPerSecond := float64(totIops) / float64(Interval)
			totalAvgWait := (float64(ReadTime) + float64(WriteTime)) / float64(totIops)

			if loadPercent <= 100.0 {
				// Just in case this this has the same bug as Python psutil has^^
				diskstats := &resultDiskIo{
					// Store counter values for next check evaluation
					Timestamp:  time.Now().Unix(),
					Device:     device,
					ReadBytes:  iostats.ReadBytes,
					WriteBytes: iostats.WriteBytes,
					ReadCount:  iostats.ReadCount,
					WriteCount: iostats.WriteCount,
					ReadTime:   iostats.ReadTime,
					WriteTime:  iostats.WriteTime,
					IoTime:     iostats.IoTime,

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

				diskResults[iostats.Name] = diskstats
			}

		} else {
			//No previous check results for calculations... wait until check runs again
			diskstats := &resultDiskIo{
				// Store counter values for next check evaluation
				Timestamp:  time.Now().Unix(),
				Device:     device,
				ReadBytes:  iostats.ReadBytes,
				WriteBytes: iostats.WriteBytes,
				ReadCount:  iostats.ReadCount,
				WriteCount: iostats.WriteCount,
				ReadTime:   iostats.ReadTime,
				WriteTime:  iostats.WriteTime,
				IoTime:     iostats.IoTime,
			}

			//Store result for next check run
			diskResults[iostats.Name] = diskstats
		}

	}

	c.lastResults = diskResults
	return diskResults, nil
}
