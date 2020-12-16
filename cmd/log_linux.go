package cmd

func platformLogFile() string {
	if testLogPath != "" {
		return testLogPath
	}
	return "/var/log/openitcockpit-agent/agent.log"
}
