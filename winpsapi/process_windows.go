package winpsapi

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"time"
	"unicode/utf16"
	"unsafe"
)

func EnumProcesses() ([]uint32, error) {
	bufSize := uint32(2048)
	for {
		processIds := make([]uint32, bufSize)
		var bytesReturned uint32
		if err := windows.EnumProcesses(processIds, &bytesReturned); err != nil {
			return nil, err
		} else {
			numProc := bytesReturned / 4
			if bufSize > numProc {
				return processIds[:numProc], nil
			}
			bufSize *= 4
		}
	}
}

func GetProcessImageFileName(handle windows.Handle) (string, error) {
	bufSize := uint32(windows.MAX_PATH + 2)
	for {
		buf := make([]uint16, bufSize)

		if n, err := getProcessImageFileName(handle, &buf[0], bufSize); err == windows.ERROR_INSUFFICIENT_BUFFER {
			bufSize *= 2
		} else if err != nil {
			return "", errors.Wrap(err, "win32 GetProcessImageFileName")
		} else {
			return windows.UTF16ToString(buf[:n]), nil
		}
	}
}

func QueryFullProcessImageName(handle windows.Handle) (string, error) {
	bufSize := uint32(windows.MAX_PATH + 2)
	for {
		buf := make([]uint16, bufSize)
		nSize := bufSize

		if err := queryFullProcessImageName(handle, 0, &buf[0], &nSize); err == windows.ERROR_INSUFFICIENT_BUFFER {
			bufSize *= 2
		} else if err != nil {
			return "", errors.Wrap(err, "win32 QueryFullProcessImageName")
		} else {
			return string(utf16.Decode(buf[:nSize])), nil
		}
	}
}

type systemTimes struct {
	CreateTime windows.Filetime
	ExitTime   windows.Filetime
	KernelTime windows.Filetime
	UserTime   windows.Filetime
}

type ProcessTimeStat struct {
	User       float64
	System     float64
	CreateTime int64
}

func GetProcessTimes(handle windows.Handle) (*ProcessTimeStat, error) {
	var times systemTimes

	if err := windows.GetProcessTimes(
		handle,
		&times.CreateTime,
		&times.ExitTime,
		&times.KernelTime,
		&times.UserTime,
	); err != nil {
		return nil, err
	}

	result := &ProcessTimeStat{
		User:   float64(times.UserTime.HighDateTime)*429.4967296 + float64(times.UserTime.LowDateTime)*1e-7,
		System: float64(times.KernelTime.HighDateTime)*429.4967296 + float64(times.KernelTime.LowDateTime)*1e-7,
	}

	var createTime windows.Systemtime
	if err := fileTimeToSystemTime(&times.CreateTime, &createTime); err == nil {
		result.CreateTime = time.Date(int(createTime.Year), time.Month(createTime.Month), int(createTime.Day), int(createTime.Hour), int(createTime.Minute), int(createTime.Second), int(createTime.Milliseconds)*1000000, time.UTC).Unix()
	} else {
		log.Debugln("win32 FileTimeToSystemTime: ", err)
	}

	return result, nil
}

type processMemoryCounters struct {
	CB                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
	PrivateUsage               uintptr
}

type ProcessMemStat struct {
	RSS uint64
	VMS uint64
}

func GetProcessMemoryInfo(handle windows.Handle) (*ProcessMemStat, error) {
	var mem processMemoryCounters
	mem.CB = uint32(unsafe.Sizeof(mem))

	if err := getProcessMemoryInfo(handle, &mem, mem.CB); err != nil {
		return nil, errors.Wrap(err, "win32 GetProcessMemoryInfo")
	}

	return &ProcessMemStat{
		RSS: uint64(mem.WorkingSetSize),
		VMS: uint64(mem.PrivateUsage),
	}, nil
}

type Process struct {
	PID      uint32
	ParentID uint32
	ExeFile  string
	TimeStat ProcessTimeStat
	MemStat  ProcessMemStat
}

func QueryProcessInformation(pid uint32) (*Process, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, pid)
	if err != nil {
		return nil, errors.Wrap(err, "win32 OpenProcess")
	}
	defer func() {
		_ = windows.CloseHandle(handle)
	}()

	process := &Process{
		PID: pid,
	}

	if path, err := QueryFullProcessImageName(handle); err != nil {
		return nil, err
	} else {
		process.ExeFile = path
	}

	if ts, err := GetProcessTimes(handle); err != nil {
		return nil, err
	} else if ts != nil {
		process.TimeStat = *ts
	}

	if mem, err := GetProcessMemoryInfo(handle); err != nil {
		return nil, err
	} else if mem != nil {
		process.MemStat = *mem
	}

	return process, nil
}

type Toolhelp32Process struct {
	PID             uint32
	ParentProcessID uint32
	ExeFile         string
}

func CreateToolhelp32Snapshot() ([]*Toolhelp32Process, error) {
	handle, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, errors.Wrap(err, "win32 CreateToolhelp32Snapshot")
	}
	defer func() {
		_ = windows.CloseHandle(handle)
	}()

	result := make([]*Toolhelp32Process, 0)

	var procEntry windows.ProcessEntry32
	procEntry.Size = uint32(unsafe.Sizeof(procEntry))
	for err = windows.Process32First(handle, &procEntry); err == nil; err = windows.Process32Next(handle, &procEntry) {
		if procEntry.ProcessID != 0 {
			result = append(result, &Toolhelp32Process{
				PID:             procEntry.ProcessID,
				ParentProcessID: procEntry.ParentProcessID,
				ExeFile:         windows.UTF16ToString(procEntry.ExeFile[:]),
			})
		}
	}
	if err != windows.ERROR_NO_MORE_FILES {
		return nil, err
	}
	return result, nil
}

//go:generate mkwinsyscall -output zsyscall_windows.go process_windows.go
//sys getProcessImageFileName(hProcess windows.Handle, lpImageFileName *uint16, nSize uint32) (n uint32, err error) [failretval==0] = psapi.GetProcessImageFileNameW
//sys queryFullProcessImageName(hProcess windows.Handle, dwFlags uint32, lpImageFileName *uint16, nSize *uint32) (err error) [failretval==0] = kernel32.QueryFullProcessImageNameW
//sys getProcessMemoryInfo(hProcess windows.Handle, mem *processMemoryCounters, cb uint32) (err error) [failretval==0] = psapi.GetProcessMemoryInfo
//sys fileTimeToSystemTime(filetime *windows.Filetime, systemtime *windows.Systemtime) (err error) [failretval==0] = kernel32.FileTimeToSystemTime
