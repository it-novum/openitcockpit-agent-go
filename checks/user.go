package checks

import (
	"context"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/shirou/gopsutil/v3/host"
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
	Started  int    `json:"started"`  // The creation time as a floating point number expressed in seconds since the epoch. - Maybe not under macOS???
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckUser) Run(ctx context.Context) (interface{}, error) {
	users, err := host.UsersWithContext(ctx)
	if err != nil {
		return nil, err
	}

	userResults := make([]*resultUser, 0, len(users))

	// TODO log errors
	for _, user := range users {
		result := &resultUser{
			Name:     user.User,
			Terminal: user.Terminal,
			Host:     user.Host,
			Started:  user.Started,
		}
		userResults = append(userResults, result)
	}
	return userResults, nil
}

// Configure the command or return false if the command was disabled
func (c *CheckUser) Configure(config *config.Configuration) (bool, error) {
	return true, nil
}
