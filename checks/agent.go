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
func (c *CheckAgent) Run(ctx context.Context) (*CheckResult, error) {
	uptime, err := host.UptimeWithContext(ctx)
	if err != nil {
		uptime = 0
	}

	kernel, _ := host.KernelVersionWithContext(ctx)
	platfrom, family, pver, _ := host.PlatformInformationWithContext(ctx)

	now := time.Now()
	return &CheckResult{
		Result: &resultAgent{
			LastUpdated:          now.String(),
			LastUpdatedTimestamp: now.Unix(),
			System:               platfrom,
			SystemUptime:         uptime,
			KernelVersion:        kernel,
			MacVersion:           pver,
			Family:               family,
			AgentVersion:         config.AgentVersion,
			TemperatureUnit:      "C",
		},
	}, nil
}

// DefaultConfiguration contains the variables for the configuration file and the default values
// can be nil if no configuration is required
func (c *CheckAgent) DefaultConfiguration() interface{} {
	return nil
}

// Configure should verify the configuration and set it
// will be run after every reload
// if DefaultConfiguration returns nil, the parameter will also be nil
func (c *CheckAgent) Configure(_ interface{}) error {
	return nil
}