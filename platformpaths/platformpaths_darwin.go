package platformpaths

import "path/filepath"

type darwinPlatformPath struct {
	basePath string
}

func (p *darwinPlatformPath) Init() error {
	p.basePath = "/Applications/openitcockpit-agent"
	return nil
}

func (p *darwinPlatformPath) LogPath() string {
	return filepath.Join(p.basePath, "agent.log")
}

func (p *darwinPlatformPath) ConfigPath() string {
	return p.basePath
}

func (p *darwinPlatformPath) AdditionalData() map[string]string {
	return map[string]string{}
}

func getPlatformPath() PlatformPath {
	return &darwinPlatformPath{}
}
