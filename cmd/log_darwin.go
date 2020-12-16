package cmd

func platformLogFile() string {
	if testLogPath != "" {
		return testLogPath
	}
	return "/Applications/openitcockpit-agent/agent.log"
}
