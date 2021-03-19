package checks

import (
	"github.com/it-novum/openitcockpit-agent-go/config"
)

// CheckMem gathers information about system memory
type CheckMem struct {
}

// Name will be used in the response as check name
func (c *CheckMem) Name() string {
	return "memory"
}

type resultMemory struct {
	Total     uint64  `json:"total"`     // Total amount of memory (RAM) in bytes
	Available uint64  `json:"available"` // Available memory in bytes (inactive_count + free_count)
	Percent   float64 `json:"percent"`   // Used memory as percentage
	Used      uint64  `json:"used"`      // Used memory in bytes (totalCount - availableCount)
	Free      uint64  `json:"free"`      // Free memory in bytes
	Active    uint64  `json:"active"`    // Active memory in bytes
	Inactive  uint64  `json:"inactive"`  // Inactive memory in bytes
	Wired     uint64  `json:"wired"`     // Wired memory in bytes - macOS and BSD only - memory that is marked to always stay in RAM. It is never moved to disk
}

// Configure the command or return false if the command was disabled
func (c *CheckMem) Configure(config *config.Configuration) (bool, error) {
	return config.Memory, nil
}
