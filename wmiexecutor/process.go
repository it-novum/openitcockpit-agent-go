package wmiexecutor

import (
	"encoding/json"

	"github.com/StackExchange/wmi"
	"github.com/it-novum/openitcockpit-agent-go/checks"

	"runtime"

	"github.com/it-novum/openitcockpit-agent-go/safemaths"
	"github.com/it-novum/openitcockpit-agent-go/winpsapi"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/sys/windows"
)

// CheckProcess gathers information about each process
type CheckProcess struct {
	verbose bool
	debug   bool

	processCacheCmdline   map[uint64]string
	processCacheIgnorePid map[uint64]uint64
}

type resultMemoryWindows struct {
	WorkingSet        uint64 `json:"working_set"`
	WorkingSetPrivate uint64 `json:"working_set_private"`
	PrivateBytes      uint64 `json:"private_bytes"`
}

type resultProcess struct {
	Pid           uint64   `json:"pid"`            // Pid of the process itself
	Ppid          uint64   `json:"ppid"`           // Pid of the parent process
	Username      string   `json:"username"`       // Username which runs the process
	Name          string   `json:"name"`           // (empty on macOS?)
	CPUPercent    float64  `json:"cpu_percent"`    // Used CPU resources as percentage
	MemoryPercent float64  `json:"memory_percent"` // Used memory resources as percentage
	Cmdline       string   `json:"cmdline"`        // command line e.g.: /Applications/Firefox.app/Contents/MacOS/firefox
	Status        []string `json:"status"`         // https://psutil.readthedocs.io/en/latest/#process-status-constants
	Exe           string   `json:"exec"`           // e.g: /Applications/Firefox.app/Contents/MacOS/firefox
	Nice          int64    `json:"nice_level"`     // e.g.: 0
	NumFds        uint64   `json:"num_fds"`        // Number of open file descriptor
	Memory        interface{}
	CreateTime    int64
}

func (c *CheckProcess) Configure(conf *Configuration) error {
	c.verbose = conf.verbose
	c.debug = conf.debug

	return nil
}

// Query WMI
// if error != nil the check result will be nil
func (c *CheckProcess) RunQuery() (string, error) {

	var processList []*checks.Win32_Process
	var processPerf []*checks.Win32_PerfFormattedData_PerfProc_Process

	if err := wmi.Query("SELECT processid,parentprocessid,commandline,name,ExecutablePath FROM Win32_Process", &processList); err != nil {
		return "", errors.Wrap(err, "could not query wmi for process list")
	}

	if err := wmi.Query("SELECT IDProcess,WorkingSet,WorkingSetPrivate,PrivateBytes,HandleCount FROM Win32_PerfFormattedData_PerfProc_Process", &processPerf); err != nil {
		return "", errors.Wrap(err, "could not query wmi for process perfdata list")
	}

	processMapPerfdata := make(map[uint64]*checks.Win32_PerfFormattedData_PerfProc_Process, len(processPerf))
	for _, p := range processPerf {
		processMapPerfdata[p.IDProcess] = p
	}

	ignorePid := c.processCacheIgnorePid
	if ignorePid == nil {
		ignorePid = map[uint64]uint64{}
	}
	newIgnorePid := map[uint64]uint64{}

	totalMemory := 0.0
	sysMem, err := mem.VirtualMemory()
	if err == nil {
		totalMemory = float64(sysMem.Total)
	}

	processResults := make([]*resultProcess, 0, len(processList))
	for _, proc := range processList {
		perfdata, ok := processMapPerfdata[proc.ProcessId]
		if !ok {
			// process probably vanished
			continue
		}

		result := &resultProcess{
			Pid:           proc.ProcessId,
			Ppid:          proc.ParentProcessId,
			Name:          proc.Name,
			MemoryPercent: safemaths.DivideFloat64(float64(perfdata.WorkingSetPrivate), totalMemory) * 100,
			Cmdline:       proc.CommandLine,
			Exe:           proc.ExecutablePath,
			NumFds:        perfdata.HandleCount,
			Memory: &resultMemoryWindows{
				WorkingSet:        perfdata.WorkingSet,
				WorkingSetPrivate: perfdata.WorkingSetPrivate,
				PrivateBytes:      perfdata.PrivateBytes,
			},
		}

		processResults = append(processResults, result)

		ignore := ignorePid[proc.ProcessId]
		if ignore == 1 {
			newIgnorePid[proc.ProcessId] = 1
			continue
		}
		if stat, err := fetchProcessInfo(proc.ProcessId); err != nil {
			//log.Debugln("could not fetch process information (", proc.ProcessId, "): ", err)
			newIgnorePid[proc.ProcessId] = 1
		} else {
			result.CreateTime = stat.TimeStat.CreateTime
			result.CPUPercent = (stat.TimeStat.User + stat.TimeStat.System) / float64(runtime.NumCPU())
		}
	}

	js, err := json.Marshal(processResults)
	if err != nil {
		return "", err
	}

	return string(js), nil
}

func fetchProcessInfo(pid uint64) (*checks.ProcessInfo, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return nil, errors.Wrap(err, "win32 OpenProcess")
	}
	defer func() {
		_ = windows.CloseHandle(handle)
	}()

	result := &checks.ProcessInfo{}
	if stat, err := winpsapi.GetProcessTimes(handle); err != nil {
		return nil, err
	} else {
		result.TimeStat = stat
	}

	return result, nil
}
