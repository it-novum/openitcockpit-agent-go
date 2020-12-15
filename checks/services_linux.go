package checks

import (
	"context"
	"fmt"

	"github.com/coreos/go-systemd/dbus"
)

// CheckServices gathers information about Systemd services
type CheckServices struct {
}

// Name will be used in the response as check name
func (c *CheckServices) Name() string {
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
func (c *CheckServices) Run(ctx context.Context) (*CheckResult, error) {
	err := nil
	if err != nil {
		return nil, err
	}

	conn, err := dbus.New()
	if err != nil {
		return nil, fmt.Errorf("couldn't get dbus connection: %s", err)
	}
	units, err := conn.ListUnits()
	conn.Close()
	fmt.Println(units)
	//return units, err

	return &CheckResult{
		Result: &resultServices{
			Load1:  1,
			Load5:  5,
			Load15: 15,
		},
	}, nil
}

// DefaultConfiguration contains the variables for the configuration file and the default values
// can be nil if no configuration is required
func (c *CheckServices) DefaultConfiguration() interface{} {
	return nil
}

// Configure should verify the configuration and set it
// will be run after every reload
// if DefaultConfiguration returns nil, the parameter will also be nil
func (c *CheckServices) Configure(_ interface{}) error {
	return nil
}
