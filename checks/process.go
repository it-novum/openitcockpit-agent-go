package checks

import (
	"context"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/shirou/gopsutil/v3/process"
)

// CheckProcess gathers information about each process
type CheckProcess struct {
}

type resultProcess struct {
	Pid           int32    `json:"pid"`            // Pid of the process itself
	Ppid          int32    `json:"ppid"`           // Pid of the parent process
	Username      string   `json:"username"`       // Username which runs the process
	Name          string   `json:"name"`           // (empty on macOS?)
	CPUPercent    float64  `json:"cpu_percent"`    // Used CPU resources as percentage
	MemoryPercent float32  `json:"memory_percent"` // Used memory resources as percentage
	Cmdline       string   `json:"cmdline"`        // command line e.g.: /Applications/Firefox.app/Contents/MacOS/firefox
	Status        []string `json:"status"`         // https://psutil.readthedocs.io/en/latest/#process-status-constants
	Exe           string   `json:"exec"`           // e.g: /Applications/Firefox.app/Contents/MacOS/firefox
	Nice          int32    `json:"nice_level"`     // e.g.: 0
	NumFds        int32    `json:"num_fds"`        // Number of open file descriptor
	Memory        struct {
		// https://psutil.readthedocs.io/en/latest/#psutil.Process.memory_info
		RSS    uint64 `json:"rss"` // Resident Set Size in bytes
		VMS    uint64 `json:"vms"` // Virtual Memory Size in bytes
		HWM    uint64 `json:"hwm"`
		Data   uint64 `json:"data"`
		Stack  uint64 `json:"stack"`
		Locked uint64 `json:"locked"`
		Swap   uint64 `json:"swap"`
	}
}

// Name will be used in the response as check name
func (c *CheckProcess) Name() string {
	return "processes"
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckProcess) Run(ctx context.Context) (interface{}, error) {
	var err error

	pids, err := process.PidsWithContext(ctx)
	if err != nil {
		return nil, err
	}

	processResults := make([]*resultProcess, 0, len(pids))

	// TODO log errors
	for _, pid := range pids {
		p, err := process.NewProcessWithContext(ctx, pid)

		if err != nil {
			// We ignore errors, because we the process just might have stopped or
			// is inaccessible in some way
			continue
		}
		result := &resultProcess{
			Pid: p.Pid,
		}
		result.Name, _ = p.NameWithContext(ctx)
		result.Username, _ = p.UsernameWithContext(ctx)
		result.CPUPercent, _ = p.CPUPercentWithContext(ctx)
		result.MemoryPercent, _ = p.MemoryPercentWithContext(ctx)
		if parent, err := p.ParentWithContext(ctx); err == nil {
			result.Ppid = parent.Pid
		}
		result.Cmdline, _ = p.CmdlineWithContext(ctx)
		result.Status, _ = p.StatusWithContext(ctx)
		result.Exe, _ = p.ExeWithContext(ctx)
		result.Nice, _ = p.NiceWithContext(ctx)
		result.NumFds, _ = p.NumFDsWithContext(ctx)
		if memoryInfo, err := p.MemoryInfoWithContext(ctx); err == nil {
			result.Memory.RSS = memoryInfo.RSS
			result.Memory.VMS = memoryInfo.VMS
			result.Memory.HWM = memoryInfo.HWM
			result.Memory.Data = memoryInfo.Data
			result.Memory.Stack = memoryInfo.Stack
			result.Memory.Locked = memoryInfo.Locked
			result.Memory.Swap = memoryInfo.Swap
		}

		processResults = append(processResults, result)
	}
	return processResults, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckProcess) Configure(config *config.Configuration) (bool, error) {
	return config.Processes, nil
}
