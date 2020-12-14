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
