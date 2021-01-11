package checks

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v3/net"
)

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
			BytesRecv, _ := Wrapdiff(float64(lastCheckResults.BytesReceived), float64(nic.BytesRecv))
			BytesSent, _ := Wrapdiff(float64(lastCheckResults.BytesSent), float64(nic.BytesSent))
			PacketsSent, _ := Wrapdiff(float64(lastCheckResults.PacketsSent), float64(nic.PacketsSent))
			PacketsRecv, _ := Wrapdiff(float64(lastCheckResults.PacketsReceived), float64(nic.PacketsRecv))
			ErrorIn, _ := Wrapdiff(float64(lastCheckResults.ErrorIn), float64(nic.Errin))
			ErrorOut, _ := Wrapdiff(float64(lastCheckResults.ErrorOut), float64(nic.Errout))
			DropIn, _ := Wrapdiff(float64(lastCheckResults.DropIn), float64(nic.Dropin))
			DropOut, _ := Wrapdiff(float64(lastCheckResults.DropOut), float64(nic.Dropout))
			Timestamp, _ := Wrapdiff(float64(lastCheckResults.Timestamp), float64(time.Now().Unix()))

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
				AvgBytesSentPerSecond:       uint64(BytesSent / Timestamp),
				AvgBytesReceivedPerSecond:   uint64(BytesRecv / Timestamp),
				AvgPacketsSentPerSecond:     uint64(PacketsSent / Timestamp),
				AvgPacketsReceivedPerSecond: uint64(PacketsRecv / Timestamp),
				AvgErrorInPerSecond:         uint64(ErrorIn / Timestamp),
				AvgErrorOutPerSecond:        uint64(ErrorOut / Timestamp),
				AvgDropInPerSecond:          uint64(DropIn / Timestamp),
				AvgDropOutPerSecond:         uint64(DropOut / Timestamp),
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
