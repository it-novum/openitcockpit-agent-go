package checks

import "time"

// CheckAgent gathers information about the agent itself
type CheckAgent struct {
	System        string // Windows e.g.: Windows Server 2016 Standard / Linux / macOS
	ReleaseId     string // Windows e.g.: 1607
	CurrentBuild  string // Windows e.g.: 14393
	Family        string // Windows e.g.: Server
	LastBootTime  time.Time
	MacVersion    string // macOS
	KernelVersion string // Linux
	CheckInterval int64  // Checkinterval of the Agent in Seconds
}

// Name will be used in the response as check name
func (c *CheckAgent) Name() string {
	return "agent"
}

type resultAgent struct {
	LastUpdated          string `json:"last_updated"`           // e.g.: 2021-01-11 15:58:35.987952 +0100 CET m=+19.945268128
	LastUpdatedTimestamp int64  `json:"last_updated_timestamp"` // w.g.: 1610377115
	System               string `json:"system"`                 // darwin | linux | Windows Server 2016 Standard
	SystemUptime         uint64 `json:"system_uptime"`          // System uptime in seconds
	KernelVersion        string `json:"kernel_version"`         // e.g.: 19.6.0
	MacVersion           string `json:"mac_version"`            // e.g.: 10.15.7
	WindowsReleaseId     string `json:"windows_release_id"`     // e.g.: 1607
	WindowsCurrentBuild  string `json:"windows_current_build"`  // e.g.: 14393
	Family               string `json:"family"`                 // Standalone Workstation | Server
	AgentVersion         string `json:"agent_version"`          // e.g. 3.0.0
	TemperatureUnit      string `json:"temperature_unit"`       // C (hardcoded)
	GOOS                 string `json:"goos"`                   // Value of runtime.GOOS
	GOARCH               string `json:"goarch"`                 // Value of runtime.ARCH
	GOVERSION            string `json:"goversion"`              // Value of runtime.Version()
	CheckInterval        int64  `json:"check_interval"`         // Check intervall in seconds of the agent
}
