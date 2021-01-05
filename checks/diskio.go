package checks

/*
import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

// CheckDiskIo gathers information about system disks IO
type CheckDiskIo struct {
	lastResult []*resultDiskIo
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

	diffKeys := []string{
		"ReadCount",
		"WriteCount",
		"IoTime",
		"ReadTime",
		"WriteTime",
		"ReadBytes",
		"WriteBytes",
		"Timestamp",
	}

	for device, iostats := range disks {

		var oldResults *resultDiskIo
		for _, oldResult := range c.lastResult {
			if oldResult.Device == device {
				oldResults = oldResult
				break
			}
		}

		diskstats := &resultDiskIo{
			Timestamp: time.Now().Unix(),
		}

		// Ich versuche hier eigentlich erstmal an die alten check ergebnisse ran zu kommen, um die dann in das Wrappdiff zu packen.
		// Wenn es keine alten checkergebnisse gibt, dann soll er die aktuellen abspeichern und muss das dann beim nächsten check machen. also für eine disk einfach gerade keine diskio
		// werte zurückgeben beim ersten aufruf.
		// in php wäre das praktisch ein !isset($oldChecks[$device]), aber da bin ich gerade zu Braindead für um das in go zu bauen.
		if oldResults != nil {
			//Wenn er die alten Werte zum vergleichen hat, kann er alles rechnen und returnen
			ReadCount, _ := c.Wrapdiff(float64(oldResults.ReadCount), float64(iostats.ReadCount))
			WriteCount, _ := c.Wrapdiff(float64(oldResults.WriteCount), float64(iostats.WriteCount))
			IoTime, _ := c.Wrapdiff(float64(oldResults.IoTime), float64(iostats.IoTime))
			ReadTime, _ := c.Wrapdiff(float64(oldResults.ReadTime), float64(iostats.ReadTime))
			WriteTime, _ := c.Wrapdiff(float64(oldResults.WriteTime), float64(iostats.WriteTime))
			ReadBytes, _ := c.Wrapdiff(float64(oldResults.ReadBytes), float64(iostats.ReadBytes))
			WriteBytes, _ := c.Wrapdiff(float64(oldResults.WriteBytes), float64(iostats.WriteBytes))
			Timestamp, _ := c.Wrapdiff(float64(oldResults.Timestamp), float64(diskstats.Timestamp))

			/*
				load_percent = IoTime / (timestamp*1000) * 100

				read_avg_wait := read_time / read_count
				read_avg_size := read_bytes / read_count

				write_avg_wait := write_time / write_count
				write_avg_size := write_bytes / write_count

				tot_ios := diskIODiff['read_count'] + diskIODiff['write_count']
				total_avg_wait := read_time + write_time / tot_ios*/
/*
			fmt.Println(device)
			fmt.Println(iostats)
		}

		diskResults = append(diskResults, diskstats)
	}

	c.lastResult = diskResults
	return &CheckResult{Result: diskResults}, nil
}

// DefaultConfiguration contains the variables for the configuration file and the default values
// can be nil if no configuration is required
func (c *CheckDiskIo) DefaultConfiguration() interface{} {
	return nil
}

// Configure should verify the configuration and set it
// will be run after every reload
// if DefaultConfiguration returns nil, the parameter will also be nil
func (c *CheckDiskIo) Configure(_ interface{}) error {
	return nil
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
*/
