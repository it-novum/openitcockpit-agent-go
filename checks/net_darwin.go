package checks

import (
	"context"
	"net"
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckNet) Run(ctx context.Context) (interface{}, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	nicResults := make(map[string]*resultNet)
	for _, nic := range ifs {
		isUp := false
		if nic.Flags&net.FlagUp != 0 {
			isUp = true
		}

		nicResults[nic.Name] = &resultNet{
			Isup: isUp,
			MTU:  int64(nic.MTU),
		}
	}

	return nicResults, nil
}
