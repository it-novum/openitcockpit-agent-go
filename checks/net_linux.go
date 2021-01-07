package checks

import (
	"context"

	"github.com/prometheus/procfs/sysfs"
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckNet) Run(ctx context.Context) (interface{}, error) {
	// https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-class-net
	fs, err := sysfs.NewFS("/sys")
	if err != nil {
		return nil, err
	}

	netClass, _ := fs.NewNetClass()

	netResults := make(map[string]*resultNet)
	for _, nic := range netClass {
		//fmt.Print(nic)

		duplex := DUPLEX_UNKNOWN

		if nic.Duplex == "full" {
			duplex = DUPLEX_FULL
		}

		if nic.Duplex == "half" {
			duplex = DUPLEX_HALF
		}

		var speed int64 = 0
		if nic.Speed != nil {
			speed = *(nic.Speed)
		}

		netResults[nic.Name] = &resultNet{
			Isup:   *(nic.Carrier) == 1,
			MTU:    *(nic.MTU),
			Speed:  speed,
			Duplex: duplex,
		}

	}

	return netResults, nil
}
