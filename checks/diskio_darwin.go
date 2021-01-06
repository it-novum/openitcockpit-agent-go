package checks

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/shirou/gopsutil/v3/disk"
)

// CheckDiskIo gathers information about system disks IO
type CheckDiskIo struct {
	lastResults []*resultDiskIo
}

// Name will be used in the response as check name
func (c *CheckDiskIo) Name() string {
	return "disk_io"
}

type resultDiskIo struct {
	ReadBytes    uint64
	WriteBytes   uint64
	ReadIops     uint64
	WriteIops    uint64
	TotalIops    uint64
	ReadCount    uint64
	WriteCount   uint64
	IoTime       uint64
	ReadAvgWait  float64
	ReadTime     uint64
	ReadAvgSize  float64
	WriteAvgWait float64
	WriteAvgSize float64
	WriteTime    uint64
	TotalAvgWait float64
	LoadPercent  int64
	Timestamp    int64
	Device       string
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckDiskIo) Run(ctx context.Context) (*CheckResult, error) {

	disks, err := disk.IOCountersWithContext(ctx)
	if err != nil {
		return nil, err
	}
	diskResults := make([]*resultDiskIo, 0, len(disks))

	for device, iostats := range disks {

		var lastCheckResults *resultDiskIo
		for _, lastCheckResultsLoop := range c.lastResults {
			if lastCheckResultsLoop.Device == device {
				lastCheckResults = lastCheckResultsLoop
				break
			}
		}

		if lastCheckResults != nil {
			ReadCount, _ := c.Wrapdiff(float64(lastCheckResults.ReadCount), float64(iostats.ReadCount))
			WriteCount, _ := c.Wrapdiff(float64(lastCheckResults.WriteCount), float64(iostats.WriteCount))
			IoTime, _ := c.Wrapdiff(float64(lastCheckResults.IoTime), float64(iostats.IoTime))
			ReadTime, _ := c.Wrapdiff(float64(lastCheckResults.ReadTime), float64(iostats.ReadTime))
			WriteTime, _ := c.Wrapdiff(float64(lastCheckResults.WriteTime), float64(iostats.WriteTime))
			ReadBytes, _ := c.Wrapdiff(float64(lastCheckResults.ReadBytes), float64(iostats.ReadBytes))
			WriteBytes, _ := c.Wrapdiff(float64(lastCheckResults.WriteBytes), float64(iostats.WriteBytes))
			Timestamp, _ := c.Wrapdiff(float64(lastCheckResults.Timestamp), float64(time.Now().Unix()))

			load_percent := IoTime / (Timestamp * 1000) * 100

			read_avg_wait := ReadTime / ReadCount
			read_avg_size := ReadBytes / ReadCount

			write_avg_wait := WriteTime / WriteCount
			write_avg_size := WriteBytes / WriteCount

			tot_ios := ReadCount + WriteCount
			total_avg_wait := (ReadTime + WriteTime) / tot_ios

			if load_percent <= 101 {
				// Just in case this this has the same bug as Python psutil has^^
				diskstats := &resultDiskIo{
					Timestamp:    time.Now().Unix(),
					ReadBytes:    uint64(ReadCount),
					WriteBytes:   uint64(WriteBytes),
					ReadIops:     uint64(ReadCount),
					WriteIops:    uint64(WriteCount),
					TotalIops:    uint64(tot_ios),
					ReadCount:    uint64(ReadCount),
					WriteCount:   uint64(WriteCount),
					IoTime:       uint64(IoTime),
					ReadAvgWait:  read_avg_wait,
					ReadTime:     uint64(ReadTime),
					ReadAvgSize:  read_avg_size,
					WriteAvgWait: write_avg_wait,
					WriteAvgSize: write_avg_size,
					WriteTime:    uint64(WriteTime),
					TotalAvgWait: total_avg_wait,
					LoadPercent:  int64(load_percent),
					Device:       device,
				}

				diskResults = append(diskResults, diskstats)
			}

		} else {
			//No previous check results for calculations... wait until check runs again
			diskstats := &resultDiskIo{
				ReadCount:  iostats.ReadCount,
				WriteCount: iostats.WriteCount,
				IoTime:     iostats.IoTime,
				ReadTime:   iostats.ReadTime,
				WriteTime:  iostats.WriteTime,
				ReadBytes:  iostats.ReadBytes,
				WriteBytes: iostats.WriteBytes,
				Timestamp:  time.Now().Unix(),
				Device:     device,
			}

			//Store result for next check run
			diskResults = append(diskResults, diskstats)
		}

	}

	c.lastResults = diskResults
	return &CheckResult{Result: diskResults}, nil
}

//Wrapdiff calculate the difference between last and curr
//If last > curr, try to guess the boundary at which the value must have wrapped
//by trying the maximum values of 64, 32 and 16 bit signed and unsigned ints.
func (c *CheckDiskIo) Wrapdiff(last, curr float64) (float64, error) {
	if last <= curr {
		return curr - last, nil
	}

	boundaries := []float64{64, 63, 32, 31, 16, 15}
	var currBoundary float64
	for _, boundary := range boundaries {
		if last > math.Pow(2, boundary) {
			currBoundary = boundary
		}
	}

	if currBoundary == 0 {
		return 0, fmt.Errorf("Couldn't determine boundary")
	}

	return math.Pow(2, currBoundary) - last + curr, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckDiskIo) Configure(config *config.Configuration) (bool, error) {
	return config.DiskIo, nil
}
