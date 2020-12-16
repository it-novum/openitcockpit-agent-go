package checks

import (
	"context"
	"testing"
)

func TestGetServiceListFromDbus(t *testing.T) {

	check := &CheckSystemd{}
	config := check.DefaultConfiguration()
	check.Configure(config)
	results, err := check.getServiceListViaDbus(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("Empty result from launchctl list")
	}

}
