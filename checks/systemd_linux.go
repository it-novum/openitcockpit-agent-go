package checks

//  https://github.com/coreos/go-systemd/issues/302#issuecomment-599412253
// 	""

import (
	"context"
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
)

// CheckSystemd gathers information about Systemd services
type CheckSystemd struct {
}

// Name will be used in the response as check name
func (c *CheckSystemd) Name() string {
	return "systemd_services"
}

type resultServices struct {
	Load1  float64 `json:"0"`
	Load5  float64 `json:"1"`
	Load15 float64 `json:"2"`
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckSystemd) Run(ctx context.Context) (*CheckResult, error) {
	return nil, fmt.Errorf("kaputt")
}

func (c *CheckSystemd) getServiceListViaDbus(ctx context.Context) ([]*CheckResult, error) {
	conn, err := dbus.New()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	allUnits, err := conn.ListUnits()
	fmt.Println(allUnits)

	systemdResults := make([]*CheckResult, 0, 1)
	return systemdResults, nil
}

// DefaultConfiguration contains the variables for the configuration file and the default values
// can be nil if no configuration is required
func (c *CheckSystemd) DefaultConfiguration() interface{} {
	return nil
}

// Configure should verify the configuration and set it
// will be run after every reload
// if DefaultConfiguration returns nil, the parameter will also be nil
func (c *CheckSystemd) Configure(_ interface{}) error {
	return nil
}
