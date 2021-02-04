package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestChecksCheckSystemdServices(t *testing.T) {
	if os.Getenv("CIBUILD") != "" {
		t.SkipNow()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := &CheckSystemd{}
	if c.Name() == "" {
		t.Error("Invalid name")
	}
	r, err := c.Run(ctx)
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

func TestGetServiceListFromDbus(t *testing.T) {
	if os.Getenv("CIBUILD") != "" {
		t.SkipNow()
	}

	check := &CheckSystemd{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	results, err := check.getServiceListViaDbus(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("Empty result from systemd / dbus")
	}

	// The dbus service is available and running - because we ask dbus
	needle := "dbus.service"
	foundNeedle := false
	for _, result := range results {
		if result.Name == needle {
			foundNeedle = true

			if result.SubState != "running" {
				t.Fatal("dbus is not running - this is impossible!")
			}

		}
	}

	if foundNeedle == false {
		t.Fatal("Needle not found: " + needle)
	}
}
