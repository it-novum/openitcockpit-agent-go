package checks

import (
	"context"

	"github.com/StackExchange/wmi"
	"github.com/it-novum/openitcockpit-agent-go/windows"
)

// CheckWinService gathers information about Systemd services
type CheckWinService struct {
}

// Name will be used in the response as check name
func (c *CheckWinService) Name() string {
	return "windows_services"
}

type resultWindowsServices struct {
	DisplayName string
	BinPath     string
	StartType   string
	Status      string
	Pid         uint32
	Name        string
	Description string
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckWinService) Run(ctx context.Context) (*CheckResult, error) {
	systemdResults, err := c.getServiceListViaWmi(ctx)
	if err != nil {
		return nil, err
	}
	return &CheckResult{Result: systemdResults}, nil
}

func (c *CheckWinService) getServiceListViaWmi(ctx context.Context) ([]*resultWindowsServices, error) {
	var dst []windows.Win32_Service
	err := wmi.Query("SELECT * FROM Win32_Service", &dst)
	if err != nil {
		return nil, err
	}

	wmiResults := make([]*resultWindowsServices, 0, len(dst))
	for _, service := range dst {
		result := &resultWindowsServices{
			DisplayName: service.DisplayName,
			BinPath:     service.PathName,
			StartType:   service.StartMode,
			Status:      service.State,
			Pid:         service.ProcessId,
			Name:        service.Name,
			Description: service.Description,
		}
		wmiResults = append(wmiResults, result)
	}

	return wmiResults, nil
}

// DefaultConfiguration contains the variables for the configuration file and the default values
// can be nil if no configuration is required
func (c *CheckWinService) DefaultConfiguration() interface{} {
	return nil
}

// Configure should verify the configuration and set it
// will be run after every reload
// if DefaultConfiguration returns nil, the parameter will also be nil
func (c *CheckWinService) Configure(_ interface{}) error {
	return nil
}
