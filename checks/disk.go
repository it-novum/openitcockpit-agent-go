package checks

import (
	"github.com/it-novum/openitcockpit-agent-go/config"
)

// CheckDisk gathers information about system disks
type CheckDisk struct {
}

// Name will be used in the response as check name
func (c *CheckDisk) Name() string {
	return "disks"
}

type resultDisk struct {
	Disk struct {
		Device     string   `json:"device"`     // e.g.: /dev/disk1s5,/dev/sda
		Mountpoint string   `json:"mountpoint"` // e.g.: /
		Fstype     string   `json:"fstype"`     // e.g.: apfs, etx4 (macOS and Linux only)
		Opts       []string `json:"opts"`       // e.g.: ["rw","nobrowse","multilabel"] (macOS and Linux only)
	} `json:"disk"`
	Usage struct {
		Total   uint64  `json:"total"`   // Total disk space in byte
		Used    uint64  `json:"used"`    // Used disk space in byte
		Free    uint64  `json:"free"`    // Free disk space in byte
		Percent float64 `json:"percent"` // Used disk spaces as percent
	} `json:"usage"`
}

// Configure the command or return false if the command was disabled
func (c *CheckDisk) Configure(config *config.Configuration) (bool, error) {
	return config.Diskstats, nil
}
