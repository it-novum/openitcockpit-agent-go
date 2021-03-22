package checks

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/StackExchange/wmi"
	"github.com/it-novum/openitcockpit-agent-go/config"
)

type resultEvent struct {
	Message     string
	Channel     string
	Level       string
	LevelRaw    uint8
	RecordID    uint64
	TimeCreated time.Time
	Provider    string
	Task        string
	Keywords    []string
}

// https://docs.microsoft.com/en-us/previous-versions/windows/desktop/eventlogprov/win32-ntlogevent
type Win32_NTLogEvent struct {
	Category         uint16
	CategoryString   string
	ComputerName     string
	Data             []uint8
	EventCode        uint16
	EventIdentifier  uint32
	EventType        uint8
	InsertionStrings []string
	Logfile          string
	Message          string
	RecordNumber     uint32
	SourceName       string
	TimeGenerated    time.Time
	TimeWritten      time.Time
	Type             string
	User             string
}

type CheckWindowsEventLog struct {
	age      time.Duration
	logfiles []string
}

// Name will be used in the response as check name
func (c *CheckWindowsEventLog) Name() string {
	return "windows_eventlog"
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckWindowsEventLog) Run(ctx context.Context) (interface{}, error) {

	now := time.Now().UTC()
	//now = now.Add((3600 * time.Second) * -1)
	now = now.Add(c.age * -1)

	// Get DMTF-DateTime for WMI.
	// PowerShell Example to generate this:
	// PS C:\Users\Administrator> $WMIDATEAGE = [System.Management.ManagementDateTimeConverter]::ToDmtfDateTime([DateTime]::UtcNow.AddDays(-14))
	// PS C:\Users\Administrator> echo $WMIDATEAGE
	// 20210308075246.091047+000
	// Docs: https://blogs.iis.net/bobbyv/working-with-wmi-dates-and-times
	//
	// Golang date formate: https://golang.org/src/time/format.go
	wmidate := now.Format("20060102150405.000000-070")

	var dst []Win32_NTLogEvent
	var sql string
	//var eventBuffer map[string][]*Win32_NTLogEvent
	eventBuffer := make(map[string][]*Win32_NTLogEvent)
	for _, logfile := range c.logfiles {
		sql = fmt.Sprintf("SELECT * FROM Win32_NTLogEvent WHERE Logfile='%v' AND TimeWritten >= '%v'", logfile, wmidate)
		//fmt.Println(sql)

		err := wmi.Query(sql, &dst)
		if err != nil {
			log.Errorln("Event Log: ", err)
			continue
		}

		for i, _ := range dst {
			eventBuffer[logfile] = append(eventBuffer[logfile], &dst[i])
		}
	}

	return eventBuffer, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckWindowsEventLog) Configure(cfg *config.Configuration) (bool, error) {
	if cfg.WindowsEventLog && len(cfg.WindowsEventLogTypes) > 0 {
		c.logfiles = cfg.WindowsEventLogTypes

		// Get the events from the last hour
		var ageSec uint64 = 3600
		if cfg.WindowsEventLogAge > 0 {
			ageSec = uint64(cfg.WindowsEventLogAge)
		}

		c.age = time.Second * time.Duration(ageSec)

		return true, nil
	}
	return false, nil
}
