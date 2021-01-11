package checks

func getPlatformChecks() []Check {
	return []Check{
		&CheckMem{},
		&CheckProcess{},
		&CheckAgent{},
		&CheckSwap{},
		&CheckUser{},
		&CheckDisk{},
		&CheckDiskIo{},
		&CheckLoad{},
		&CheckCpu{},
		&CheckNet{},
		&CheckSensor{},
		&CheckDocker{},
		&CheckLaunchd{},
	}
}
