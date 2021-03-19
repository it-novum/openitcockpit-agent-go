// +build linux darwin

package checks

import (
	"context"
	"runtime"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/shirou/gopsutil/v3/host"
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckAgent) Run(ctx context.Context) (interface{}, error) {
	uptime, err := host.UptimeWithContext(ctx)
	if err != nil {
		uptime = 0
	}

	kernel, _ := host.KernelVersionWithContext(ctx)
	platfrom, family, pver, _ := host.PlatformInformationWithContext(ctx)

	now := time.Now()
	return &resultAgent{
		LastUpdated:          now.String(),
		LastUpdatedTimestamp: now.Unix(),
		System:               platfrom,
		SystemUptime:         uptime,
		KernelVersion:        kernel,
		MacVersion:           pver,
		Family:               family,
		AgentVersion:         config.AgentVersion,
		TemperatureUnit:      "C",
		GOOS:                 runtime.GOOS,
	}, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckAgent) Configure(config *config.Configuration) (bool, error) {
	return true, nil
}
