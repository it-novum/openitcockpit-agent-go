//+build linux darwin

package checks

import (
	"context"
	"github.com/it-novum/openitcockpit-agent-go/safemaths"
	"github.com/shirou/gopsutil/v3/net"
	log "github.com/sirupsen/logrus"
	"time"
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
			BytesRecv := WrapDiffUint64(lastCheckResults.BytesReceived, nic.BytesRecv)
			BytesSent := WrapDiffUint64(lastCheckResults.BytesSent, nic.BytesSent)
			PacketsSent := WrapDiffUint64(lastCheckResults.PacketsSent, nic.PacketsSent)
			PacketsRecv := WrapDiffUint64(lastCheckResults.PacketsReceived, nic.PacketsRecv)
			ErrorIn := WrapDiffUint64(lastCheckResults.ErrorIn, nic.Errin)
			ErrorOut := WrapDiffUint64(lastCheckResults.ErrorOut, nic.Errout)
			DropIn := WrapDiffUint64(lastCheckResults.DropIn, nic.Dropin)
			DropOut := WrapDiffUint64(lastCheckResults.DropOut, nic.Dropout)
			Interval := uint64(time.Now().Unix() - lastCheckResults.Timestamp)

			// prevent divide by zero
			if Interval == 0 {
				log.Errorln("NetIO: Interval == 0")
				return c.lastResults, nil
			}

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
				AvgBytesSentPerSecond:       safemaths.DivideUint64(BytesSent, Interval),
				AvgBytesReceivedPerSecond:   safemaths.DivideUint64(BytesRecv, Interval),
				AvgPacketsSentPerSecond:     safemaths.DivideUint64(PacketsSent, Interval),
				AvgPacketsReceivedPerSecond: safemaths.DivideUint64(PacketsRecv, Interval),
				AvgErrorInPerSecond:         safemaths.DivideUint64(ErrorIn, Interval),
				AvgErrorOutPerSecond:        safemaths.DivideUint64(ErrorOut, Interval),
				AvgDropInPerSecond:          safemaths.DivideUint64(DropIn, Interval),
				AvgDropOutPerSecond:         safemaths.DivideUint64(DropOut, Interval),
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
