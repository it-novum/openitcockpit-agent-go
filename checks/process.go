package checks

import (
	"github.com/it-novum/openitcockpit-agent-go/config"
)

// CheckProcess gathers information about each process
type CheckProcess struct {
	// processCacheCmdline for windows checks
	processCacheCmdline   map[uint64]string
	processCacheIgnorePid map[uint64]uint64
}

type resultMemoryPosix struct {
	// https://psutil.readthedocs.io/en/latest/#psutil.Process.memory_info
	// Darwin/Linux

	RSS    uint64 `json:"rss"` // Resident Set Size in bytes
	VMS    uint64 `json:"vms"` // Virtual Memory Size in bytes
	HWM    uint64 `json:"hwm"`
	Data   uint64 `json:"data"`
	Stack  uint64 `json:"stack"`
	Locked uint64 `json:"locked"`
	Swap   uint64 `json:"swap"` // page
}

type resultMemoryWindows struct {
	WorkingSet        uint64 `json:"working_set"`
	WorkingSetPrivate uint64 `json:"working_set_private"`
	PrivateBytes      uint64 `json:"private_bytes"`
}

type resultProcess struct {
	Pid           uint64   `json:"pid"`            // Pid of the process itself
	Ppid          uint64   `json:"ppid"`           // Pid of the parent process
	Username      string   `json:"username"`       // Username which runs the process
	Name          string   `json:"name"`           // (empty on macOS?)
	CPUPercent    float64  `json:"cpu_percent"`    // Used CPU resources as percentage
	MemoryPercent float64  `json:"memory_percent"` // Used memory resources as percentage
	Cmdline       string   `json:"cmdline"`        // command line e.g.: /Applications/Firefox.app/Contents/MacOS/firefox
	Status        []string `json:"status"`         // https://psutil.readthedocs.io/en/latest/#process-status-constants
	Exe           string   `json:"exec"`           // e.g: /Applications/Firefox.app/Contents/MacOS/firefox
	Nice          int64    `json:"nice_level"`     // e.g.: 0
	NumFds        uint64   `json:"num_fds"`        // Number of open file descriptor
	Memory        interface{}
	CreateTime    int64
}

// Name will be used in the response as check name
func (c *CheckProcess) Name() string {
	return "processes"
}

// Configure the command or return false if the command was disabled
func (c *CheckProcess) Configure(config *config.Configuration) (bool, error) {
	return config.Processes, nil
}
