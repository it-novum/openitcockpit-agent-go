package checks

import (
	"github.com/it-novum/openitcockpit-agent-go/config"
)

// CheckSwap gathers information about system swap
type CheckSwap struct {
}

// Name will be used in the response as check name
func (c *CheckSwap) Name() string {
	return "swap"
}

type resultSwap struct {
	Total   uint64  `json:"total"`   // Total amount of swap space in bytes
	Percent float64 `json:"percent"` // Used swap space as percentage
	Used    uint64  `json:"used"`    // Used swap space in bytes
	Free    uint64  `json:"free"`    // Free swap space in bytes
	Sin     uint64  `json:"sin"`     // Linux only - Number of bytes the system has swapped in from disk
	Sout    uint64  `json:"sout"`    // Linux only - Number of bytes the system has swapped out to disk
}

// Configure the command or return false if the command was disabled
func (c *CheckSwap) Configure(config *config.Configuration) (bool, error) {
	return config.Swap, nil
}
