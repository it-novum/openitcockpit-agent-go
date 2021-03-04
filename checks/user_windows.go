package checks

import (
	"context"
	"github.com/it-novum/openitcockpit-agent-go/winpsapi"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/process"
)

// Run the actual check
// if error != nil the check result will be nil
// ctx can be canceled and runs the timeout
// CheckResult will be serialized after the return and should not change until the next call to Run
func (c *CheckUser) Run(ctx context.Context) (interface{}, error) {
	procs, err := winpsapi.CreateToolhelp32Snapshot()
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch process list")
	}
	explorer := make([]uint32, 0, 16)
	for _, proc := range procs {
		if proc.ExeFile == "explorer.exe" {
			explorer = append(explorer, proc.PID)
		}
	}

	users := make(map[string]*resultUser)
	for _, pid := range explorer {
		p, err := process.NewProcessWithContext(ctx, int32(pid))
		if err != nil {
			continue
		}
		username, err := p.UsernameWithContext(ctx)
		if err != nil {
			continue
		}
		user, ok := users[username]
		if !ok {
			user = &resultUser{
				Name:    username,
				Started: 0,
			}
			users[username] = user
		}
		createTime, err := p.CreateTime()
		if err != nil {
			continue
		}
		if user.Started == 0 || user.Started > createTime {
			user.Started = createTime
		}
	}

	result := make([]*resultUser, 0, len(users))
	for _, user := range users {
		result = append(result, user)
	}

	return result, nil
}
