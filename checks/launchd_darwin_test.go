package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestChecksCheckLaunchdServices(t *testing.T) {
	checks := []Check{
		&CheckLaunchd{},
	}

	for _, c := range checks {
		if c.Name() == "" {
			t.Error("Invalid name")
		}
		r, err := c.Run(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if r == nil {
			t.Fatal("invalid result")
		}
		js, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(string(js))
	}
}

func TestGetServiceListFromLaunchctl(t *testing.T) {

	check := &CheckLaunchd{}
	results, err := check.getServiceListViaLaunchctl(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("Empty result from launchctl list")
	}

	// The Finder service is probably available and running on all Macs?
	needle := "com.apple.Finder"
	foundNeedle := false
	for _, result := range results {
		if result.Label == needle {
			foundNeedle = true
		}
	}

	if foundNeedle == false {
		t.Fatal("Needle not found: " + needle)
	}

}
