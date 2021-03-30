package checks

import (
	"context"
	"runtime"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/it-novum/openitcockpit-agent-go/config"
	"golang.org/x/sys/windows/registry"
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckAgent) Run(ctx context.Context) (interface{}, error) {

	uptime := time.Since(c.LastBootTime)

	now := time.Now()
	return &resultAgent{
		LastUpdated:          now.String(),
		LastUpdatedTimestamp: now.Unix(),
		System:               c.System,
		SystemUptime:         uint64(uptime.Seconds()),
		Family:               c.Family,
		AgentVersion:         config.AgentVersion,
		TemperatureUnit:      "C",
		WindowsReleaseId:     c.ReleaseId,
		WindowsCurrentBuild:  c.CurrentBuild,
		GOOS:                 runtime.GOOS,
		GOARCH:               runtime.GOARCH,
	}, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckAgent) Configure(config *config.Configuration) (bool, error) {
	c.Init()
	return true, nil
}

func (c *CheckAgent) Init() {
	c.System, _ = readRegistryStringKey("ProductName")
	c.ReleaseId, _ = readRegistryStringKey("ReleaseId")
	c.CurrentBuild, _ = readRegistryStringKey("CurrentBuild")
	c.Family, _ = readRegistryStringKey("InstallationType")

	var dst []Win32_OperatingSystem
	err := wmi.Query("SELECT * FROM Win32_OperatingSystem", &dst)
	c.LastBootTime = time.Now()

	if err == nil {
		if len(dst) > 0 {
			var info Win32_OperatingSystem = dst[0]
			c.LastBootTime = info.LastBootUpTime
		}
	}
}

func readRegistryStringKey(keyName string) (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	value, _, err := k.GetStringValue(keyName)
	if err != nil {
		return "", err
	}

	return value, nil
}
