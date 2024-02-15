//go:build !libvirt
// +build !libvirt

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
		&CheckNetIo{},
		&CheckSensor{},
		&CheckDocker{},
		&CheckSystemd{},
		&CheckNtp{},
	}
}
