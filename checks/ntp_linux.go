package checks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// Massive shout out and thanks to the Prometheus Community
// https://github.com/prometheus/node_exporter/blob/master/collector/timex.go
// Also "man adjtimex" provides a lot of details

const (

	// Clock is not synchronized to a time server (TIME_ERROR)
	NTP_TIME_ERROR = 5

	// timex.Status time resolution bit (STA_NANO),
	// If the STA_NANO is set, all values are in nanoseconds, otherwise in microseconds
	// resolution (0 = us, 1 = ns)
	// Source: https://github.com/torvalds/linux/blob/7e90b5c295ec1e47c8ad865429f046970c549a66/include/uapi/linux/timex.h#L187
	STA_NANO = 0x2000

	// Convertions
	NANOSECONDS_TO_SECONDS  = 1000000000
	MICROSECONDS_TO_SECONDS = 1000000
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckNtp) Run(ctx context.Context) (interface{}, error) {
	var divisor float64

	var timex = new(unix.Timex)
	status, err := unix.Adjtimex(timex)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return nil, fmt.Errorf("Permission denied for timex checks %v", err)
		}
		return nil, err
	}

	syncStatus := true
	if status == NTP_TIME_ERROR {
		syncStatus = false
	}

	divisor = MICROSECONDS_TO_SECONDS
	if timex.Status&STA_NANO != 0 {
		divisor = NANOSECONDS_TO_SECONDS
	}

	result := &resultNtp{
		Timestamp:  time.Now().Unix(),
		SyncStatus: syncStatus,
		Offset:     float64(timex.Offset) / divisor,
	}

	return result, nil

}
