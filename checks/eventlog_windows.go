package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/utils"
)

/*
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
}*/

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

type resultEvent struct {
	MachineName    string
	Category       string
	CategoryNumber int64
	EventID        int64
	EntryType      int64
	Message        string
	Source         string
	TimeGenerated  int64
	TimeWritten    int64
	Index          int64
}

type JsonEventLog struct {
	MachineName    string
	Category       string
	CategoryNumber int64
	EventID        int64
	EntryType      int64
	Message        string
	Source         string
	TimeGenerated  string
	TimeWritten    string
	Index          int64
}

type CheckWindowsEventLog struct {
	age      time.Duration
	logfiles []string
}

type EventlogCheckOptions struct {
	Age      time.Duration
	Logfiles []string
}

// Name will be used in the response as check name
func (c *CheckWindowsEventLog) Name() string {
	return "windows_eventlog"
}

// This check is a Memory Leak as a Service
// See: https://github.com/StackExchange/wmi/issues/55
// https://github.com/go-ole/go-ole/issues/135
//
// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckWindowsEventLog) Run(ctx context.Context) (interface{}, error) {

	now := time.Now().UTC()
	//now = now.Add((3600 * time.Second) * -1)
	now = now.Add(c.age * -1)

	// Golang date formate: https://golang.org/src/time/format.go
	datetime := now.Format("2006-01-02T15:04:05")

	//var eventBuffer map[string][]*Win32_NTLogEvent
	eventBuffer := make(map[string][]*resultEvent)
	for _, logfile := range c.logfiles {
		timeout := time.Duration(30 * time.Second)

		// Unix timestamp with timezone :/
		//cmd := fmt.Sprintf("Get-EventLog -LogName %s -After %s | Select-Object MachineName, Category, CategoryNumber, EventID, EntryType, Message, Source, @{n='TimeGenerated';e={Get-Date ($_.timegenerated) -UFormat %%s }}, @{n='TimeWritten';e={Get-Date ($_.timegenerated) -UFormat %%s }}, Index | ConvertTo-Json -depth 100", logfile, datetime)

		// Date as ISO-8601
		// Format: https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.utility/get-date?view=powershell-7.1#notes
		cmd := fmt.Sprintf("[Console]::OutputEncoding = [Text.UTF8Encoding]::UTF8\r\nGet-EventLog -LogName %s -After %s | Select-Object MachineName, Category, CategoryNumber, EventID, EntryType, Message, Source, @{n='TimeGenerated';e={Get-Date ($_.timegenerated) -UFormat %%Y-%%m-%%dT%%H:%%M:%%S%%Z }}, @{n='TimeWritten';e={Get-Date ($_.timegenerated) -UFormat %%Y-%%m-%%dT%%H:%%M:%%S%%Z }}, Index | ConvertTo-Json -depth 100", logfile, datetime)
		log.Errorln("Cmd: ", cmd)
		commandResult, err := utils.RunCommand(ctx, utils.CommandArgs{
			Timeout: timeout,
			Command: cmd,
			Shell:   "powershell_command",
		})

		if err != nil {
			log.Errorln("Event Log Error: ", commandResult.Stdout)
			continue
		}

		if commandResult.RC > 0 {
			if commandResult.Stdout != "" {
				// Otherwise the event log is maybe just empty
				log.Errorln("Event Log Error: ", commandResult.Stdout)
			}

			// Add empty array to result
			eventBuffer[logfile] = make([]*resultEvent, 0)
			continue
		}

		/*
			If only one record gets returned powershell will turn this into an object instead of an array ob objects
			PowerShell -AsArray is not supported on my Windows 10
			{
				"MachineName":  "DESKTOP-BCBF1TR",
				"Category":  "(0)",
				"CategoryNumber":  0,
				"EventID":  1,
				"EntryType":  1,
				"Message":  "My first log",
				"Source":  "MYEVENTSOURCE",
				"TimeGenerated":  "2021-06-01T09:48:41+02",
				"TimeWritten":  "2021-06-01T09:48:41+02",
				"Index":  1
			}
		*/

		var dst []*JsonEventLog
		var jsonError error
		if len(strings.TrimRight(commandResult.Stdout, "\r\n")) > 0 {
			firstCharacter := commandResult.Stdout[0:1]
			if firstCharacter == "{" {
				// Only one event log record
				var singleRecord *JsonEventLog
				jsonError = json.Unmarshal([]byte(commandResult.Stdout), &singleRecord)
				if jsonError == nil {
					dst = []*JsonEventLog{
						singleRecord,
					}
				}
			} else {
				// Array of event log records
				jsonError = json.Unmarshal([]byte(commandResult.Stdout), &dst)
			}
		}

		if jsonError != nil {
			return nil, jsonError
		}

		for _, event := range dst {
			// Resolve Memory Leak

			TimeGenerated, _ := time.Parse("2006-01-02T15:04:05-07", event.TimeGenerated)
			TimeWritten, _ := time.Parse("2006-01-02T15:04:05-07", event.TimeWritten)

			eventBuffer[logfile] = append(eventBuffer[logfile], &resultEvent{
				MachineName:    event.MachineName,
				Category:       event.Category,
				CategoryNumber: event.CategoryNumber,
				EventID:        event.EventID,
				EntryType:      event.EntryType,
				Message:        event.Message,
				Source:         event.Source,
				TimeGenerated:  TimeGenerated.Unix(),
				TimeWritten:    TimeWritten.Unix(),
				Index:          event.Index,
			})
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
