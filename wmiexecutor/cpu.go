package wmiexecutor

import (
	"encoding/json"

	"github.com/StackExchange/wmi"
	"github.com/it-novum/openitcockpit-agent-go/checks"
)

// CheckCpu gathers information about system CPU load
type CheckCpu struct {
	verbose bool
	debug   bool
}

func (c *CheckCpu) Configure(conf *Configuration) error {
	c.verbose = conf.verbose
	c.debug = conf.debug

	return nil
}

// Query WMI
// if error != nil the check result will be nil
func (c *CheckCpu) RunQuery() (string, error) {

	var dst []checks.Win32_PerfFormattedData_PerfOS_Processor
	err := wmi.Query("SELECT * FROM Win32_PerfFormattedData_PerfOS_Processor", &dst)
	if err != nil {
		return "", err
	}

	js, err := json.Marshal(dst)
	if err != nil {
		return "", err
	}

	return string(js), nil
}
