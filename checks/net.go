package checks

import "github.com/it-novum/openitcockpit-agent-go/config"

// CheckNet gathers information about system network interfaces (netstats or net_states in the Python version)
type CheckNet struct {
}

// Name will be used in the response as check name
func (c *CheckNet) Name() string {
	return "net_stats"
}

const DUPLEX_FULL = 2
const DUPLEX_HALF = 1
const DUPLEX_UNKNOWN = 0

type resultNet struct {
	Isup   bool  `json:"isup"`   // True if up else false
	Duplex int   `json:"duplex"` // 0=unknown | 1=half | 2=full
	Speed  int64 `json:"speed"`  // Interface speed in Mbit/s - Linux and Windows only (0 on macOS)
	MTU    int64 `json:"mtu"`    // e.g.: 1500
}

// Configure the command or return false if the command was disabled
func (c *CheckNet) Configure(config *config.Configuration) (bool, error) {
	return true, nil
}
