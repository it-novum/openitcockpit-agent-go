package checks

import (
	"context"

	"github.com/it-novum/openitcockpit-agent-go/config"
)

// Check should gather the required information
type Check interface {
	// Name will be used in the response as check name
	Name() string

	// Run the actual check
	// if error != nil the check result will be nil
	// ctx can be canceled and runs the timeout
	// CheckResult will be serialized after the return and should not change until the next call to Run
	Run(ctx context.Context) (interface{}, error)

	// Configure the command or return false if the command was disabled
	Configure(config *config.Configuration) (bool, error)
}

func ChecksForConfiguration(config *config.Configuration) ([]Check, error) {
	res := []Check{}
	checks := getPlatformChecks()
	for _, check := range checks {
		ok, err := check.Configure(config)
		if err != nil {
			return nil, err
		}
		if ok {
			res = append(res, check)
		}
	}
	return res, nil
}
