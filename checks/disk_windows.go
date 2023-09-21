package checks

import (
	"context"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/it-novum/openitcockpit-agent-go/safemaths"
	"github.com/it-novum/openitcockpit-agent-go/utils"
	"github.com/prometheus-community/windows_exporter/perflib"
)

// Credit to:
// https://github.com/prometheus-community/windows_exporter/blob/9723aa221885f593ac77019566c1ced9d4d746fd/collector/logical_disk.go#L168-L187
// https://docs.microsoft.com/de-de/windows/win32/wmisdk/retrieving-raw-and-formatted-performance-data?redirectedfrom=MSDN
// https://msdn.microsoft.com/en-us/library/ms803973.aspx - LogicalDisk object reference
// nolint:underscore
type Perf_LogicalDisk struct {
	Name                   string
	AvgDiskQueueLength     float64 `perflib:"Avg. Disk Queue Length"`    // Type: QUEUELEN
	CurrentDiskQueueLength float64 `perflib:"Current Disk Queue Length"` // Type: Gauge
	DiskReadBytesPerSec    float64 `perflib:"Disk Read Bytes/sec"`       // Type: Counter
	DiskReadsPerSec        float64 `perflib:"Disk Reads/sec"`            // Type: Counter
	DiskWriteBytesPerSec   float64 `perflib:"Disk Write Bytes/sec"`      // Type: Counter
	DiskWritesPerSec       float64 `perflib:"Disk Writes/sec"`           // Type: Counter
	PercentDiskReadTime    float64 `perflib:"% Disk Read Time"`          // Type: Counter
	PercentDiskWriteTime   float64 `perflib:"% Disk Write Time"`         // Type: Counter
	PercentFreeSpace       float64 `perflib:"% Free Space_Base"`         // Type: Gauge - Total disk space in MB (yes) https://docs.microsoft.com/en-us/previous-versions/windows/embedded/ms938601(v=msdn.10)
	PercentFreeSpace_Base  float64 `perflib:"Free Megabytes"`            // Type: Gauge - Free disk space in MB
	PercentIdleTime        float64 `perflib:"% Idle Time"`               // Type: Counter
	SplitIOPerSec          float64 `perflib:"Split IO/Sec"`              // Type: Counter
	AvgDiskSecPerRead      float64 `perflib:"Avg. Disk sec/Read"`        // Type: Counter
	AvgDiskSecPerWrite     float64 `perflib:"Avg. Disk sec/Write"`       // Type: Counter
	AvgDiskSecPerTransfer  float64 `perflib:"Avg. Disk sec/Transfer"`    // Type: Counter
	DiskTime               float64 `perflib:"% Disk Time"`
	DiskTime_Base          float64 `perflib:"% Disk Time_Base"`
	IdleTime               float64 `perflib:"% Idle Time"`
}

var (
	perflibDiskQuery = strconv.FormatUint(uint64(perflib.QueryNameTable("Counter 009").LookupIndex("LogicalDisk")), 10)
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckDisk) Run(_ context.Context) (interface{}, error) {
	objects, err := perflib.QueryPerformanceData(perflibDiskQuery)
	diskResults := make([]*resultDisk, 0)
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
			log.Errorln("Check Disk: could not query perflib: ", err)
			continue
		}

		for _, disk := range dst {
			if disk.Name != "_Total" {
				//Do the math
				totalDiskSpaceBytes := disk.PercentFreeSpace * 1024 * 1024
				freeDiskSpaceBytes := disk.PercentFreeSpace_Base * 1024 * 1024
				usedDiskSpaceBytes := totalDiskSpaceBytes - freeDiskSpaceBytes

				freeDiskSpacePercentage := safemaths.DivideFloat64(freeDiskSpaceBytes, totalDiskSpaceBytes) * 100.0
				usedDiskSpacePercentage := 100.0 - freeDiskSpacePercentage

				//Save to struct
				result := &resultDisk{}

				result.Disk.Device = disk.Name
				result.Disk.Mountpoint = disk.Name

				result.Usage.Total = uint64(totalDiskSpaceBytes)
				result.Usage.Used = uint64(usedDiskSpaceBytes)
				result.Usage.Free = uint64(freeDiskSpaceBytes)
				result.Usage.Percent = usedDiskSpacePercentage

				diskResults = append(diskResults, result)
			}
		}
	}

	return diskResults, nil
}
