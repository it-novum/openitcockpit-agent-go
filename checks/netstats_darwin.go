package checks

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/net"
)

// CheckNet gathers information about system network interfaces (netstats or net_states in the Python version)
type CheckNetstats struct {
}

// Name will be used in the response as check name
func (c *CheckNetstats) Name() string {
	return "net_io"
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckNetstats) Run(ctx context.Context) (interface{}, error) {
	nics, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	nicIo, err := net.IOCountersWithContext(ctx, false)
	if err != nil {
		return nil, err
	}
	fmt.Println(nicIo)

	nicResults := make(map[string]*resultNet)

	for _, nic := range nics {
		isUp := false
		for _, flag := range nic.Flags {
			if flag == "up" {
				isUp = true
			}
		}

		nicResults[nic.Name] = &resultNet{
			Isup: isUp,
			MTU:  int64(nic.MTU),
		}

	}

	return nicResults, nil
}
