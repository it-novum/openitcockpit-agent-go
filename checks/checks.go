package checks

import (
	"context"
	"fmt"
	"math"

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
	var res []Check
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

//Wrapdiff calculate the difference between last and curr
//If last > curr, try to guess the boundary at which the value must have wrapped
//by trying the maximum values of 64, 32 and 16 bit signed and unsigned ints.
func Wrapdiff(last, curr float64) (float64, error) {
	if last <= curr {
		return curr - last, nil
	}

	boundaries := []float64{64, 63, 32, 31, 16, 15}
	var currBoundary float64
	for _, boundary := range boundaries {
		if last > math.Pow(2, boundary) {
			currBoundary = boundary
		}
	}

	if currBoundary == 0 {
		return 0, fmt.Errorf("Couldn't determine boundary")
	}

	return math.Pow(2, currBoundary) - last + curr, nil
}

func WrapDiffUint32(last, curr uint32) uint32 {
	if last <= curr {
		return curr - last
	}

	return (math.MaxInt32 - last) + curr
}

func WrapDiffUint64(last, curr uint64) uint64 {
	if last <= curr {
		return curr - last
	}

	return (math.MaxUint64 - last) + curr
}

func WrapDiffInt32(last, curr int32) int32 {
	if last <= curr {
		return curr - last
	}

	return (math.MaxInt32 - last) + curr
}

func WrapDiffInt64(last, curr int64) int64 {
	if last <= curr {
		return curr - last
	}

	return (math.MaxInt64 - last) + curr
}
