package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestChecksCheckWmiServices(t *testing.T) {
	checks := []Check{
		&CheckWinService{},
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

func TestGetServiceListFromWmi(t *testing.T) {
	check := &CheckWinService{}
	results, err := check.getServiceListViaWmi(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("Empty result from WMI")
	}

	// Is Winmgmt running on al windows systems by default?
	needle := "Winmgmt"
	foundNeedle := false
	for _, result := range results {
		//js, _ := json.Marshal(result)
		//fmt.Println(string(js))

		if result.Name == needle {
			foundNeedle = true

			if result.Status != "Running" {
				t.Fatal("Winmgmt is not running")
			}

		}
	}

	if foundNeedle == false {
		t.Fatal("Needle not found: " + needle)
	}

}
