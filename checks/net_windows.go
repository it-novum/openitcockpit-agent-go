package checks

import (
	"context"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/winifmib"
	"golang.org/x/sys/windows"
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

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckNet) Run(_ context.Context) (interface{}, error) {
	netResults := make(map[string]*resultNet)

	mibTable, err := winifmib.GetIfTable2Ex(true)
	if err != nil {
		return nil, err
	}
	defer mibTable.Close()

	table := mibTable.Slice()
	for _, nic := range table {
		name := windows.UTF16ToString(nic.Alias[:])
		if name != "" {
			netResults[name] = &resultNet{
				Isup:   nic.OperStatus == 1,
				MTU:    int64(nic.Mtu),
				Speed:  int64(nic.ReceiveLinkSpeed) / 1000 / 1000, // bits/s to mbits/s
				Duplex: DUPLEX_UNKNOWN,
			}
		}
	}

	return netResults, nil
}
