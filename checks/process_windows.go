package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/utils"
	"github.com/it-novum/openitcockpit-agent-go/winpsapi"
	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

type Win32_Process struct {
	ProcessId       uint64
	ParentProcessId uint64
	CommandLine     string
	Name            string
	ExecutablePath  string
}

type Win32_PerfFormattedData_PerfProc_Process struct {
	IDProcess         uint64
	WorkingSet        uint64
	WorkingSetPrivate uint64
	PrivateBytes      uint64
	HandleCount       uint64
}

type ProcessInfo struct {
	TimeStat *winpsapi.ProcessTimeStat
}

func fetchProcessInfo(pid uint64) (*ProcessInfo, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return nil, errors.Wrap(err, "win32 OpenProcess")
	}
	defer func() {
		_ = windows.CloseHandle(handle)
	}()

	result := &ProcessInfo{}
	if stat, err := winpsapi.GetProcessTimes(handle); err != nil {
		return nil, err
	} else {
		result.TimeStat = stat
	}

	return result, nil
}

// Unfortunately the WMI library is suffering from a memory leak
// especially on windows Server 2016 and Windows 10.
// For this reason all WMI queries have been moved to an external binary (fork -> exec) to avoid any memory issues.
//
// Hopefully the memory issues will be fixed one day.
// This check used to look like this: https://github.com/it-novum/openitcockpit-agent-go/blob/a8ec01146e419a2db246844ca95cbe4ea560d9e6/checks/process_windows.go

func (c *CheckProcess) Run(ctx context.Context) (interface{}, error) {
	// exec wmiexecutor.exe to avoid memory leak
	timeout := 10 * time.Second
	commandResult, err := utils.RunCommand(ctx, utils.CommandArgs{
		Command: c.WmiExecutorPath + " --command process",
		Shell:   "",
		Timeout: timeout,
		Env: []string{
			"OITC_AGENT_WMI_EXECUTOR=1",
		},
	})

	if err != nil {
		return nil, err
	}

	if commandResult.RC > 0 {
		return nil, fmt.Errorf(commandResult.Stdout)
	}

	var processResults []*resultProcess
	err = json.Unmarshal([]byte(commandResult.Stdout), &processResults)

	if err != nil {
		return nil, err
	}

	return processResults, nil
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
/*
func (c *CheckProcess) Run(ctx context.Context) (interface{}, error) {
	procs, err := winpsapi.CreateToolhelp32Snapshot()
	if err != nil {
		return nil, err
	}

	processResults := make([]*resultProcess, 0, len(procs))

	cacheCmdline := c.processCacheCmdline
	if cacheCmdline == nil {
		cacheCmdline = map[uint64]string{}
	}
	newCacheCmdline := map[uint64]string{}

	ignorePid := c.processCacheIgnorePid
	if ignorePid == nil {
		ignorePid = map[uint64]uint64{}
	}
	newIgnorePid := map[uint64]uint64{}

	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	totalMem := memStat.Total

	for _, proc := range procs {
		pid64 := uint64(proc.PID)
		ignore := ignorePid[pid64]
		if ignore == 1 {
			newIgnorePid[pid64] = 1
			continue
		}

		if stat, err := winpsapi.QueryProcessInformation(proc.PID); err != nil {
			log.Debugln("Could not query process information (", proc.PID, "): ", err)
			newIgnorePid[pid64] = 1
		} else {
			var cmdline string

			cachedCmd, found := cacheCmdline[pid64]
			if found {
				cmdline = cachedCmd
			} else {
				var wmiResult []wmiProcessCmdline
				if err := wmi.Query(fmt.Sprint("SELECT CommandLine FROM win32_process where processid = ", proc.PID), &wmiResult); err != nil {
					log.Debugln("WMI query commandline failed (", proc.PID, ": ", err)
				} else {
					if len(wmiResult) > 0 {
						cmdline = wmiResult[0].CommandLine
					}
				}
			}

			newResult :=  &resultProcess{
				Pid:           pid64,
				Ppid:          uint64(proc.ParentProcessID),
				Name:          proc.ExeFile,
				CPUPercent:    stat.TimeStat.System + stat.TimeStat.User,
				MemoryPercent: safemaths.DivideFloat64(float64(stat.MemStat.RSS), float64(totalMem)) * 100,
				Cmdline:       cmdline,
				Exe:           stat.ExeFile,
				Memory: resultProcessMemory{
					RSS:    stat.MemStat.RSS,
					VMS:    stat.MemStat.VMS,
				},
				CreateTime: stat.TimeStat.CreateTime,
			}

			newCacheCmdline[pid64] = cmdline
			processResults = append(processResults, newResult)
		}
	}

	c.processCacheCmdline = newCacheCmdline
	c.processCacheIgnorePid = newIgnorePid

	return processResults, nil
}

*/
