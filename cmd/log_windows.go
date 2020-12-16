package cmd

import (
	"path"

	"golang.org/x/sys/windows/registry"
)

var (
	registryKey  = registry.LOCAL_MACHINE
	registryPath = `SOFTWARE\it-novum\InstalledProducts\openitcockpit-agent`
	registryName = "InstallLocation"
)

func platformLogFile() string {
	if testLogPath != "" {
		return testLogPath
	}
	k, err := registry.OpenKey(registryKey, registryPath, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer k.Close()

	s, _, err := k.GetStringValue(registryName)
	if err != nil {
		return ""
	}

	return path.Join(s, "agent.log")
}
