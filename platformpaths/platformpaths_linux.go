package platformpaths

type linuxPlatformPath struct {
}

func (p *linuxPlatformPath) Init() error {
	return nil
}

func (p *linuxPlatformPath) LogPath() string {
	return "/var/log/agent.log"
}

func (p *linuxPlatformPath) ConfigPath() string {
	return "/etc/openitcockpit-agent/"
}

func getPlatformPath() PlatformPath {
	return &linuxPlatformPath{}
}
