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
	Pid           int32    `json:"pid"`
	Ppid          int32    `json:"ppid"`
	Username      string   `json:"username"`
	Name          string   `json:"name"`
	CPUPercent    float64  `json:"cpu_percent"`
	MemoryPercent float32  `json:"memory_percent"`
	Cmdline       string   `json:"cmdline"`
	Status        []string `json:"status"`
	Exe           string   `json:"exec"`
	Nice          int32    `json:"nice_level"`
	NumFds        int32    `json:"num_fds"`
	Memory        struct {
		RSS    uint64 `json:"rss"`
		VMS    uint64 `json:"vms"`
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
func (c *CheckProcess) Run(ctx context.Context) (*CheckResult, error) {
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
	return &CheckResult{Result: processResults}, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckProcess) Configure(config *config.Configuration) (bool, error) {
	return config.Processes, nil
}
