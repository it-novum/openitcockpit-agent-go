package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestChecksWithDefault(t *testing.T) {
	checks := getPlatformChecks()

	for _, c := range checks {
		if c.Name() == "" {
			t.Error("Invalid name")
		}
		r, err := c.Run(context.Background())
		if err != nil {
			t.Errorf("Test of check %s failed with error: %s", c.Name(), err)
			continue
		}
		if r.Result == nil {
			t.Errorf("Test of check %s returned nil", c.Name())
			continue
		}
		js, err := json.Marshal(r.Result)
		if err != nil {
			t.Errorf("Test of check %s returned result that can't be marshaled: %s", c.Name(), err)
			continue
		}
		fmt.Println(string(js))
	}
}
