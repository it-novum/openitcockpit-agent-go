package checks

import (
	"context"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/safemaths"
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
			ReadSectors := WrapDiffUint64(lastCheckResults.ReadBytes, iostats.ReadSectors)
			WriteSectors := WrapDiffUint64(lastCheckResults.WriteBytes, iostats.WriteSectors)
			ReadCount := WrapDiffUint64(lastCheckResults.ReadCount, iostats.ReadIOs)
			WriteCount := WrapDiffUint64(lastCheckResults.WriteCount, iostats.WriteIOs)
			ReadTime := WrapDiffUint64(lastCheckResults.ReadTime, iostats.ReadTicks)
			WriteTime := WrapDiffUint64(lastCheckResults.WriteTime, iostats.WriteTicks)
			IoTime := WrapDiffUint64(lastCheckResults.IoTime, iostats.IOsTotalTicks) //BusyTime
			Interval := uint64(time.Now().Unix() - lastCheckResults.Timestamp)
			IntervalFloat := float64(Interval)

			// http://www.mjmwired.net/kernel/Documentation/block/stat.txt
			// The "sectors" in question are the standard UNIX 512-byte
			// sectors, not any device- or filesystem-specific block size.
			var BytesPerSector uint64 = 512
			ReadBytes := ReadSectors * BytesPerSector
			WriteBytes := WriteSectors * BytesPerSector

			loadPercent := safemaths.DivideFloat64(float64(IoTime), (IntervalFloat*1000.0)) * 100.0
			if loadPercent >= 100.0 {
				// Just in case this this has the same bug as Python psutil has^^
				loadPercent = 100.0
			}

			readIopsPerSecond := safemaths.DivideUint64(ReadCount, Interval)
			readBytesPerSecond := safemaths.DivideFloat64(float64(ReadBytes), IntervalFloat)
			readAvgWait := safemaths.DivideFloat64(float64(ReadTime), float64(ReadCount))
			readAvgSize := safemaths.DivideFloat64(float64(ReadBytes), float64(ReadCount))

			writeIopsPerSecond := safemaths.DivideUint64(WriteCount, Interval)
			writeBytesPerSecond := safemaths.DivideFloat64(float64(WriteBytes), IntervalFloat)
			writeAvgWait := safemaths.DivideFloat64(float64(WriteTime), float64(WriteCount))
			writeAvgSize := safemaths.DivideFloat64(float64(WriteBytes), float64(WriteCount))

			totIops := ReadCount + WriteCount

			totIopsPerSecond := safemaths.DivideFloat64(float64(totIops), IntervalFloat)
			totalAvgWait := safemaths.DivideFloat64(float64(ReadTime)+float64(WriteTime), float64(totIops))

			diskstats := &resultDiskIo{
				// Store counter values for next check evaluation
				Timestamp:  time.Now().Unix(),
				Device:     iostats.DeviceName,
				ReadBytes:  iostats.ReadSectors,
				WriteBytes: iostats.WriteSectors,
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

		} else {
			//No previous check results for calculations... wait until check runs again
			diskstats := &resultDiskIo{
				// Store counter values for next check evaluation
				Timestamp:  time.Now().Unix(),
				Device:     iostats.DeviceName,
				ReadBytes:  iostats.ReadSectors,
				WriteBytes: iostats.WriteSectors,
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
