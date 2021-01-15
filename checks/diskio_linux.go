package checks

import (
	"context"
	"time"

	"github.com/prometheus/procfs/blockdevice"
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckDiskIo) Run(ctx context.Context) (interface{}, error) {
	// https://www.kernel.org/doc/html/latest/admin-guide/abi-testing.html#symbols-under-proc-diskstats
	// https://github.com/giampaolo/psutil/blob/f18438d135c12f7eb186f49622e0f6683c37f7f5/psutil/_pslinux.py#L1093

	fs, err := blockdevice.NewFS("/proc", "/sys")
	if err != nil {
		return nil, err
	}

	stats, err := fs.ProcDiskstats()

	if err != nil {
		return nil, err
	}

	diskResults := make(map[string]*resultDiskIo)

	for _, iostats := range stats {
		if lastCheckResults, ok := c.lastResults[iostats.DeviceName]; ok {
			ReadSectors, _ := Wrapdiff(float64(lastCheckResults.ReadBytes), float64(iostats.ReadSectors))
			WriteSectors, _ := Wrapdiff(float64(lastCheckResults.WriteBytes), float64(iostats.WriteSectors))
			ReadCount, _ := Wrapdiff(float64(lastCheckResults.ReadCount), float64(iostats.ReadIOs))
			WriteCount, _ := Wrapdiff(float64(lastCheckResults.WriteCount), float64(iostats.WriteIOs))
			ReadTime, _ := Wrapdiff(float64(lastCheckResults.ReadTime), float64(iostats.ReadTicks))
			WriteTime, _ := Wrapdiff(float64(lastCheckResults.WriteTime), float64(iostats.WriteTicks))
			IoTime, _ := Wrapdiff(float64(lastCheckResults.IoTime), float64(iostats.IOsTotalTicks)) //BusyTime
			Interval, _ := Wrapdiff(float64(lastCheckResults.Timestamp), float64(time.Now().Unix()))

			// http://www.mjmwired.net/kernel/Documentation/block/stat.txt
			// The "sectors" in question are the standard UNIX 512-byte
			// sectors, not any device- or filesystem-specific block size.
			BytesPerSector := 512.0
			ReadBytes := ReadSectors * BytesPerSector
			WriteBytes := WriteSectors * BytesPerSector

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
					Device:     iostats.DeviceName,
					ReadBytes:  iostats.ReadMerges,
					WriteBytes: iostats.WriteMerges,
					ReadCount:  iostats.ReadIOs,
					WriteCount: iostats.WriteIOs,
					ReadTime:   iostats.ReadTicks,
					WriteTime:  iostats.WriteTicks,
					IoTime:     iostats.IOsTotalTicks,

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

				diskResults[iostats.DeviceName] = diskstats
			}

		} else {
			//No previous check results for calculations... wait until check runs again
			diskstats := &resultDiskIo{
				// Store counter values for next check evaluation
				Timestamp:  time.Now().Unix(),
				Device:     iostats.DeviceName,
				ReadBytes:  iostats.ReadMerges,
				WriteBytes: iostats.WriteMerges,
				ReadCount:  iostats.ReadIOs,
				WriteCount: iostats.WriteIOs,
				ReadTime:   iostats.ReadTicks,
				WriteTime:  iostats.WriteTicks,
				IoTime:     iostats.IOsTotalTicks,
			}

			//Store result for next check run
			diskResults[iostats.DeviceName] = diskstats
		}

	}

	c.lastResults = diskResults
	return diskResults, nil
}
