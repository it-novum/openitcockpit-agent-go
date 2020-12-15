package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
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
			t.Fatal(err)
		}
		if r.Result == nil {
			t.Fatal("invalid result")
		}
		js, err := json.Marshal(r.Result)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(string(js))
	}
}

func TestChecksCheckLaunchdServices(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.SkipNow()
	}

	checks := []Check{
		&CheckLaunchd{},
	}

	for _, c := range checks {
		config := c.DefaultConfiguration()
		c.Configure(config)
		if c.Name() == "" {
			t.Error("Invalid name")
		}
		r, err := c.Run(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if r.Result == nil {
			t.Fatal("invalid result")
		}
		js, err := json.Marshal(r.Result)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(string(js))
	}
}
