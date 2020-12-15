package utils

import (
	"os"
)

func killProcess(p *os.Process) {
	p.Kill()
}
