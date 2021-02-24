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
		&CheckNet{},
		&CheckNetIo{},
		&CheckCpu{},
		&CheckDocker{},
		&CheckWinService{},
		&CheckWindowsEventLog{},
	}
}
