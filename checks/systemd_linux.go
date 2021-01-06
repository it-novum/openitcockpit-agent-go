package checks

import (
	"context"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/it-novum/openitcockpit-agent-go/config"
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
func (c *CheckSystemd) Run(ctx context.Context) (interface{}, error) {
	return c.getServiceListViaDbus(ctx)
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

// Configure the command or return false if the command was disabled
func (c *CheckSystemd) Configure(config *config.Configuration) (bool, error) {
	return config.SystemdServices, nil
}
