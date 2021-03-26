package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/utils"
)

// WMI Structs
// https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-networkadapter
type Win32_NetworkAdapter struct {
	AdapterType                 string
	AdapterTypeID               uint16
	AutoSense                   bool
	Availability                uint16
	Caption                     string
	ConfigManagerErrorCode      uint32
	ConfigManagerUserConfig     bool
	CreationClassName           string
	Description                 string
	DeviceID                    string
	ErrorCleared                bool
	ErrorDescription            string
	GUID                        string
	Index                       uint32
	InstallDate                 time.Time
	Installed                   bool
	InterfaceIndex              uint32
	LastErrorCode               uint32
	MACAddress                  string
	Manufacturer                string
	MaxNumberControlled         uint32
	MaxSpeed                    uint64
	Name                        string
	NetConnectionID             string
	NetConnectionStatus         uint16
	NetEnabled                  bool
	NetworkAddresses            []string
	PermanentAddress            string
	PhysicalAdapter             bool
	PNPDeviceID                 string
	PowerManagementCapabilities []uint16
	PowerManagementSupported    bool
	ProductName                 string
	ServiceName                 string
	Speed                       uint64
	Status                      string
	StatusInfo                  uint16
	SystemCreationClassName     string
	SystemName                  string
	TimeOfLastReset             time.Time
}

// https://docs.microsoft.com/en-us/previous-versions/aa394293(v=vs.85)
type Win32_PerfFormattedData_Tcpip_NetworkInterface struct {
	BytesReceivedPerSec             uint32
	BytesSentPerSec                 uint32
	BytesTotalPerSec                uint64
	Caption                         string
	CurrentBandwidth                uint32
	Description                     string
	Frequency_Object                uint64
	Frequency_PerfTime              uint64
	Frequency_Sys100NS              uint64
	Name                            string
	OutputQueueLength               uint32
	PacketsOutboundDiscarded        uint32
	PacketsOutboundErrors           uint32
	PacketsPerSec                   uint32
	PacketsReceivedDiscarded        uint32
	PacketsReceivedErrors           uint32
	PacketsReceivedNonUnicastPerSec uint32
	PacketsReceivedPerSec           uint32
	PacketsReceivedUnicastPerSec    uint32
	PacketsReceivedUnknown          uint32
	PacketsSentNonUnicastPerSec     uint32
	PacketsSentPerSec               uint32
	PacketsSentUnicastPerSec        uint32
	Timestamp_Object                uint64
	Timestamp_PerfTime              uint64
	Timestamp_Sys100NS              uint64
}

type MSFT_NetAdapter struct {
	Name              string
	Status            string
	FullDuplex        bool
	MediaDuplexState  uint32
	MtuSize           uint32
	VlanID            uint16
	TransmitLinkSpeed uint64
	ReceiveLinkSpeed  uint64
}

// Unfortunately the WMI library is suffering from a memory leak
// especially on windows Server 2016 and Windows 10.
// For this reason all WMI queries have been moved to an external binary (fork -> exec) to avoid any memory issues.
//
// Hopefully the memory issues will be fixed one day.
// This check used to look like this: https://github.com/it-novum/openitcockpit-agent-go/blob/a8ec01146e419a2db246844ca95cbe4ea560d9e6/checks/net_windows.go

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckNet) Run(ctx context.Context) (interface{}, error) {
	// exec wmiexecutor.exe to avoid memory leak
	timeout := 10 * time.Second
	commandResult, err := utils.RunCommand(ctx, utils.CommandArgs{
		Command: c.WmiExecutorPath + " --command net",
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

	var dst map[string]*resultNet
	err = json.Unmarshal([]byte(commandResult.Stdout), &dst)

	if err != nil {
		return nil, err
	}

	return dst, nil
}
