package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestChecksWithDefault(t *testing.T) {
	checks := []Check{
		&CheckMem{},
		//&CheckProcess{},
		&CheckAgent{},
		&CheckSwap{},
		&CheckUser{},
		&CheckDisk{},
		//&CheckDiskIo{}, //Check that all calcs are done right
		&CheckLoad{},
		&CheckNic{},
		&CheckSensor{},
	}

	for _, c := range checks {
		config := c.DefaultConfiguration()
		c.Configure(config)
		if c.Name() == "" {
			t.Error("Invalid name")
		}
		r, err := c.Run(context.Background())
		if err != nil {
			t.Fatalf("Test of check %s failed with error: %s", c.Name(), err)
		}
		if r.Result == nil {
			t.Fatalf("Test of check %s returned nil", c.Name())
		}
		js, err := json.Marshal(r.Result)
		if err != nil {
			t.Fatalf("Test of check %s returned result that can't be marshaled: %s", c.Name(), err)
		}
		fmt.Println(string(js))
	}
}
