package checks

import (
	"github.com/it-novum/openitcockpit-agent-go/config"
)

// CheckUser gathers information about system users
type CheckUser struct {
}

// Name will be used in the response as check name
func (c *CheckUser) Name() string {
	return "users"
}

type resultUser struct {
	Name     string `json:"name"`     // The name of the user
	Terminal string `json:"terminal"` // The tty or pseudo-tty associated with the user, if an
	Host     string `json:"host"`     // The host name associated with the entry, if any
	Started  int64  `json:"started"`  // The creation time as a floating point number expressed in seconds since the epoch. - Maybe not under macOS???
}

// Configure the command or return false if the command was disabled
func (c *CheckUser) Configure(config *config.Configuration) (bool, error) {
	return config.User, nil
}
