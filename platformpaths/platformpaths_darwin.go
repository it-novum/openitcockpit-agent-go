package platformpaths

import "path"

type darwinPlatformPath struct {
	basePath string
}

func (p *darwinPlatformPath) Init() error {
	p.basePath = "/Applications/openitcockpit-agent"
	return nil
}

func (p *darwinPlatformPath) LogPath() string {
	return path.Join(p.basePath, "agent.log")
}

func (p *darwinPlatformPath) ConfigPath() string {
	return p.basePath
}

func getPlatformPath() PlatformPath {
	return &darwinPlatformPath{}
}
