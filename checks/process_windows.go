package checks

import (
	"context"
	"runtime"

	"github.com/it-novum/openitcockpit-agent-go/safemaths"
	"github.com/it-novum/openitcockpit-agent-go/winpsapi"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/mem"
	log "github.com/sirupsen/logrus"
	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows"
)

// https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-process
type Win32_Process struct {
	ProcessId       uint64
	ParentProcessId uint64
	CommandLine     string
	Name            string
	ExecutablePath  string
}

// https://wutils.com/wmi/root/cimv2/win32_perfformatteddata_perfproc_process/
// PercentProcessorTime is not buggy as https://stackoverflow.com/a/11565773 claims
// It is the total CPU percentage and has to be divided by the number of CPU cores
type Win32_PerfFormattedData_PerfProc_Process struct {
	IDProcess            uint64
	WorkingSet           uint64
	WorkingSetPrivate    uint64
	PrivateBytes         uint64
	HandleCount          uint64
	PercentProcessorTime uint16
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

func (c *CheckProcess) Run(_ context.Context) (interface{}, error) {
	var processList []*Win32_Process
	var processPerf []*Win32_PerfFormattedData_PerfProc_Process

	if err := wmi.Query("SELECT processid,parentprocessid,commandline,name,ExecutablePath FROM Win32_Process", &processList); err != nil {
		return nil, errors.Wrap(err, "could not query wmi for process list")
	}

	if err := wmi.Query("SELECT IDProcess,WorkingSet,WorkingSetPrivate,PrivateBytes,HandleCount,PercentProcessorTime FROM Win32_PerfFormattedData_PerfProc_Process", &processPerf); err != nil {
		return nil, errors.Wrap(err, "could not query wmi for process perfdata list")
	}

	processMapPerfdata := make(map[uint64]*Win32_PerfFormattedData_PerfProc_Process, len(processPerf))
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
			log.Debugln("could not fetch process information (", proc.ProcessId, "): ", err)
			newIgnorePid[proc.ProcessId] = 1
		} else {
			result.CreateTime = stat.TimeStat.CreateTime
			//result.CPUPercent = (stat.TimeStat.User + stat.TimeStat.System) / float64(runtime.NumCPU())
			result.CPUPercent = float64(perfdata.PercentProcessorTime) / float64(runtime.NumCPU())
		}
	}

	c.processCacheIgnorePid = newIgnorePid
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
