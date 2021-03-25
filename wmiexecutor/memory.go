package wmiexecutor

import (
	"encoding/json"

	"github.com/StackExchange/wmi"
	"github.com/it-novum/openitcockpit-agent-go/checks"
)

// CheckMem gathers information about system memory
type CheckMem struct {
	verbose bool
	debug   bool
}

func (c *CheckMem) Configure(conf *Configuration) error {
	c.verbose = conf.verbose
	c.debug = conf.debug

	return nil
}

// Query WMI
// if error != nil the check result will be nil
func (c *CheckMem) RunQuery() (string, error) {

	var dst []checks.Win32_OperatingSystem
	err := wmi.Query("SELECT * FROM Win32_OperatingSystem", &dst)
	if err != nil {
		return "", err
	}

	js, err := json.Marshal(dst)
	if err != nil {
		return "", err
	}

	return string(js), nil
}
