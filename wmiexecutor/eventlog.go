package wmiexecutor

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/it-novum/openitcockpit-agent-go/checks"
)

// CheckWindowsEventLog gathers information from the Windows Event Log
type CheckWindowsEventLog struct {
	verbose bool
	debug   bool
}

func (c *CheckWindowsEventLog) Configure(conf *Configuration) error {
	c.verbose = conf.verbose
	c.debug = conf.debug

	return nil
}

// Query WMI
// if error != nil the check result will be nil
func (c *CheckWindowsEventLog) RunQuery() (string, error) {

	// Read check options as JSON from stdin
	var options checks.EventlogCheckOptions
	err := json.NewDecoder(os.Stdin).Decode(&options)
	if err != nil {
		return "", err
	}

	age := time.Second * time.Duration(options.Age)
	now := time.Now().UTC()
	//now = now.Add((3600 * time.Second) * -1)
	now = now.Add(age * -1)

	// Get DMTF-DateTime for WMI.
	// PowerShell Example to generate this:
	// PS C:\Users\Administrator> $WMIDATEAGE = [System.Management.ManagementDateTimeConverter]::ToDmtfDateTime([DateTime]::UtcNow.AddDays(-14))
	// PS C:\Users\Administrator> echo $WMIDATEAGE
	// 20210308075246.091047+000
	// Docs: https://blogs.iis.net/bobbyv/working-with-wmi-dates-and-times
	//
	// Golang date formate: https://golang.org/src/time/format.go
	wmidate := now.Format("20060102150405.000000-070")

	var dst []checks.Win32_NTLogEvent
	var sql string
	//var eventBuffer map[string][]*Win32_NTLogEvent
	eventBuffer := make(map[string][]*checks.Win32_NTLogEvent)
	for _, logfile := range options.Logfiles {
		sql = fmt.Sprintf("SELECT * FROM Win32_NTLogEvent WHERE Logfile='%v' AND TimeWritten >= '%v'", logfile, wmidate)
		//fmt.Println(sql)

		err := wmi.Query(sql, &dst)
		if err != nil {
			//log.Errorln("Event Log: ", err)
			continue
		}

		for _, event := range dst {
			// Resolve Memory Leak
			eventBuffer[logfile] = append(eventBuffer[logfile], &checks.Win32_NTLogEvent{
				Category:         event.Category,
				CategoryString:   event.CategoryString,
				ComputerName:     event.ComputerName,
				Data:             event.Data,
				EventCode:        event.EventCode,
				EventIdentifier:  event.EventIdentifier,
				EventType:        event.EventType,
				InsertionStrings: event.InsertionStrings,
				Logfile:          event.Logfile,
				Message:          event.Message,
				RecordNumber:     event.RecordNumber,
				SourceName:       event.SourceName,
				TimeGenerated:    event.TimeGenerated,
				TimeWritten:      event.TimeWritten,
				Type:             event.Type,
				User:             event.User,
			})
		}
	}

	js, err := json.Marshal(eventBuffer)
	if err != nil {
		return "", err
	}

	return string(js), nil
}
