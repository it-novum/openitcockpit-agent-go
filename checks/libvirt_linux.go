// +build libvirt

// sudo apt-get install libvirt-dev

package checks

import (
	"context"
	"math"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/safemaths"
	libvirt "libvirt.org/libvirt-go"
)

// CheckLibvirt gathers information about running VMs
type CheckLibvirt struct {
	lastNetstatResults   map[string]map[string]*lastNetstatResultsForDelta
	lastDiskstatsResults map[string]map[string]*lastDiskstatsResultsForDelta
	lastCpuResults       map[string]*lastCpuResultsForDelta
}

type lastNetstatResultsForDelta struct {
	Timestamp int64
	RxBytes   uint64
	RxPackets uint64
	RxErrs    uint64
	RxDrop    uint64
	TxBytes   uint64
	TxPackets uint64
	TxErrs    uint64
	TxDrop    uint64
}

type lastDiskstatsResultsForDelta struct {
	Timestamp int64
	WrReq     uint64
	RdReq     uint64
	RdBytes   uint64
	WrBytes   uint64
}

type lastCpuResultsForDelta struct {
	Timestamp  int64
	CpuTime    uint64
	UserTime   uint64
	SystemTime uint64
	VcpuTime   uint64
}

// Name will be used in the response as check name
func (c *CheckLibvirt) Name() string {
	return "libvirt"
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
	Path                string  `json:"path/filepath"`
	ReadIopsPerSecond   uint64  // Number of read iops per second
	WriteIopsPerSecond  uint64  // Number of write iops per second
	TotalIopsPerSecond  uint64  // Number of read and write iops per second
	ReadBytesPerSecond  uint64  // Number of bytes read from disk per second
	WriteBytesPerSecond uint64  // Number of bytes written to disk per second
	ReadAvgSize         float64 // Average request size of reads in bytes
	WriteAvgSize        float64 // Average request size of writes in bytes
}

type resultLibvirtCpuUsage struct {
	HostPercent  float64 // CPU Usage in % of the Host
	GuestPercent float64 // CPU Usage in % of the Guest (VM itself)
}

type resultLibvirtDomain struct {
	Name              string // Name of the Domain (VM)
	Uuid              string // UUID of the Domain
	IsRunning         bool   // Is VM powered on
	GuestAgentRunning bool
	Memory            *virtMemory // Memory stats in bytes
	Interfaces        map[string]*resultLibvirtNetwork
	Diskio            map[string]*resultLibvirtDiskio
	CpuUsage          *resultLibvirtCpuUsage
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
	if c.lastCpuResults == nil {
		c.lastCpuResults = make(map[string]*lastCpuResultsForDelta)
	}

	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	nodeInfo, err := conn.GetNodeInfo()
	if err != nil {
		return nil, err
	}

	nodeCpuCount := nodeInfo.Cpus

	// List active+inactive domains (VMs)
	var flags libvirt.ConnectListAllDomainsFlags = 0
	doms, err := conn.ListAllDomains(flags)
	if err != nil {
		return nil, err
	}

	libvirtResults := make(map[string]*resultLibvirtDomain)
	if len(doms) == 0 {
		return libvirtResults, nil
	}

	for _, dom := range doms {
		var statsTypes = libvirt.DOMAIN_STATS_STATE |
			libvirt.DOMAIN_STATS_CPU_TOTAL |
			libvirt.DOMAIN_STATS_BALLOON |
			libvirt.DOMAIN_STATS_VCPU |
			libvirt.DOMAIN_STATS_INTERFACE |
			libvirt.DOMAIN_STATS_BLOCK |
			libvirt.DOMAIN_STATS_PERF |
			libvirt.DOMAIN_STATS_IOTHREAD |
			libvirt.DOMAIN_STATS_MEMORY

		domArr := []*libvirt.Domain{
			&dom,
		}
		// GetAllDomainStats is not as powerfull as it looks like
		domStatsArr, err := conn.GetAllDomainStats(domArr, statsTypes, 0)
		if err != nil {
			_ = domArr[0].Free()
			_ = dom.Free()
			continue
		}

		domStats := domStatsArr[0]
		name, _ := dom.GetName()
		uuid, _ := dom.GetUUIDString()
		isDomRunning := domStats.State.State == libvirt.DOMAIN_RUNNING

		result := &resultLibvirtDomain{
			Name:      name,
			Uuid:      uuid,
			IsRunning: isDomRunning,
		}

		if !isDomRunning {
			// Add current VM to results list
			libvirtResults[uuid] = result
			_ = domStats.Domain.Free()
			_ = domArr[0].Free()
			_ = dom.Free()
			continue
		}

		// Docs: https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainMemoryStats
		result.Memory = &virtMemory{
			Total:      domStats.Balloon.Current,
			Ununsed:    domStats.Balloon.Unused,
			Available:  domStats.Balloon.Available,
			Rss:        domStats.Balloon.Rss,
			SwapIn:     domStats.Balloon.SwapIn,
			SwapOut:    domStats.Balloon.SwapOut,
			MinorFault: domStats.Balloon.MinorFault,
			MajorFault: domStats.Balloon.MajorFault,
		}

		// Get net stats
		// Hash map to store counter values for deltas
		if _, uuidExists := c.lastNetstatResults[uuid]; !uuidExists {
			c.lastNetstatResults[uuid] = make(map[string]*lastNetstatResultsForDelta)
		}

		result.Interfaces = make(map[string]*resultLibvirtNetwork)
		for _, iface := range domStats.Net {
			// Get last result to calculate bytes/s packets/s

			if lastCheckResults, ok := c.lastNetstatResults[uuid][iface.Name]; ok {
				Interval := time.Now().Unix() - lastCheckResults.Timestamp
				RxBytesDiff := WrapDiffUint64(lastCheckResults.RxBytes, iface.RxBytes)
				RxPacketsDiff := WrapDiffUint64(lastCheckResults.RxPackets, iface.RxPkts)
				RxErrsDiff := WrapDiffUint64(lastCheckResults.RxErrs, iface.RxErrs)
				RxDropDiff := WrapDiffUint64(lastCheckResults.RxDrop, iface.RxDrop)
				TxBytesDiff := WrapDiffUint64(lastCheckResults.TxBytes, iface.TxBytes)
				TxPacketsDiff := WrapDiffUint64(lastCheckResults.TxPackets, iface.TxPkts)
				TxErrsDiff := WrapDiffUint64(lastCheckResults.TxErrs, iface.TxErrs)
				TxDropDiff := WrapDiffUint64(lastCheckResults.TxDrop, iface.TxDrop)

				result.Interfaces[iface.Name] = &resultLibvirtNetwork{
					Name:                        iface.Name,
					AvgBytesSentPerSecond:       safemaths.DivideUint64(TxBytesDiff, uint64(Interval)),
					AvgBytesReceivedPerSecond:   safemaths.DivideUint64(RxBytesDiff, uint64(Interval)),
					AvgPacketsSentPerSecond:     safemaths.DivideUint64(TxPacketsDiff, uint64(Interval)),
					AvgPacketsReceivedPerSecond: safemaths.DivideUint64(RxPacketsDiff, uint64(Interval)),
					AvgErrorInPerSecond:         safemaths.DivideUint64(RxErrsDiff, uint64(Interval)),
					AvgErrorOutPerSecond:        safemaths.DivideUint64(TxErrsDiff, uint64(Interval)),
					AvgDropInPerSecond:          safemaths.DivideUint64(RxDropDiff, uint64(Interval)),
					AvgDropOutPerSecond:         safemaths.DivideUint64(TxDropDiff, uint64(Interval)),
				}
			}

			// Store counter values for next check evaluation
			c.lastNetstatResults[uuid][iface.Name] = &lastNetstatResultsForDelta{
				Timestamp: time.Now().Unix(),
				RxBytes:   iface.RxBytes,
				RxPackets: iface.RxPkts,
				RxErrs:    iface.RxErrs,
				RxDrop:    iface.RxDrop,
				TxBytes:   iface.TxBytes,
				TxPackets: iface.TxPkts,
				TxErrs:    iface.TxErrs,
				TxDrop:    iface.TxDrop,
			}
		}

		// Get disk io of block devices
		// Hash map to store counter values for deltas
		if _, uuidExists := c.lastDiskstatsResults[uuid]; !uuidExists {
			c.lastDiskstatsResults[uuid] = make(map[string]*lastDiskstatsResultsForDelta)
		}

		result.Diskio = make(map[string]*resultLibvirtDiskio)
		for _, block := range domStats.Block {
			if lastCheckResults, ok := c.lastDiskstatsResults[uuid][block.Name]; ok {
				Interval := time.Now().Unix() - lastCheckResults.Timestamp
				WrReqDiff := WrapDiffUint64(lastCheckResults.WrReq, block.WrReqs)
				RdReqDiff := WrapDiffUint64(lastCheckResults.RdReq, block.RdReqs)
				RdBytesDiff := WrapDiffUint64(lastCheckResults.RdBytes, block.RdBytes)
				WrBytesDiff := WrapDiffUint64(lastCheckResults.WrBytes, block.WrBytes)

				ReadIopsPerSecond := safemaths.DivideUint64(RdReqDiff, uint64(Interval))
				WriteIopsPerSecond := safemaths.DivideUint64(WrReqDiff, uint64(Interval))
				ReadBytesPerSecond := safemaths.DivideUint64(RdBytesDiff, uint64(Interval))
				WriteBytesPerSecond := safemaths.DivideUint64(WrBytesDiff, uint64(Interval))

				ReadAvgSize := safemaths.DivideFloat64(float64(ReadBytesPerSecond), float64(Interval))
				WriteAvgSize := safemaths.DivideFloat64(float64(WriteBytesPerSecond), float64(Interval))

				result.Diskio[block.Name] = &resultLibvirtDiskio{
					Name:                block.Name,
					Path:                block.Path,
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
			c.lastDiskstatsResults[uuid][block.Name] = &lastDiskstatsResultsForDelta{
				Timestamp: time.Now().Unix(),
				WrReq:     block.WrReqs,
				RdReq:     block.RdReqs,
				RdBytes:   block.RdBytes,
				WrBytes:   block.WrBytes,
			}
		}

		// Get cpu usage
		var vCpuTimeTotal uint64 = 0
		for _, vCpu := range domStats.Vcpu {
			vCpuTimeTotal = vCpuTimeTotal + vCpu.Time
		}
		if lastCheckResults, ok := c.lastCpuResults[uuid]; ok {
			Interval := time.Now().Unix() - lastCheckResults.Timestamp
			CpuTimeDiff := WrapDiffUint64(lastCheckResults.CpuTime, domStats.Cpu.Time)

			// CpuTime is in nanoseconds - convert interval from seconds to nanoseconds
			// Credit to:
			// https://github.com/virt-manager/virt-manager/blob/b17914591aeefedd50a0a0634f479222a7ff591c/virtManager/lib/statsmanager.py#L149-L190
			percentage_base := (float64(CpuTimeDiff) * 100.0) / (float64(Interval) * 1000.0 * 1000.0 * 1000.0)

			cpuHostPercent := percentage_base / float64(nodeCpuCount)

			guestcpus := len(domStats.Vcpu)
			cpuGuestPercent := safemaths.DivideFloat64(float64(percentage_base), float64(guestcpus))

			cpuHostPercent = math.Max(0.0, math.Min(100.0, cpuHostPercent))
			cpuGuestPercent = math.Max(0.0, math.Min(100.0, cpuGuestPercent))

			result.CpuUsage = &resultLibvirtCpuUsage{
				HostPercent:  cpuHostPercent,
				GuestPercent: cpuGuestPercent,
			}
		}

		// Store counter values for next check evaluation
		c.lastCpuResults[uuid] = &lastCpuResultsForDelta{
			Timestamp:  time.Now().Unix(),
			CpuTime:    domStats.Cpu.Time,
			UserTime:   domStats.Cpu.User,
			SystemTime: domStats.Cpu.System,
			VcpuTime:   vCpuTimeTotal,
		}

		// Add current VM to results list
		libvirtResults[uuid] = result

		_ = domStats.Domain.Free()
		_ = domArr[0].Free()
		_ = dom.Free()
	}

	return libvirtResults, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckLibvirt) Configure(config *config.Configuration) (bool, error) {
	return config.Libvirt, nil
}
