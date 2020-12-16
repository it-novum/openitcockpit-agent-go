package checks

import (
	"context"

	"github.com/coreos/go-systemd/v22/dbus"
)

// CheckSystemd gathers information about Systemd services
type CheckSystemd struct {
}

// Name will be used in the response as check name
func (c *CheckSystemd) Name() string {
	return "systemd_services"
}

type resultSystemdServices struct {
	ActiveState string
	Description string
	LoadState   string
	Name        string
	SubState    string
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckSystemd) Run(ctx context.Context) (*CheckResult, error) {
	systemdResults, err := c.getServiceListViaDbus(ctx)
	if err != nil {
		return nil, err
	}
	return &CheckResult{Result: systemdResults}, nil
}

func (c *CheckSystemd) getServiceListViaDbus(ctx context.Context) ([]*resultSystemdServices, error) {
	conn, err := dbus.New()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	units, err := conn.ListUnits()
	if err != nil {
		return nil, err
	}

	systemdResults := make([]*resultSystemdServices, 0, len(units))

	for _, unit := range units {
		result := &resultSystemdServices{
			ActiveState: unit.ActiveState,
			Description: unit.Description,
			LoadState:   unit.LoadState,
			Name:        unit.Name,
			SubState:    unit.SubState,
		}
		systemdResults = append(systemdResults, result)
	}

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
