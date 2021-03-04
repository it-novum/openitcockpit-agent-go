// +build linux darwin

package checks

import (
	"context"

	"github.com/shirou/gopsutil/v3/host"
)

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

	for _, user := range users {
		result := &resultUser{
			Name:     user.User,
			Terminal: user.Terminal,
			Host:     user.Host,
			Started:  int64(user.Started),
		}
		userResults = append(userResults, result)
	}
	return userResults, nil
}
