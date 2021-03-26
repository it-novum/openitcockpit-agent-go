package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/utils"
)

// https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-operatingsystem
type Win32_OperatingSystem struct {
	BootDevice                                string
	BuildNumber                               string
	BuildType                                 string
	Caption                                   string
	CodeSet                                   string
	CountryCode                               string
	CreationClassName                         string
	CSCreationClassName                       string
	CSDVersion                                string
	CSName                                    string
	CurrentTimeZone                           int16
	DataExecutionPrevention_Available         bool
	DataExecutionPrevention_32BitApplications bool
	DataExecutionPrevention_Drivers           bool
	DataExecutionPrevention_SupportPolicy     uint8
	Debug                                     bool
	Description                               string
	Distributed                               bool
	EncryptionLevel                           uint32
	ForegroundApplicationBoost                uint8
	FreePhysicalMemory                        uint64
	FreeSpaceInPagingFiles                    uint64
	FreeVirtualMemory                         uint64
	LargeSystemCache                          uint32
	LastBootUpTime                            time.Time
	Locale                                    string
	Manufacturer                              string
	MaxNumberOfProcesses                      uint32
	MaxProcessMemorySize                      uint64
	MUILanguages                              []string
	Name                                      string
	NumberOfLicensedUsers                     uint32
	NumberOfProcesses                         uint32
	NumberOfUsers                             uint32
	OperatingSystemSKU                        uint32
	Organization                              string
	OSArchitecture                            string
	OSLanguage                                uint32
	OSProductSuite                            uint32
	OSType                                    uint16
	OtherTypeDescription                      string
	PAEEnabled                                bool
	PlusProductID                             string
	PlusVersionNumber                         string
	Primary                                   bool
	ProductType                               uint32
	RegisteredUser                            string
	SerialNumber                              string
	ServicePackMajorVersion                   uint16
	ServicePackMinorVersion                   uint16
	SizeStoredInPagingFiles                   uint64
	Status                                    string
	SuiteMask                                 uint32
	SystemDevice                              string
	SystemDirectory                           string
	SystemDrive                               string
	TotalSwapSpaceSize                        uint64
	TotalVirtualMemorySize                    uint64
	TotalVisibleMemorySize                    uint64
	Version                                   string
	WindowsDirectory                          string
}

// Unfortunately the WMI library is suffering from a memory leak
// especially on windows Server 2016 and Windows 10.
// For this reason all WMI queries have been moved to an external binary (fork -> exec) to avoid any memory issues.
//
// Hopefully the memory issues will be fixed one day.
// This check used to look like this: https://github.com/it-novum/openitcockpit-agent-go/blob/a8ec01146e419a2db246844ca95cbe4ea560d9e6/checks/memory_windows.go

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckMem) Run(ctx context.Context) (interface{}, error) {
	// exec wmiexecutor.exe to avoid memory leak
	timeout := 10 * time.Second
	commandResult, err := utils.RunCommand(ctx, utils.CommandArgs{
		Command: c.WmiExecutorPath + " --command memory",
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

	var dst []*Win32_OperatingSystem
	err = json.Unmarshal([]byte(commandResult.Stdout), &dst)

	if err != nil {
		return nil, err
	}

	var info *Win32_OperatingSystem = dst[0]

	total := info.TotalVisibleMemorySize * 1024
	free := info.FreePhysicalMemory * 1024
	used := total - free
	var percent float64 = ((float64(total) - float64(free)) * 100.0) / float64(total)

	// https://github.com/shirou/gopsutil/blob/master/v3/mem/mem_windows.go#L42-L47
	return &resultMemory{
		Total:     total,
		Available: free,
		Free:      free,
		Used:      used,
		Percent:   percent,
	}, nil
}
