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
			ReadBytes, _ := Wrapdiff(float64(lastCheckResults.ReadBytes), float64(iostats.ReadBytes))    // Number of bytes read from disk (Counter)
			WriteBytes, _ := Wrapdiff(float64(lastCheckResults.WriteBytes), float64(iostats.WriteBytes)) // Number of bytes written to disk (Counter)
			ReadCount, _ := Wrapdiff(float64(lastCheckResults.ReadCount), float64(iostats.ReadCount))    // Number of read iops (Counter)
			WriteCount, _ := Wrapdiff(float64(lastCheckResults.WriteCount), float64(iostats.WriteCount)) // Number of write iops (Counter)
			ReadTime, _ := Wrapdiff(float64(lastCheckResults.ReadTime), float64(iostats.ReadTime))       // Time spent reading from disk (in milliseconds)
			WriteTime, _ := Wrapdiff(float64(lastCheckResults.WriteTime), float64(iostats.WriteTime))    // Time spent writing to disk (in milliseconds)
			IoTime, _ := Wrapdiff(float64(lastCheckResults.IoTime), float64(iostats.IoTime))             // Time spent doing actual I/Os (in milliseconds)
			Interval, _ := Wrapdiff(float64(lastCheckResults.Timestamp), float64(time.Now().Unix()))     // Time between current and last check (in seconds)

			loadPercent := IoTime / (Interval * 1000.0) * 100.0

			readIopsPerSecond := ReadCount / Interval
			readBytesPerSecond := ReadBytes / Interval
			readAvgWait := ReadTime / ReadCount
			readAvgSize := ReadBytes / ReadCount

			writeIopsPerSecond := WriteCount / Interval
			writeBytesPerSecond := WriteBytes / Interval
			writeAvgWait := WriteTime / WriteCount
			writeAvgSize := WriteBytes / WriteCount

			totIops := ReadCount + WriteCount
			totIopsPerSecond := totIops / Interval
			totalAvgWait := (ReadTime + WriteTime) / totIops

			if loadPercent <= 101 {
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
