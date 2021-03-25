package wmiexecutor

import (
	"encoding/json"

	"github.com/StackExchange/wmi"
	"github.com/it-novum/openitcockpit-agent-go/checks"
)

// CheckDiskIo gathers information about system disks IO
type CheckDiskIo struct {
	verbose bool
	debug   bool
}

func (c *CheckDiskIo) Configure(conf *Configuration) error {
	c.verbose = conf.verbose
	c.debug = conf.debug

	return nil
}

// Query WMI
// if error != nil the check result will be nil
func (c *CheckDiskIo) RunQuery() (string, error) {

	var dst []checks.Win32_PerfRawData_PerfDisk_PhysicalDisk
	err := wmi.Query("SELECT * FROM Win32_PerfRawData_PerfDisk_PhysicalDisk WHERE Name <> '_Total'", &dst)
	if err != nil {
		return "", err
	}

	js, err := json.Marshal(dst)
	if err != nil {
		return "", err
	}

	return string(js), nil
}
