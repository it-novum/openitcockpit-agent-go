package checks

import (
	"context"

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
	Name     string `json:"name"`
	Terminal string `json:"terminal"`
	Host     string `json:"host"`
	Started  int    `json:"started"`
}

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckUser) Run(ctx context.Context) (*CheckResult, error) {
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
	return &CheckResult{Result: userResults}, nil
}

// DefaultConfiguration contains the variables for the configuration file and the default values
// can be nil if no configuration is required
func (c *CheckUser) DefaultConfiguration() interface{} {
	return nil
}

// Configure should verify the configuration and set it
// will be run after every reload
// if DefaultConfiguration returns nil, the parameter will also be nil
func (c *CheckUser) Configure(_ interface{}) error {
	return nil
}
