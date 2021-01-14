package checks

import (
	"context"
	"time"

	"github.com/prometheus/procfs/blockdevice"
	"github.com/shirou/gopsutil/v3/disk"
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
		if lastCheckResults, ok := c.lastResults[disk.Name]; ok {
			ReadCount, _ := Wrapdiff(float64(lastCheckResults.ReadCount), float64(iostats.ReadIOs))
			WriteCount, _ := Wrapdiff(float64(lastCheckResults.WriteCount), float64(iostats.WriteIOs))
			IoTime, _ := Wrapdiff(float64(lastCheckResults.IoTime), float64(iostats.IOsTotalTicks)) //BusyTime
			ReadTime, _ := Wrapdiff(float64(lastCheckResults.ReadTime), float64(iostats.ReadTicks))
			WriteTime, _ := Wrapdiff(float64(lastCheckResults.WriteTime), float64(iostats.WriteTicks))
			ReadBytes, _ := Wrapdiff(float64(lastCheckResults.ReadBytes), float64(iostats.ReadMerges))
			WriteBytes, _ := Wrapdiff(float64(lastCheckResults.WriteBytes), float64(iostats.WriteMerges))
			Timestamp, _ := Wrapdiff(float64(lastCheckResults.Timestamp), float64(time.Now().Unix()))

			loadPercent := IoTime / (Timestamp * 1000) * 100

			readAvgWait := ReadTime / ReadCount
			readAvgSize := ReadBytes / ReadCount

			writeAvgWait := WriteTime / WriteCount
			writeAvgSize := WriteBytes / WriteCount

			totIos := ReadCount + WriteCount
			totalAvgWait := (ReadTime + WriteTime) / totIos

			if loadPercent <= 101 {
				// Just in case this this has the same bug as Python psutil has^^
				diskstats := &resultDiskIo{
					Timestamp:    time.Now().Unix(),
					ReadBytes:    uint64(ReadCount),
					WriteBytes:   uint64(WriteBytes),
					ReadIops:     uint64(ReadCount),
					WriteIops:    uint64(WriteCount),
					TotalIops:    uint64(totIos),
					ReadCount:    uint64(ReadCount),
					WriteCount:   uint64(WriteCount),
					IoTime:       uint64(IoTime),
					ReadAvgWait:  readAvgWait,
					ReadTime:     uint64(ReadTime),
					ReadAvgSize:  readAvgSize,
					WriteAvgWait: writeAvgWait,
					WriteAvgSize: writeAvgSize,
					WriteTime:    uint64(WriteTime),
					TotalAvgWait: totalAvgWait,
					LoadPercent:  int64(loadPercent),
					Device:       iostats.DeviceName,
				}

				diskResults[disk.Name] = diskstats
			}

		} else {
			//No previous check results for calculations... wait until check runs again
			diskstats := &resultDiskIo{
				ReadCount:  iostats.ReadIOs,
				WriteCount: iostats.WriteIOs,
				IoTime:     iostats.IOsTotalTicks,
				ReadTime:   iostats.ReadTicks,
				WriteTime:  iostats.WriteTicks,
				ReadBytes:  iostats.ReadMerges,
				WriteBytes: iostats.WriteMerges,
				Timestamp:  time.Now().Unix(),
				Device:     iostats.DeviceName,
			}

			//Store result for next check run
			diskResults[disk.Name] = diskstats
		}

	}

	c.lastResults = diskResults
	return diskResults, nil
}
