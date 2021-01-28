// +build !windows

package utils

import (
	"errors"
	"os"
	"syscall"
)

var (
	commandSysproc *syscall.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
)

func handleCommandError(arg0 string, err error) int {
	if os.IsNotExist(err) { // does not work with windows
		return NotFound
	}

	if os.IsPermission(err) { // does not work with windows
		return NotExecutable
	}

	return Unknown
}

func killProcessGroup(p *os.Process) error {
	if p.Pid == -1 {
		return errors.New("os: process already released")
	}
	if p.Pid == 0 {
		return errors.New("os: process not initialized")
	}
	sig := os.Kill
	s, ok := sig.(syscall.Signal)
	if !ok {
		return errors.New("os: unsupported signal type")
	}
	if e := syscall.Kill(-p.Pid, s); e != nil {
		if e == syscall.ESRCH {
			return errors.New("os: process already finished")
		}
		return e
	}
	return nil
}
