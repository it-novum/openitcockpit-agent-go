package wmiexecutor

import (
	"encoding/json"

	"github.com/it-novum/openitcockpit-agent-go/winifmib"
	"golang.org/x/sys/windows"
)

// CheckNet gathers information about system network interfaces (netstats or net_states in the Python version)
type CheckNet struct {
	verbose bool
	debug   bool
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

func (c *CheckNet) Configure(conf *Configuration) error {
	c.verbose = conf.verbose
	c.debug = conf.debug

	return nil
}

func (c *CheckNet) RunCheck() (string, error) {

	netResults := make(map[string]*resultNet)

	mibTable, err := winifmib.GetIfTable2Ex(true)
	if err != nil {
		return "", err
	}
	defer mibTable.Close()

	table := mibTable.Slice()
	for _, nic := range table {
		name := windows.UTF16ToString(nic.Alias[:])
		if name != "" {
			netResults[name] = &resultNet{
				Isup:   nic.OperStatus == 1,
				MTU:    int64(nic.Mtu),
				Speed:  int64(nic.ReceiveLinkSpeed) / 1000 / 1000, // bits/s to mbits/s
				Duplex: DUPLEX_UNKNOWN,
			}
		}
	}

	js, err := json.Marshal(netResults)
	if err != nil {
		return "", err
	}

	return string(js), nil
}
