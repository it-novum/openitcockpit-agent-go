package winifmib

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"unsafe"
)

type NetworkInterface struct {
	Name            string
	BytesReceived   uint64 //InOctets
	DropIn          uint64 //InDiscards
	ErrorIn         uint64 //InErrors
	BytesSent       uint64 //OutOctets
	DropOut         uint64 //OutDiscards
	ErrorOut        uint64 //OutErrors
	PacketsSent     uint64 //OutUcastPkts+OutNUcastPkts
	PacketsReceived uint64 //InUcastPkts+InNUcastPkts
}

type MibIfEntryLevel uint32

const (
	MibIfEntryNormal                  = MibIfEntryLevel(0)
	MibIfEntryNormalWithoutStatistics = MibIfEntryLevel(2)
	IfMaxStringSize                   = 256
	IfMaxPhysAddressLength            = 32
)

type Ulong uint32

type NetLUID uint64
type NetIfIndex Ulong
type TunnelType uint32
type NdisMedium uint32
type NdisPhysicalMedium uint32
type NetIfAccessType uint32
type NetIfDirectionType uint32
type InterfaceAndOperStatusFlags [8]byte // this field is padded in the struct and is not just 1 byte
type IfOperStatus uint32
type NetIfAdminStatus uint32
type NetIfMediaConnectState uint32
type NetIfConnectionType uint32

/* MibIfRow2 from https://docs.microsoft.com/en-us/windows/win32/api/netioapi/ns-netioapi-mib_if_row2

Important: This C struct is padded by the C compiler, so 1 byte fields are actually 8 byte
Go uses a different padding scheme depending on GOARCH... the windows C compiler does not...
*/
type MibIfRow2 struct {
	InterfaceLuid               NetLUID
	InterfaceIndex              NetIfIndex
	InterfaceGuid               windows.GUID
	Alias                       [IfMaxStringSize + 1]uint16
	Description                 [IfMaxStringSize + 1]uint16
	PhysicalAddressLength       Ulong
	PhysicalAddress             [IfMaxPhysAddressLength]byte
	PermanentPhysicalAddress    [IfMaxPhysAddressLength]byte
	Mtu                         Ulong
	Type                        Ulong
	TunnelType                  TunnelType
	MediaType                   NdisMedium
	PhysicalMediumType          NdisPhysicalMedium
	AccessType                  NetIfAccessType
	DirectionType               NetIfDirectionType
	InterfaceAndOperStatusFlags InterfaceAndOperStatusFlags
	OperStatus                  IfOperStatus
	AdminStatus                 NetIfAdminStatus
	MediaConnectState           NetIfMediaConnectState
	NetworkGuid                 windows.GUID
	ConnectionType              NetIfConnectionType
	TransmitLinkSpeed           uint64
	ReceiveLinkSpeed            uint64
	InOctets                    uint64
	InUcastPkts                 uint64
	InNUcastPkts                uint64
	InDiscards                  uint64
	InErrors                    uint64
	InUnknownProtos             uint64
	InUcastOctets               uint64
	InMulticastOctets           uint64
	InBroadcastOctets           uint64
	OutOctets                   uint64
	OutUcastPkts                uint64
	OutNUcastPkts               uint64
	OutDiscards                 uint64
	OutErrors                   uint64
	OutUcastOctets              uint64
	OutMulticastOctets          uint64
	OutBroadcastOctets          uint64
	OutQLen                     uint64
}

func (r *MibIfRow2) NetworkInterface() *NetworkInterface {
	return &NetworkInterface{
		Name:            windows.UTF16ToString(r.Alias[:]),
		BytesReceived:   r.InOctets,
		DropIn:          r.InDiscards,
		ErrorIn:         r.InErrors,
		BytesSent:       r.OutOctets,
		DropOut:         r.OutDiscards,
		ErrorOut:        r.OutErrors,
		PacketsSent:     r.OutUcastOctets + r.OutNUcastPkts,
		PacketsReceived: r.InUcastPkts + r.InNUcastPkts,
	}
}

/* MibIfTable2 from https://docs.microsoft.com/en-us/windows/win32/api/netioapi/ns-netioapi-mib_if_table2

Important: This C struct is padded by the C compiler, so 1 byte fields are actually 8 byte
Go uses a different padding scheme depending on GOARCH... the windows C compiler does not...

In this case NumEntires IS a 4 byte value, BUT the C compiler adds 4 bytes for memory alignment... Go would
add the memory at the end, so this would break everything.

*/
type MibIfTable2 struct {
	NumEntries Ulong
	_          [4]byte //pad
	// Table isn't a pointer but a C array value... C doesn't care that the array field length is 1, so in C we
	// can iterate over this. In Go this is not allowed, so have to calculate the actual memory address.
	Table [1]MibIfRow2
}

type PMibIfTable *MibIfTable2

var MibIfRow2Size = unsafe.Sizeof(MibIfRow2{})

// Get returns a pointer to a value in the C array
func (t *MibIfTable2) Get(index Ulong) *MibIfRow2 {
	if index < 0 || index >= t.NumEntries {
		panic("out of bounds")
	}
	offset := MibIfRow2Size * uintptr(index)
	base := uintptr(unsafe.Pointer(&t.Table))
	return (*MibIfRow2)(unsafe.Pointer(base + offset))
}

func (t *MibIfTable2) Slice() []*MibIfRow2 {
	table := make([]*MibIfRow2, t.NumEntries)
	for i := Ulong(0); i < t.NumEntries; i++ {
		table[i] = t.Get(i)
	}
	return table
}

func NetworkInterfaceStatistics() ([]*NetworkInterface, error) {
	t, err := GetIfTable2Ex(true)
	if err != nil {
		return nil, err
	}
	defer t.Close()

	result := make([]*NetworkInterface, 0, t.NumEntries)
	for i := Ulong(0); i < t.NumEntries; i++ {
		row := t.Get(i)
		if row.InterfaceIndex == 1 {
			log.Errorln("type: ", row.Type)
		}
		result = append(result, row.NetworkInterface())
	}
	return result, nil
}

func (t *MibIfTable2) Close() {
	FreeMibTable(t)
}

func GetIfEntry2Ex(index NetIfIndex, stats bool) (*MibIfRow2, error) {
	row := &MibIfRow2{
		InterfaceIndex: index,
	}
	level := MibIfEntryNormal
	if !stats {
		level = MibIfEntryNormalWithoutStatistics
	}
	err := getIfEntry2Ex(level, row)
	return row, err
}

// GetIfTable2Ex generates interface statistics of all interfaces with one syscall.
// Unfortunatly it uses dynamic memory magic and must be freed with FreeMibTable!
// The syscall safes a pointer to the magic memory space, so this is why we have to allocate the pointer
// and call the syscall with a pointer pointer.
func GetIfTable2Ex(stats bool) (*MibIfTable2, error) {
	var table PMibIfTable
	level := MibIfEntryNormal
	if !stats {
		level = MibIfEntryNormalWithoutStatistics
	}
	err := getIfTable2Ex(level, &table)

	return table, err
}

func FreeMibTable(memory PMibIfTable) {
	freeMibTable(uintptr(unsafe.Pointer(memory)))
}

//go:generate mkwinsyscall -output zsyscall_windows.go ifmib_windows.go
//sys getIfEntry2Ex(level MibIfEntryLevel, row *MibIfRow2) (err error) [failretval!=0] = Iphlpapi.GetIfEntry2Ex
//sys getIfTable2Ex(level MibIfEntryLevel, table *PMibIfTable) (err error) [failretval!=0] = Iphlpapi.GetIfTable2Ex
//sys freeMibTable(memory uintptr) = Iphlpapi.FreeMibTable
