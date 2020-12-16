package checks

import (
	"context"
	"testing"
)

func TestGetServiceListFromLaunchctl(t *testing.T) {

	check := &CheckLaunchd{}
	config := check.DefaultConfiguration()
	check.Configure(config)
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
