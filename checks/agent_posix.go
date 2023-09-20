//go:build linux || darwin
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

	now := time.Now()
	return &resultAgent{
		LastUpdated:          now.String(),
		LastUpdatedTimestamp: now.Unix(),
		System:               c.System,
		SystemUptime:         uptime,
		KernelVersion:        c.KernelVersion,
		MacVersion:           c.MacVersion,
		Family:               c.Family,
		AgentVersion:         config.AgentVersion,
		TemperatureUnit:      "C",
		GOOS:                 runtime.GOOS,
		GOARCH:               runtime.GOARCH,
		GOVERSION:            runtime.Version(),
	}, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckAgent) Configure(config *config.Configuration) (bool, error) {
	c.Init()
	return true, nil
}

func (c *CheckAgent) Init() {
	kernel, _ := host.KernelVersionWithContext(context.Background())
	platfrom, family, pver, _ := host.PlatformInformationWithContext(context.Background())

	c.System = platfrom
	c.KernelVersion = kernel
	c.MacVersion = pver
	c.Family = family
}
