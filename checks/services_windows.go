package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"time"
	"unsafe"

	"github.com/StackExchange/wmi"
	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
)

// Win32_Service from https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-service
// DelayedAutoStart: This property is not supported before Windows Server 2016 and Windows 10.
// nolint:underscore
type Win32_Service struct {
	Description string
	Name        string
	ProcessId   uint32
	State       string
	StartMode   string
	PathName    string
	DisplayName string
	/*
		AcceptPause       bool
		AcceptStop        bool
		Caption           string
		CheckPoint        uint32
		CreationClassName string
		//DelayedAutoStart        bool
		DesktopInteract         bool
		ErrorControl            string
		ExitCode                uint32
		InstallDate             time.Duration
		ServiceSpecificExitCode uint32
		ServiceType             string
		Started                 bool
		StartName               string
		Status                  string
		SystemCreationClassName string
		SystemName              string
		TagId                   uint32
		WaitHint                uint32
	*/
}

// CheckWinService gathers information about Windows Services services
type CheckWinService struct {
	serviceConfigCache map[string]*serviceConfig
	WmiExecutorPath    string // Used for Windows
}

// Name will be used in the response as check name
func (c *CheckWinService) Name() string {
	return "windows_services"
}

type serviceConfig struct {
	Description string
	BinPath     string
	StartType   string
}

type ResultWindowsServices struct {
	DisplayName string // Xbox Live Authentifizierungs-Manager
	BinPath     string // C:\\Windows\\system32\\svchost.exe -k netsvcs -p
	StartType   string // Manual
	Status      string // Stopped
	Pid         uint32 // 1337
	Name        string // XblAuthManager
	Description string // Stellt Authentifizierungs- und Autorisierungsservices für Xbox Live bereit. Wenn der Service beendet wird, funktionieren einige Anwendungen möglicherweise nicht richtig.
}

// Unfortunately the WMI library is suffering from a memory leak
// especially on windows Server 2016 and Windows 10.
// For this reason all WMI queries have been moved to an external binary (fork -> exec) to avoid any memory issues.
//
// Hopefully the memory issues will be fixed one day.
// This check used to look like this: https://github.com/it-novum/openitcockpit-agent-go/blob/a8ec01146e419a2db246844ca95cbe4ea560d9e6/checks/services_windows.go

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckWinService) Run(ctx context.Context) (interface{}, error) {
	// exec wmiexecutor.exe to avoid memory leak
	timeout := 10 * time.Second
	commandResult, err := utils.RunCommand(ctx, utils.CommandArgs{
		Command: c.WmiExecutorPath + " --command services",
		Shell:   "",
		Timeout: timeout,
		Env: []string{
			"OITC_AGENT_WMI_EXECUTOR=1",
		},
	})

	if err != nil {
		return nil, err
	}

	if commandResult.RC > 0 {
		return nil, fmt.Errorf(commandResult.Stdout)
	}
	fmt.Println(commandResult.Stdout)

	var dst []*ResultWindowsServices
	err = json.Unmarshal([]byte(commandResult.Stdout), &dst)

	if err != nil {
		return nil, err
	}

	return dst, nil
}

func GetServiceViaWmi(name string) (*ResultWindowsServices, error) {
	var dst []Win32_Service
	err := wmi.Query(fmt.Sprintf("SELECT * FROM Win32_Service WHERE name = '%s'", name), &dst)
	if err != nil {
		return nil, err
	}
	if len(dst) > 0 {
		service := dst[0]
		return &ResultWindowsServices{
			DisplayName: service.DisplayName,
			BinPath:     service.PathName,
			StartType:   service.StartMode,
			Status:      service.State,
			Pid:         service.ProcessId,
			Name:        service.Name,
			Description: service.Description,
		}, nil
	}
	return nil, nil
}

func (c *CheckWinService) GetServiceListViaWmi(_ context.Context) ([]*ResultWindowsServices, error) {
	var dst []Win32_Service
	err := wmi.Query("SELECT * FROM Win32_Service", &dst)
	if err != nil {
		return nil, err
	}

	wmiResults := make([]*ResultWindowsServices, 0, len(dst))
	for _, service := range dst {
		result := &ResultWindowsServices{
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

func serviceStatusBufferToSlice(buf []byte, servicesReturned uint32) []windows.ENUM_SERVICE_STATUS_PROCESS {
	var processList []windows.ENUM_SERVICE_STATUS_PROCESS
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&processList))
	sliceHeader.Cap = int(servicesReturned)
	sliceHeader.Len = int(servicesReturned)
	sliceHeader.Data = uintptr(unsafe.Pointer(&buf[0]))
	return processList
}

var mapServiceState = map[uint32]string{
	windows.SERVICE_CONTINUE_PENDING: "Continue Pending",
	windows.SERVICE_PAUSE_PENDING:    "Pause Pending",
	windows.SERVICE_PAUSED:           "Paused",
	windows.SERVICE_RUNNING:          "Running",
	windows.SERVICE_START_PENDING:    "Start Pending",
	windows.SERVICE_STOP_PENDING:     "Stop Pending",
	windows.SERVICE_STOPPED:          "Stopped",
}

var mapServiceStartType = map[uint32]string{
	windows.SERVICE_AUTO_START:   "Auto",
	windows.SERVICE_BOOT_START:   "Boot",
	windows.SERVICE_DEMAND_START: "Manual",
	windows.SERVICE_DISABLED:     "Disabled",
	windows.SERVICE_SYSTEM_START: "System",
}

func fetchServiceConfiguration(scManager windows.Handle, serviceName *uint16, service *ResultWindowsServices) error {
	srvHandle, err := windows.OpenService(scManager, serviceName, windows.GENERIC_READ)
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("could not query service configuration for service ", windows.UTF16PtrToString(serviceName)))
	}
	defer func() {
		_ = windows.CloseServiceHandle(srvHandle)
	}()

	// we have to query both before return

	errSC := queryServiceConfig(srvHandle, serviceName, service)
	errSD := queryServiceDescription(srvHandle, serviceName, service)

	if errSC != nil {
		return errSC
	}

	if errSD != nil {
		return errSD
	}

	return nil
}

func queryServiceConfig(srvHandle windows.Handle, serviceName *uint16, service *ResultWindowsServices) error {
	var (
		bytesNeeded uint32
	)
	if err := windows.QueryServiceConfig(srvHandle, nil, 0, &bytesNeeded); err != windows.ERROR_INSUFFICIENT_BUFFER {
		return errors.Wrap(err, fmt.Sprint("could not fetch required buffer size for windows service configuration query for service ", windows.UTF16PtrToString(serviceName)))
	}

	buf := make([]byte, bytesNeeded)
	var serviceConfig = (*windows.QUERY_SERVICE_CONFIG)(unsafe.Pointer(&buf[0]))

	if err := windows.QueryServiceConfig(srvHandle, serviceConfig, bytesNeeded, &bytesNeeded); err != nil {
		return errors.Wrap(err, fmt.Sprint("could not fetch windows service configuration for service ", windows.UTF16PtrToString(serviceName)))
	}

	service.BinPath = windows.UTF16PtrToString(serviceConfig.BinaryPathName)

	service.StartType = mapServiceStartType[serviceConfig.StartType]

	serviceConfig = nil
	buf = nil

	bytesNeeded = 0

	if err := windows.QueryServiceConfig2(srvHandle, windows.SERVICE_CONFIG_DESCRIPTION, nil, 0, &bytesNeeded); err != windows.ERROR_INSUFFICIENT_BUFFER {
		return errors.Wrap(err, fmt.Sprint("could not fetch required buffer size for windows service description query for service ", windows.UTF16PtrToString(serviceName)))
	}

	buf = make([]byte, bytesNeeded)
	var serviceDescription = (*windows.SERVICE_DESCRIPTION)(unsafe.Pointer(&buf[0]))
	if err := windows.QueryServiceConfig2(srvHandle, windows.SERVICE_CONFIG_DESCRIPTION, &buf[0], bytesNeeded, &bytesNeeded); err != windows.ERROR_INSUFFICIENT_BUFFER {
		return errors.Wrap(err, fmt.Sprint("could not fetch windows service description for service ", windows.UTF16PtrToString(serviceName)))
	}

	service.Description = windows.UTF16PtrToString(serviceDescription.Description)
	serviceDescription = nil
	buf = nil

	return nil
}

func queryServiceDescription(srvHandle windows.Handle, serviceName *uint16, service *ResultWindowsServices) error {
	var (
		bytesNeeded uint32
	)

	if err := windows.QueryServiceConfig2(srvHandle, windows.SERVICE_CONFIG_DESCRIPTION, nil, 0, &bytesNeeded); err != windows.ERROR_INSUFFICIENT_BUFFER {
		return errors.Wrap(err, fmt.Sprint("could not fetch required buffer size for windows service description query for service ", windows.UTF16PtrToString(serviceName)))
	}

	buf := make([]byte, bytesNeeded)
	var serviceDescription = (*windows.SERVICE_DESCRIPTION)(unsafe.Pointer(&buf[0]))
	if err := windows.QueryServiceConfig2(srvHandle, windows.SERVICE_CONFIG_DESCRIPTION, &buf[0], bytesNeeded, &bytesNeeded); err != nil {
		return errors.Wrap(err, fmt.Sprint("could not fetch windows service description for service ", windows.UTF16PtrToString(serviceName)))
	}

	service.Description = windows.UTF16PtrToString(serviceDescription.Description)
	serviceDescription = nil
	buf = nil

	return nil
}

func (c *CheckWinService) GetServiceListViaCAPI(_ context.Context) ([]*ResultWindowsServices, error) {
	if c.serviceConfigCache == nil {
		c.serviceConfigCache = make(map[string]*serviceConfig)
	}

	log.Debugln("Check Services: open windows service manager")
	scManager, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_ENUMERATE_SERVICE)
	if err != nil {
		return nil, errors.Wrap(err, "could not open windows service manager")
	}
	defer func() {
		_ = windows.CloseServiceHandle(scManager)
	}()

	var (
		bytesNeeded      uint32
		servicesReturned uint32
		resumeHandle     uint32
	)

	log.Debugln("Check Services: query service list")
	if err := windows.EnumServicesStatusEx(scManager, windows.SC_STATUS_PROCESS_INFO, windows.SERVICE_WIN32, windows.SERVICE_STATE_ALL, nil, 0, &bytesNeeded, &servicesReturned, &resumeHandle, nil); err != windows.ERROR_MORE_DATA {
		return nil, errors.Wrap(err, "could not fetch buffer size for EnumServicesStatusEx")
	}

	buf := make([]byte, bytesNeeded)
	if err := windows.EnumServicesStatusEx(scManager, windows.SC_STATUS_PROCESS_INFO, windows.SERVICE_WIN32, windows.SERVICE_STATE_ALL, &buf[0], bytesNeeded, &bytesNeeded, &servicesReturned, &resumeHandle, nil); err != nil {
		return nil, errors.Wrap(err, "could not query windows service list")
	}

	services := serviceStatusBufferToSlice(buf, servicesReturned)
	newCache := make(map[string]*serviceConfig, len(services))

	log.Debugln("Check Services: query service configuration")
	result := make([]*ResultWindowsServices, 0)
	for _, service := range services {
		srvResult := &ResultWindowsServices{
			DisplayName: windows.UTF16PtrToString(service.DisplayName),
			Status:      mapServiceState[service.ServiceStatusProcess.CurrentState],
			Pid:         service.ServiceStatusProcess.ProcessId,
			Name:        windows.UTF16PtrToString(service.ServiceName),
		}
		sc, ok := c.serviceConfigCache[srvResult.Name]
		if !ok {
			if err := fetchServiceConfiguration(scManager, service.ServiceName, srvResult); err != nil {
				if err == windows.ERROR_ACCESS_DENIED {
					// we don't have access to all service configurations so we have to use wmi for this
					if wmiSrv, err := GetServiceViaWmi(srvResult.Name); err != nil || wmiSrv == nil {
						log.Errorln("could not fetch service configuration: ", err)
					} else {
						srvResult.Description = wmiSrv.Description
						srvResult.BinPath = wmiSrv.BinPath
						srvResult.StartType = wmiSrv.StartType
					}
				} else {
					log.Errorln("could not fetch service configuration: ", err)
				}
			}
			sc = &serviceConfig{
				Description: srvResult.Description,
				BinPath:     srvResult.BinPath,
				StartType:   srvResult.StartType,
			}
		} else {
			srvResult.BinPath = sc.BinPath
			srvResult.Description = sc.Description
			srvResult.StartType = sc.StartType
		}
		newCache[srvResult.Name] = sc

		result = append(result, srvResult)
	}

	c.serviceConfigCache = newCache

	log.Debugln("Check Services: found ", len(result), " services")
	return result, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckWinService) Configure(config *config.Configuration) (bool, error) {
	if config.WindowsServices && runtime.GOOS == "windows" {
		// Check is enabled
		agentBinary, err := os.Executable()
		if err == nil {
			wmiPath := filepath.Dir(agentBinary) + string(os.PathSeparator) + "wmiexecutor.exe"
			c.WmiExecutorPath = wmiPath
		}
	}

	return config.WindowsServices, nil
}
