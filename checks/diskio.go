package checks

import (
	"github.com/it-novum/openitcockpit-agent-go/config"
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

// Configure the command or return false if the command was disabled
func (c *CheckDiskIo) Configure(config *config.Configuration) (bool, error) {
	return config.DiskIo, nil
}
