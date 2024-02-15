package checks

import (
	"context"
	"time"
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckNtp) Run(ctx context.Context) (interface{}, error) {

	// macOS does not support timex like Linux.
	// The NTP service has changed seferale times over the last macOS versions. https://apple.stackexchange.com/a/117865
	// While it is possible to query an offset, it requires admin privileges and you have to parse the command line output
	// Please see for more information https://openitcockpit.atlassian.net/browse/OA-52
	//
	// For macOS Catalina (and newer) you can query the configured NTP server via
	// sudo systemsetup -getnetworktimeserver
	// and query the ntp server using
	// sntp -d time.apple.com
	// How ever, the output is not very parser friendly.
	// I found this commands in https://github.com/ConSol-Monitoring/snclient/blob/706454ce37b860ac636c79cb9cebbf0eae65fd14/pkg/snclient/check_ntp_offset.go#L426-L448C86 and ChatGPT
	//
	// For now, we send a unixtimestamp in microseconds to the openITCOCKPIT server and compare those two clocks

	result := &resultNtp{
		Timestamp:      time.Now().Unix(),
		TimestampMicro: time.Now().UnixMicro(),
		SyncStatus:     false,
		Offset:         0,
	}

	return result, nil

}
