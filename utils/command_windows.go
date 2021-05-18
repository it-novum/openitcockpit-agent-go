package utils

import (
	"context"
	"os"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/v3/process"
)

var (
	commandSysproc = &syscall.SysProcAttr{
		// Do not open any cmd windows
		// Run all processes in an own process group to be able to kill the process group and all child processes
		HideWindow:    true, //This will also hide the powershell windows if you run go test.
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
)

func handleCommandError(arg0 string, err error) int {
	if strings.HasSuffix(err.Error(), "file does not exist") {
		if _, err := os.Stat(arg0); os.IsNotExist(err) {
			return NotFound
		}
		return NotExecutable
	}
	return Unknown
}

func findProcessChildren(ctx context.Context, pid int32) ([]int32, error) {
	result := []int32{}

	// we have to fetch the process list every time, because it is possible that more processes have been created
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	for _, proc := range procs {
		parentId, err := proc.PpidWithContext(ctx)
		if err != nil {
			continue
		}
		if parentId == pid {
			result = append(result, proc.Pid)
		}
	}
	return result, nil
}

func killWithChildrenChildren(ctx context.Context, pid int32) {
	pidSlice := []int32{pid}

	for len(pidSlice) > 0 {
		currentPid := pidSlice[0]
		pidSlice = pidSlice[1:]

		proc, err := os.FindProcess(int(currentPid))
		if err != nil {
			continue
		}

		// we have to kill the parent process first, so no new child processes will be created
		// this is also the reason why we have to compare the blank process id's
		_ = proc.Kill()

		children, err := findProcessChildren(ctx, currentPid)
		if err != nil {
			continue
		}
		pidSlice = append(pidSlice, children...)
	}

}

func killProcessGroup(p *os.Process) error {
	pid := p.Pid
	if pid == -1 {
		return nil
	}

	killWithChildrenChildren(context.Background(), int32(pid))

	return nil
}
