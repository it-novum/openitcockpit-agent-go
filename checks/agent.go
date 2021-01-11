package checks

import (
	"context"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/shirou/gopsutil/v3/host"
)

// CheckAgent gathers information about the agent itself
type CheckAgent struct {
}

// Name will be used in the response as check name
func (c *CheckAgent) Name() string {
	return "agent"
}

type resultAgent struct {
	LastUpdated          string `json:"last_updated"`
	LastUpdatedTimestamp int64  `json:"last_updated_timestamp"`
	System               string `json:"system"`
	SystemUptime         uint64 `json:"system_uptime"`
	KernelVersion        string `json:"kernel_version"`
	MacVersion           string `json:"mac_version"`
	Family               string `json:"family"`
	AgentVersion         string `json:"agent_version"`
	TemperatureUnit      string `json:"temperature_unit"`
}

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
		LastUpdated:          now.String(),        // e.g.: 2021-01-11 15:58:35.987952 +0100 CET m=+19.945268128
		LastUpdatedTimestamp: now.Unix(),          // w.g.: 1610377115
		System:               platfrom,            // darwin | linux | windows
		SystemUptime:         uptime,              // System uptime in seconds
		KernelVersion:        kernel,              // e.g.: 19.6.0
		MacVersion:           pver,                // e.g.: 10.15.7
		Family:               family,              // Standalone Workstation | Server
		AgentVersion:         config.AgentVersion, // e.g. 3.0.0
		TemperatureUnit:      "C",
	}, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckAgent) Configure(config *config.Configuration) (bool, error) {
	return true, nil
}
