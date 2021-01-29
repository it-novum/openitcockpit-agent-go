package platformpaths

import (
	"errors"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

type windowsPlatformPath struct {
	basePath       string
	registryKey    registry.Key
	registryPath   string
	registryName   string
	additionalData map[string]string
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

	keyNames, err := k.ReadValueNames(0)
	if err != nil {
		return err
	}

	p.additionalData = map[string]string{}
	for _, keyName := range keyNames {
		if s, _, err := k.GetStringValue(keyName); err != nil {
			if !errors.Is(err, registry.ErrUnexpectedType) {
				return err
			}
		} else {
			p.additionalData[keyName] = s
		}
	}

	return nil
}

func (p *windowsPlatformPath) LogPath() string {
	return filepath.Join(p.basePath, "agent.log")
}

func (p *windowsPlatformPath) ConfigPath() string {
	return p.basePath
}

func (p *windowsPlatformPath) AdditionalData() map[string]string {
	return p.additionalData
}

func getPlatformPath() PlatformPath {
	return &windowsPlatformPath{
		registryKey:  registry.LOCAL_MACHINE,
		registryPath: `SOFTWARE\it-novum\InstalledProducts\openitcockpit-agent`,
		registryName: "InstallLocation",
	}
}
