package utils

import (
	"os"
	"syscall"
)

func killProcess(p *os.Process) {
	syscall.Kill(p.Pid, syscall.SIGKILL)
}
