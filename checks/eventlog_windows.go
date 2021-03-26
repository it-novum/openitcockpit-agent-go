package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/utils"
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
	age             uint64
	logfiles        []string
	WmiExecutorPath string // Used for Windows
}

type EventlogCheckOptions struct {
	Age      uint64
	Logfiles []string
}

// Name will be used in the response as check name
func (c *CheckWindowsEventLog) Name() string {
	return "windows_eventlog"
}

// Unfortunately the WMI library is suffering from a memory leak
// especially on windows Server 2016 and Windows 10.
// For this reason all WMI queries have been moved to an external binary (fork -> exec) to avoid any memory issues.
//
// Hopefully the memory issues will be fixed one day.
// This check used to look like this: https://github.com/it-novum/openitcockpit-agent-go/blob/a8ec01146e419a2db246844ca95cbe4ea560d9e6/checks/eventlog_windows.go

// This check is a Memory Leak as a Service
// See: https://github.com/StackExchange/wmi/issues/55
// https://github.com/go-ole/go-ole/issues/135
//
// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckWindowsEventLog) Run(ctx context.Context) (interface{}, error) {
	// exec wmiexecutor.exe to avoid memory leak

	options, _ := json.Marshal(&EventlogCheckOptions{
		Age:      c.age,
		Logfiles: c.logfiles,
	})

	timeout := 10 * time.Second
	commandResult, err := utils.RunCommand(ctx, utils.CommandArgs{
		Command: c.WmiExecutorPath + " --command eventlog",
		Shell:   "",
		Timeout: timeout,
		Env: []string{
			"OITC_AGENT_WMI_EXECUTOR=1",
		},
		Stdin: string(options) + "\n",
	})

	if err != nil {
		return nil, err
	}

	if commandResult.RC > 0 {
		return nil, fmt.Errorf(commandResult.Stdout)
	}

	var dst map[string][]*Win32_NTLogEvent
	err = json.Unmarshal([]byte(commandResult.Stdout), &dst)

	if err != nil {
		return nil, err
	}

	return dst, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckWindowsEventLog) Configure(cfg *config.Configuration) (bool, error) {
	if cfg.WindowsEventLog && runtime.GOOS == "windows" && len(cfg.WindowsEventLogTypes) > 0 {
		// Check is enabled
		agentBinary, err := os.Executable()
		if err == nil {
			wmiPath := filepath.Dir(agentBinary) + string(os.PathSeparator) + "wmiexecutor.exe"
			c.WmiExecutorPath = wmiPath
		}

		var ageSec uint64 = 3600
		if cfg.WindowsEventLogAge > 0 {
			ageSec = uint64(cfg.WindowsEventLogAge)
		}
		c.age = ageSec
		c.logfiles = cfg.WindowsEventLogTypes

		return true, nil
	}

	return false, nil
}
