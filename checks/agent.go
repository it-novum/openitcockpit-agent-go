package checks

import (
	"context"
	"runtime"
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
	LastUpdated          string `json:"last_updated"`           // e.g.: 2021-01-11 15:58:35.987952 +0100 CET m=+19.945268128
	LastUpdatedTimestamp int64  `json:"last_updated_timestamp"` // w.g.: 1610377115
	System               string `json:"system"`                 // darwin | linux | windows
	SystemUptime         uint64 `json:"system_uptime"`          // System uptime in seconds
	KernelVersion        string `json:"kernel_version"`         // e.g.: 19.6.0
	MacVersion           string `json:"mac_version"`            // e.g.: 10.15.7
	Family               string `json:"family"`                 // Standalone Workstation | Server
	AgentVersion         string `json:"agent_version"`          // e.g. 3.0.0
	TemperatureUnit      string `json:"temperature_unit"`       // C (hardcoded)
	GOOS                 string `json:"goos"`                   // wValue of runtime.GOOS
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
