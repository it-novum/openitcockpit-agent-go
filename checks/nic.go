package checks

import (
	"context"
	"fmt"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/shirou/gopsutil/v3/net"
)

// CheckNic gathers information about system network interfaces (netstats or net_states in the Python version)
type CheckNic struct {
}

// Name will be used in the response as check name
func (c *CheckNic) Name() string {
	return "net_stats"
}

type resultNic struct {
	Isup   bool `json:"isup"`
	Duplex int  `json:"duplex"`
	Speed  int  `json:"speed"`
	MTU    int  `json:"mtu"`
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckNic) Run(ctx context.Context) (*CheckResult, error) {
	nics, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	nicIo, err := net.IOCountersWithContext(ctx, false)
	if err != nil {
		return nil, err
	}
	fmt.Println(nicIo)

	nicResults := make(map[string]*resultNic)

	for _, nic := range nics {
		isUp := false
		for _, flag := range nic.Flags {
			if flag == "up" {
				isUp = true
			}
		}

		nicResults[nic.Name] = &resultNic{
			Isup: isUp,
			MTU:  nic.MTU,
		}

	}

	return &CheckResult{Result: nicResults}, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckNic) Configure(config *config.Configuration) (bool, error) {
	return true, nil
}
