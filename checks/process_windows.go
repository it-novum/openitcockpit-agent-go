package checks

import (
	"context"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/shirou/gopsutil/v3/process"
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckProcess) Run(ctx context.Context) (interface{}, error) {
	var err error

	if c.ProcessCache == nil {
		c.ProcessCache = map[int32]*resultProcess{}
	}

	machineMemory, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not get system memory information")
	}
	total := machineMemory.Total

	pids, err := process.PidsWithContext(ctx)
	if err != nil {
		return nil, err
	}

	processResults := make([]*resultProcess, 0, len(pids))
	newCache := make(map[int32]*resultProcess)

	for _, pid := range pids {
		p, err := process.NewProcessWithContext(ctx, pid)
		if err != nil {
			// We ignore errors, because we the process just might have stopped or
			// is inaccessible in some way
			continue
		}
		createTime, err := p.CreateTimeWithContext(ctx)
		if err != nil {
			continue
		}

		result, ok := c.ProcessCache[p.Pid]
		if ok {
			if result.CreateTime != createTime {
				result = nil
			}
		}

		if result == nil {
			result = &resultProcess{
				Pid:        pid,
				CreateTime: createTime,
			}
			result.Name, _ = p.NameWithContext(ctx)
			result.Username, _ = p.UsernameWithContext(ctx)
			if parent, err := p.ParentWithContext(ctx); err == nil {
				result.Ppid = parent.Pid
			}
			result.Cmdline, _ = p.CmdlineWithContext(ctx)
			result.Exe, _ = p.ExeWithContext(ctx)
		}

		result.CPUPercent, _ = p.CPUPercentWithContext(ctx)
		result.Nice, _ = p.NiceWithContext(ctx)
		if memoryInfo, err := p.MemoryInfoWithContext(ctx); err == nil {
			result.Memory.RSS = memoryInfo.RSS
			result.Memory.VMS = memoryInfo.VMS
			result.Memory.HWM = memoryInfo.HWM
			result.Memory.Data = memoryInfo.Data
			result.Memory.Stack = memoryInfo.Stack
			result.Memory.Locked = memoryInfo.Locked
			result.Memory.Swap = memoryInfo.Swap
		}
		result.MemoryPercent = 100 * float32(result.Memory.RSS) / float32(total)

		newCache[pid] = result

		processResults = append(processResults, result)
	}
	c.ProcessCache = newCache

	return processResults, nil
}
