package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
			// Windows
			if strings.Contains(err.Error(), "error during connect: This error may") {
				fmt.Println("Docker not installed or running on this system ???")
				t.SkipNow()
			}

			// macOS and Linux
			if strings.Contains(err.Error(), "Is the docker daemon running") {
				fmt.Println("Docker not installed or running on this system ???")
				t.SkipNow()
			}

			if strings.Contains(err.Error(), "Got permission denied while trying to connect to the Docker daemon socket") {
				fmt.Println(err)
				t.SkipNow()
			}
		}

		if err != nil {
			t.Errorf("Test of check %s failed with error: %s", c.Name(), err)
			continue
		}
		if r == nil {
			t.Errorf("Test of check %s returned nil", c.Name())
			continue
		}
		js, err := json.Marshal(r)
		if err != nil {
			t.Errorf("Test of check %s returned result that can't be marshaled: %s", c.Name(), err)
			continue
		}
		fmt.Println(string(js))
	}
}
