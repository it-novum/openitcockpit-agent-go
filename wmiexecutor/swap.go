package wmiexecutor

import (
	"encoding/json"

	"github.com/StackExchange/wmi"
	"github.com/it-novum/openitcockpit-agent-go/checks"
)

// CheckSwap gathers information about system swap
type CheckSwap struct {
	verbose bool
	debug   bool
}

func (c *CheckSwap) Configure(conf *Configuration) error {
	c.verbose = conf.verbose
	c.debug = conf.debug

	return nil
}

// Query WMI
// if error != nil the check result will be nil
func (c *CheckSwap) RunQuery() (string, error) {

	var dst []checks.Win32_PageFileUsage
	err := wmi.Query("SELECT * FROM Win32_PageFileUsage", &dst)
	if err != nil {
		return "", err
	}

	js, err := json.Marshal(dst)
	if err != nil {
		return "", err
	}

	return string(js), nil
}
