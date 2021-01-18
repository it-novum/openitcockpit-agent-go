package checks

import (
	"context"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/shirou/gopsutil/v3/net"
)

// CheckNetIo gathers information about system network interface IO (net_io in the Python version)
type CheckNetIo struct {
	lastResults map[string]*resultNetIo
}

// Name will be used in the response as check name
func (c *CheckNetIo) Name() string {
	return "net_io"
}

type resultNetIo struct {
	Name                        string `json:"name"`                // Name of the network interface
	Timestamp                   int64  `json:"timestamp"`           // Timestamp of the last check evaluation
	BytesSent                   uint64 `json:"bytes_sent"`          // Number of bytes sent
	BytesReceived               uint64 `json:"bytes_recv"`          // Number of bytes received
	PacketsSent                 uint64 `json:"packets_sent"`        // Number of packets sent
	PacketsReceived             uint64 `json:"packets_recv"`        // Number of bytes received
	ErrorIn                     uint64 `json:"errin"`               // Total number of errors while receiving
	ErrorOut                    uint64 `json:"errout"`              // Total number of errors while sending
	DropIn                      uint64 `json:"dropin"`              // Total number of incoming packets which were dropped
	DropOut                     uint64 `json:"dropout"`             // Total number of outgoing packets which were dropped (always 0 on macOS and BSD)
	AvgBytesSentPerSecond       uint64 `json:"avg_bytes_sent_ps"`   // Average bytes sent per second
	AvgBytesReceivedPerSecond   uint64 `json:"avg_bytes_recv_ps"`   // Average bytes received per second
	AvgPacketsSentPerSecond     uint64 `json:"avg_packets_sent_ps"` // Average packets sent per second
	AvgPacketsReceivedPerSecond uint64 `json:"avg_packets_recv_ps"` // Average packets received per second
	AvgErrorInPerSecond         uint64 `json:"avg_errin"`           // Average errors while receiving per second
	AvgErrorOutPerSecond        uint64 `json:"avg_errout"`          // Average errors while sending per second
	AvgDropInPerSecond          uint64 `json:"avg_dropin"`          // Average incoming dropped packets per second
	AvgDropOutPerSecond         uint64 `json:"avg_dropout"`         // Average outgoing dropped packets per second
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckNetIo) Run(ctx context.Context) (interface{}, error) {
	stats, err := net.IOCountersWithContext(ctx, true)

	if err != nil {
		return nil, err
	}

	netResults := make(map[string]*resultNetIo)

	for _, nic := range stats {

		if lastCheckResults, ok := c.lastResults[nic.Name]; ok {
			BytesRecv := WrapDiffUint64(lastCheckResults.BytesReceived, nic.BytesRecv)
			BytesSent := WrapDiffUint64(lastCheckResults.BytesSent, nic.BytesSent)
			PacketsSent := WrapDiffUint64(lastCheckResults.PacketsSent, nic.PacketsSent)
			PacketsRecv := WrapDiffUint64(lastCheckResults.PacketsReceived, nic.PacketsRecv)
			ErrorIn := WrapDiffUint64(lastCheckResults.ErrorIn, nic.Errin)
			ErrorOut := WrapDiffUint64(lastCheckResults.ErrorOut, nic.Errout)
			DropIn := WrapDiffUint64(lastCheckResults.DropIn, nic.Dropin)
			DropOut := WrapDiffUint64(lastCheckResults.DropOut, nic.Dropout)
			Interval := uint64(time.Now().Unix() - lastCheckResults.Timestamp)

			// Just in case this this has the same bug as Python psutil has^^
			netResults[nic.Name] = &resultNetIo{
				Name:                        nic.Name,
				Timestamp:                   time.Now().Unix(),
				BytesSent:                   nic.BytesSent,
				BytesReceived:               nic.BytesRecv,
				PacketsSent:                 nic.PacketsSent,
				PacketsReceived:             nic.PacketsRecv,
				ErrorIn:                     nic.Errin,
				ErrorOut:                    nic.Errout,
				DropIn:                      nic.Dropin,
				DropOut:                     nic.Dropout,
				AvgBytesSentPerSecond:       uint64(BytesSent / Interval),
				AvgBytesReceivedPerSecond:   uint64(BytesRecv / Interval),
				AvgPacketsSentPerSecond:     uint64(PacketsSent / Interval),
				AvgPacketsReceivedPerSecond: uint64(PacketsRecv / Interval),
				AvgErrorInPerSecond:         uint64(ErrorIn / Interval),
				AvgErrorOutPerSecond:        uint64(ErrorOut / Interval),
				AvgDropInPerSecond:          uint64(DropIn / Interval),
				AvgDropOutPerSecond:         uint64(DropOut / Interval),
			}

		} else {
			//No previous check results for calculations... wait until check runs again
			//Store result for next check run
			netResults[nic.Name] = &resultNetIo{
				Name:            nic.Name,
				Timestamp:       time.Now().Unix(),
				BytesSent:       nic.BytesSent,
				BytesReceived:   nic.BytesRecv,
				PacketsSent:     nic.PacketsSent,
				PacketsReceived: nic.PacketsRecv,
				ErrorIn:         nic.Errin,
				ErrorOut:        nic.Errout,
				DropIn:          nic.Dropin,
				DropOut:         nic.Dropout,
			}
		}

	}

	c.lastResults = netResults
	return netResults, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckNetIo) Configure(config *config.Configuration) (bool, error) {
	return true, nil
}
