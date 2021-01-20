// +build darwin linux

// sudo apt-get install libvirt-dev

package checks

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/safemaths"
	libvirt "libvirt.org/libvirt-go"
)

// CheckLibvirt gathers information about running VMs
type CheckLibvirt struct {
	lastNetstatResults   map[string]map[string]*lastNetstatResultsForDelta
	lastDiskstatsResults map[string]map[string]*lastDiskstatsResultsForDelta
}

type lastNetstatResultsForDelta struct {
	Timestamp int64
	RxBytes   int64
	RxPackets int64
	RxErrs    int64
	RxDrop    int64
	TxBytes   int64
	TxPackets int64
	TxErrs    int64
	TxDrop    int64
}

type lastDiskstatsResultsForDelta struct {
	Timestamp int64
	WrReq     int64
	RdReq     int64
	RdBytes   int64
	WrBytes   int64
}

// Name will be used in the response as check name
func (c *CheckLibvirt) Name() string {
	return "libvirt"
}

// This code is highly inspired by
// https://github.com/feiskyer/go-examples/blob/master/libvirt/libvirt-stats/libvirt.go
// Many thanks
type XmlDomain struct {
	Name    string     `xml:"name"`
	Uuid    string     `xml:"uuid"`
	Devices XmlDevices `xml:"devices"`
}

type XmlDisk struct {
	Type   string        `xml:"type,attr"`
	Source XmlDiskSource `xml:"source"`
	Target XmlDiskTarget `xml:"target"`
}
type XmlDiskSource struct {
	File string `xml:"file,attr"`
}
type XmlDiskTarget struct {
	Dev string `xml:"dev,attr"`
}

type XmlDevices struct {
	Disks      []XmlDisk      `xml:"disk"`
	Interfaces []XmlInterface `xml:"interface"`
}
type XmlInterface struct {
	Type   string
	Device XmlInterfaceTarget `xml:"target"`
}

type XmlInterfaceTarget struct {
	Dev string `xml:"dev,attr"`
}

type virtMemory struct {
	Total      uint64 // total memory in bytes
	Ununsed    uint64
	Available  uint64
	Rss        uint64
	SwapIn     uint64
	SwapOut    uint64
	MinorFault uint64
	MajorFault uint64
}

type resultLibvirtNetwork struct {
	Name                        string `json:"name"`                // Name of the network interface
	AvgBytesSentPerSecond       uint64 `json:"avg_bytes_sent_ps"`   // Average bytes sent per second
	AvgBytesReceivedPerSecond   uint64 `json:"avg_bytes_recv_ps"`   // Average bytes received per second
	AvgPacketsSentPerSecond     uint64 `json:"avg_packets_sent_ps"` // Average packets sent per second
	AvgPacketsReceivedPerSecond uint64 `json:"avg_packets_recv_ps"` // Average packets received per second
	AvgErrorInPerSecond         uint64 `json:"avg_errin"`           // Average errors while receiving per second
	AvgErrorOutPerSecond        uint64 `json:"avg_errout"`          // Average errors while sending per second
	AvgDropInPerSecond          uint64 `json:"avg_dropin"`          // Average incoming dropped packets per second
	AvgDropOutPerSecond         uint64 `json:"avg_dropout"`         // Average outgoing dropped packets per second
}

type resultLibvirtDiskio struct {
	Name                string  `json:"name"`
	ReadIopsPerSecond   uint64  // Number of read iops per second
	WriteIopsPerSecond  uint64  // Number of write iops per second
	TotalIopsPerSecond  uint64  // Number of read and write iops per second
	ReadBytesPerSecond  uint64  // Number of bytes read from disk per second
	WriteBytesPerSecond uint64  // Number of bytes written to disk per second
	ReadAvgSize         float64 // Average request size of reads in bytes
	WriteAvgSize        float64 // Average request size of writes in bytes
}

type resultLibvirtDomain struct {
	Name              string // Name of the Domain (VM)
	Uuid              string // UUID of the Domain
	GuestAgentRunning bool
	Memory            *virtMemory // Memory stats in bytes
	Interfaces        map[string]*resultLibvirtNetwork
	Diskio            map[string]*resultLibvirtDiskio
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckLibvirt) Run(ctx context.Context) (interface{}, error) {
	if c.lastNetstatResults == nil {
		c.lastNetstatResults = make(map[string]map[string]*lastNetstatResultsForDelta)
	}
	if c.lastDiskstatsResults == nil {
		c.lastDiskstatsResults = make(map[string]map[string]*lastDiskstatsResultsForDelta)
	}

	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// List active+inactive domains (VMs)
	var flags libvirt.ConnectListAllDomainsFlags = 0
	doms, err := conn.ListAllDomains(flags)
	if err != nil {
		return nil, err
	}

	libvirtResults := make(map[string]*resultLibvirtDomain)
	for _, dom := range doms {
		name, _ := dom.GetName()
		uuid, _ := dom.GetUUIDString()

		var state libvirt.DomainState
		state, _, _ = dom.GetState()

		isDomRunning := state == libvirt.DOMAIN_RUNNING

		result := &resultLibvirtDomain{
			Name: name,
			Uuid: uuid,
		}

		// Get XML dump of the VM ()
		var flags libvirt.DomainXMLFlags = libvirt.DOMAIN_XML_SECURE
		xmlStr, err := dom.GetXMLDesc(flags)
		if err != nil {
			//Todo log error
			continue
		}

		var xmlDomain XmlDomain
		err = xml.Unmarshal([]byte(xmlStr), &xmlDomain)
		if err != nil {
			//Todo log error
			continue
		}

		// Get memory info per VM
		memory := &virtMemory{}
		memStats, _ := dom.MemoryStats(uint32(libvirt.DOMAIN_MEMORY_STAT_LAST), 0)

		for _, stat := range memStats {
			switch int(stat.Tag) {
			case int(libvirt.DOMAIN_MEMORY_STAT_ACTUAL_BALLOON):
				memory.Total = stat.Val

			case int(libvirt.DOMAIN_MEMORY_STAT_UNUSED):
				memory.Ununsed = stat.Val

			case int(libvirt.DOMAIN_MEMORY_STAT_AVAILABLE):
				memory.Available = stat.Val

			case int(libvirt.DOMAIN_MEMORY_STAT_RSS):
				memory.Rss = stat.Val

			case int(libvirt.DOMAIN_MEMORY_STAT_SWAP_IN):
				memory.SwapIn = stat.Val

			case int(libvirt.DOMAIN_MEMORY_STAT_SWAP_OUT):
				memory.SwapOut = stat.Val

			case int(libvirt.DOMAIN_MEMORY_STAT_MINOR_FAULT):
				memory.MinorFault = stat.Val

			case int(libvirt.DOMAIN_MEMORY_STAT_MAJOR_FAULT):
				memory.MajorFault = stat.Val
			}
		}

		result.Memory = memory

		// Hash map to store counter values for deltas
		if _, uuidExists := c.lastNetstatResults[uuid]; !uuidExists {
			c.lastNetstatResults[uuid] = make(map[string]*lastNetstatResultsForDelta)
		}

		// Get network stats
		for _, iface := range xmlDomain.Devices.Interfaces {

			if isDomRunning {
				result.Interfaces = make(map[string]*resultLibvirtNetwork)
				ifaceStats, err := dom.InterfaceStats(iface.Device.Dev)
				if err == nil {

					// Get last result to calculate bytes/s packets/s
					if lastCheckResults, ok := c.lastNetstatResults[uuid][iface.Device.Dev]; ok {
						Interval := time.Now().Unix() - lastCheckResults.Timestamp
						RxBytesDiff := WrapDiffInt64(lastCheckResults.RxBytes, ifaceStats.RxBytes)
						RxPacketsDiff := WrapDiffInt64(lastCheckResults.RxPackets, ifaceStats.RxPackets)
						RxErrsDiff := WrapDiffInt64(lastCheckResults.RxErrs, ifaceStats.RxErrs)
						RxDropDiff := WrapDiffInt64(lastCheckResults.RxDrop, ifaceStats.RxDrop)
						TxBytesDiff := WrapDiffInt64(lastCheckResults.TxBytes, ifaceStats.TxBytes)
						TxPacketsDiff := WrapDiffInt64(lastCheckResults.TxPackets, ifaceStats.TxPackets)
						TxErrsDiff := WrapDiffInt64(lastCheckResults.TxErrs, ifaceStats.TxErrs)
						TxDropDiff := WrapDiffInt64(lastCheckResults.TxDrop, ifaceStats.TxDrop)

						result.Interfaces[iface.Device.Dev] = &resultLibvirtNetwork{
							Name:                        iface.Device.Dev,
							AvgBytesSentPerSecond:       uint64(safemaths.DivideInt64(TxBytesDiff, Interval)),
							AvgBytesReceivedPerSecond:   uint64(safemaths.DivideInt64(RxBytesDiff, Interval)),
							AvgPacketsSentPerSecond:     uint64(safemaths.DivideInt64(TxPacketsDiff, Interval)),
							AvgPacketsReceivedPerSecond: uint64(safemaths.DivideInt64(RxPacketsDiff, Interval)),
							AvgErrorInPerSecond:         uint64(safemaths.DivideInt64(RxErrsDiff, Interval)),
							AvgErrorOutPerSecond:        uint64(safemaths.DivideInt64(TxErrsDiff, Interval)),
							AvgDropInPerSecond:          uint64(safemaths.DivideInt64(RxDropDiff, Interval)),
							AvgDropOutPerSecond:         uint64(safemaths.DivideInt64(TxDropDiff, Interval)),
						}
					}

					// Store counter values for next check evaluation
					c.lastNetstatResults[uuid][iface.Device.Dev] = &lastNetstatResultsForDelta{
						Timestamp: time.Now().Unix(),
						RxBytes:   ifaceStats.RxBytes,
						RxPackets: ifaceStats.RxPackets,
						RxErrs:    ifaceStats.RxErrs,
						RxDrop:    ifaceStats.RxDrop,
						TxBytes:   ifaceStats.TxBytes,
						TxPackets: ifaceStats.TxPackets,
						TxErrs:    ifaceStats.TxErrs,
						TxDrop:    ifaceStats.TxDrop,
					}
				}
			} else {
				// Vm is not running - no traffic
				c.lastNetstatResults[uuid][iface.Device.Dev] = &lastNetstatResultsForDelta{
					Timestamp: time.Now().Unix(),
				}
			}
		}

		// Hash map to store counter values for deltas
		if _, uuidExists := c.lastDiskstatsResults[uuid]; !uuidExists {
			c.lastDiskstatsResults[uuid] = make(map[string]*lastDiskstatsResultsForDelta)
		}
		for _, disk := range xmlDomain.Devices.Disks {
			if isDomRunning {
				result.Diskio = make(map[string]*resultLibvirtDiskio)
				blockStats, err := dom.BlockStats(disk.Target.Dev)
				if err == nil {

					// Get last result to calculate bytes/s packets/s
					if lastCheckResults, ok := c.lastDiskstatsResults[uuid][disk.Target.Dev]; ok {
						Interval := time.Now().Unix() - lastCheckResults.Timestamp
						WrReqDiff := WrapDiffInt64(lastCheckResults.WrReq, blockStats.WrReq)
						RdReqDiff := WrapDiffInt64(lastCheckResults.RdReq, blockStats.RdReq)
						RdBytesDiff := WrapDiffInt64(lastCheckResults.RdBytes, blockStats.RdBytes)
						WrBytesDiff := WrapDiffInt64(lastCheckResults.WrBytes, blockStats.WrBytes)

						ReadIopsPerSecond := safemaths.DivideInt64(RdReqDiff, Interval)
						WriteIopsPerSecond := safemaths.DivideInt64(WrReqDiff, Interval)
						ReadBytesPerSecond := safemaths.DivideInt64(RdBytesDiff, Interval)
						WriteBytesPerSecond := safemaths.DivideInt64(WrBytesDiff, Interval)

						ReadAvgSize := safemaths.DivideFloat64(float64(ReadBytesPerSecond), float64(Interval))
						WriteAvgSize := safemaths.DivideFloat64(float64(WriteBytesPerSecond), float64(Interval))

						result.Diskio[disk.Target.Dev] = &resultLibvirtDiskio{
							Name:                disk.Target.Dev,
							ReadIopsPerSecond:   uint64(ReadIopsPerSecond),
							WriteIopsPerSecond:  uint64(WriteIopsPerSecond),
							TotalIopsPerSecond:  uint64(ReadIopsPerSecond + WriteIopsPerSecond),
							ReadBytesPerSecond:  uint64(ReadBytesPerSecond),
							WriteBytesPerSecond: uint64(WriteBytesPerSecond),
							ReadAvgSize:         ReadAvgSize,
							WriteAvgSize:        WriteAvgSize,
						}
					}

					// Store counter values for next check evaluation
					c.lastDiskstatsResults[uuid][disk.Target.Dev] = &lastDiskstatsResultsForDelta{
						Timestamp: time.Now().Unix(),
						WrReq:     blockStats.WrReq,
						RdReq:     blockStats.RdReq,
						RdBytes:   blockStats.RdBytes,
						WrBytes:   blockStats.WrBytes,
					}
				}
			} else {
				// Vm is not running - no traffic
				c.lastDiskstatsResults[uuid][disk.Target.Dev] = &lastDiskstatsResultsForDelta{
					Timestamp: time.Now().Unix(),
				}
			}
		}

		// Get CPU usage (total)
		// https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainGetCPUStats
		// https://github.com/virt-manager/virt-manager/blob/b17914591aeefedd50a0a0634f479222a7ff591c/virtManager/lib/statsmanager.py#L149-L190

		// Get CPU Time of host system
		nparams, err := dom.GetCPUStats(-1, 1, 0)
		if err != nil {
			fmt.Println(err)
		}

		nparamsJs, _ := json.Marshal(nparams)
		fmt.Println("++++begin+++")
		fmt.Println(string(nparamsJs))
		fmt.Println("+++end++++")

		// Get network ip addresses stats
		var src libvirt.DomainInterfaceAddressesSource = libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_ARP
		interfaces, err := dom.ListAllInterfaceAddresses(src)
		if err != nil {
			fmt.Println(err)
			continue
		}

		ifaceJs, _ := json.Marshal(interfaces)
		fmt.Println(string(ifaceJs))

		// Add current VM to results list
		libvirtResults[uuid] = result

		dom.Free()
	}

	return libvirtResults, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckLibvirt) Configure(config *config.Configuration) (bool, error) {
	return config.Libvirt, nil
}
