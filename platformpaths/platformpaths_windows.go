package platformpaths

import (
	"path"

	"golang.org/x/sys/windows/registry"
)

type windowsPlatformPath struct {
	basePath     string
	registryKey  registry.Key
	registryPath string
	registryName string
}

func (p *windowsPlatformPath) Init() error {
	k, err := registry.OpenKey(p.registryKey, p.registryPath, registry.QUERY_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	s, _, err := k.GetStringValue(p.registryName)
	if err != nil {
		return err
	}
	p.basePath = s
	return nil
}

func (p *windowsPlatformPath) LogPath() string {
	return path.Join(p.basePath, "agent.log")
}

func (p *windowsPlatformPath) ConfigPath() string {
	return p.basePath
}

func getPlatformPath() PlatformPath {
	return &windowsPlatformPath{
		registryKey:  registry.LOCAL_MACHINE,
		registryPath: `SOFTWARE\it-novum\InstalledProducts\openitcockpit-agent`,
		registryName: "InstallLocation",
	}
}
