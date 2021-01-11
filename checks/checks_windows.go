package checks

func getPlatformChecks() []Check {
	return []Check{
		&CheckMem{},
		&CheckProcess{},
		&CheckAgent{},
		&CheckSwap{},
		&CheckUser{},
		&CheckDisk{},
		//&CheckDiskIo{}, //Check that all calcs are done right
		//&CheckNet{},
		&CheckNetIo{},
		&CheckCpu{},
		&CheckDocker{},
		&CheckWinService{},
	}
}
