package checks

import (
	"context"

	"github.com/shirou/gopsutil/v3/process"
)

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
