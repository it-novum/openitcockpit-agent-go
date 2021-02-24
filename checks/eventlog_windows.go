package checks

import (
	"context"
	"fmt"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/evlog"
)

type CheckWindowsEventLog struct {
	eventLog *evlog.EventLog
}

// Name will be used in the response as check name
func (c *CheckWindowsEventLog) Name() string {
	return "windows_eventlog"
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckWindowsEventLog) Run(ctx context.Context) (interface{}, error) {
	events, err := c.eventLog.Query()
	if err != nil {
		return nil, err
	}

	return events, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckWindowsEventLog) Configure(cfg *config.Configuration) (bool, error) {
	if cfg.WindowsEventLog && len(cfg.WindowsEventLogTypes) > 0 {
		avail, err := evlog.Available()
		if err != nil {
			return false, fmt.Errorf("event log availability check: %s", err)
		}
		if !avail {
			return false, fmt.Errorf("windows event log is not available")
		}

		c.eventLog = &evlog.EventLog{
			LogChannel: cfg.WindowsEventLogTypes,
		}
	}
	return false, nil
}
