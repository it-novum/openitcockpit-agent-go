package checks

import (
	"github.com/it-novum/openitcockpit-agent-go/config"
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

// Configure the command or return false if the command was disabled
func (c *CheckNetIo) Configure(config *config.Configuration) (bool, error) {
	return config.NetIo, nil
}
