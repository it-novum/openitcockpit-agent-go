package checks

import (
	"context"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/it-novum/openitcockpit-agent-go/utils"
	"github.com/prometheus-community/windows_exporter/perflib"
)

// Massive shout out and thanks to the Prometheus Community
// Credit to:
// https://github.com/prometheus-community/windows_exporter/blob/b5284aca85433c097fdbd64671b4c6dcaff10037/pkg/collector/time/time.go#L112-L119
// https://github.com/prometheus-community/windows_exporter/issues/532
// nolint:underscore
type Perf_WindowsTimeService struct {
	//ClockFrequencyAdjustmentPPBTotal float64 `perflib:"Clock Frequency Adjustment (ppb)"` // On my Windows 10 System the counter name is "Clock Frequency Adjustment (PPB)" Total adjustment made to the local system clock frequency by W32Time in Parts Per Billion (PPB) units.
	ComputedTimeOffset              float64 `perflib:"Computed Time Offset"`          // bsolute time offset between the system clock and the chosen time source in microseconds
	NTPClientTimeSourceCount        float64 `perflib:"NTP Client Time Source Count"`  // Active number of NTP Time sources being used by the client
	NTPRoundtripDelay               float64 `perflib:"NTP Roundtrip Delay"`           // Roundtrip delay experienced by the NTP client in receiving a response from the server for the most recent request, in microseconds
	NTPServerIncomingRequestsTotal  float64 `perflib:"NTP Server Incoming Requests"`  // Total number of requests received by NTP server
	NTPServerOutgoingResponsesTotal float64 `perflib:"NTP Server Outgoing Responses"` // Total number of requests responded to by NTP server
}

// Prometheus does a lot of magic here but basically:
//  1. The collector.go ask the collector for the performance counter (this is a string)
//     https://github.com/prometheus-community/windows_exporter/blob/b5284aca85433c097fdbd64671b4c6dcaff10037/pkg/collector/collector.go#L170
//
// 1.2) "Windows Time Service" in this case https://github.com/prometheus-community/windows_exporter/blob/b5284aca85433c097fdbd64671b4c6dcaff10037/pkg/collector/time/time.go#L54
// 2) The collector now calls "perflib.MapCounterToIndex" which basically "Counter 009" hardcoded
// https://github.com/prometheus-community/windows_exporter/blob/b5284aca85433c097fdbd64671b4c6dcaff10037/pkg/perflib/nametable.go#L15-L43
var (
	perflibTimeQuery = strconv.FormatUint(uint64(perflib.QueryNameTable("Counter 009").LookupIndex("Windows Time Service")), 10)
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckNtp) Run(ctx context.Context) (interface{}, error) {

	objects, err := perflib.QueryPerformanceData(perflibTimeQuery)
	if err != nil {
		return nil, err
	}
	for _, obj := range objects {
		if obj.Name != "Windows Time Service" {
			continue
		}

		var dst []Perf_WindowsTimeService
		err = utils.UnmarshalObject(obj, &dst)
		if err != nil {
			log.Errorln("Check NTP: could not query perflib: ", err)
			continue
		}

		syncStatus := false
		if dst[0].NTPClientTimeSourceCount > 0 {
			syncStatus = true
		}

		return &resultNtp{
			Timestamp:      time.Now().Unix(),
			TimestampMicro: time.Now().UnixMicro(),
			Offset:         float64(dst[0].ComputedTimeOffset / 1000000), // Converts microseconds into seconds
			SyncStatus:     syncStatus,
		}, nil

	}

	return &resultNtp{
		Timestamp:      time.Now().Unix(),
		TimestampMicro: time.Now().UnixMicro(),
		Offset:         0,
		SyncStatus:     false,
	}, nil

}
