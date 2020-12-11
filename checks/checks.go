package checks

import "context"

// CheckResult will be sent to oitc
type CheckResult struct {
	// Result will be serialized to json
	Result interface{}
}

// Check should gather the required information
type Check interface {
	// Name will be used in the response as check name
	Name() string

	// Run the actual check
	// if error != nil the check result will be nil
	// ctx can be canceled and runs the timeout
	// CheckResult will be serialized after the return and should not change until the next call to Run
	Run(ctx context.Context) (*CheckResult, error)

	// DefaultConfiguration contains the variables for the configuration file and the default values
	// can be nil if no configuration is required
	DefaultConfiguration() interface{}

	// Configure should verify the configuration and set it
	// will be run after every reload
	// if DefaultConfiguration returns nil, the parameter will also be nil
	Configure(interface{}) error
}
