package checks

import (
	"context"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/it-novum/openitcockpit-agent-go/config"
)

// Win32_Service from https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-service
// nolint:underscore
type Win32_Service struct {
	AcceptPause             bool
	AcceptStop              bool
	Caption                 string
	CheckPoint              uint32
	CreationClassName       string
	DelayedAutoStart        bool
	Description             string
	DesktopInteract         bool
	DisplayName             string
	ErrorControl            string
	ExitCode                uint32
	InstallDate             time.Duration
	Name                    string
	PathName                string
	ProcessId               uint32
	ServiceSpecificExitCode uint32
	ServiceType             string
	Started                 bool
	StartMode               string
	StartName               string
	State                   string
	Status                  string
	SystemCreationClassName string
	SystemName              string
	TagId                   uint32
	WaitHint                uint32
}

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
func (c *CheckWinService) Run(ctx context.Context) (interface{}, error) {
	return c.getServiceListViaWmi(ctx)
}

func (c *CheckWinService) getServiceListViaWmi(ctx context.Context) ([]*resultWindowsServices, error) {
	var dst []Win32_Service
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

// Configure the command or return false if the command was disabled
func (c *CheckWinService) Configure(config *config.Configuration) (bool, error) {
	return true, nil
}
