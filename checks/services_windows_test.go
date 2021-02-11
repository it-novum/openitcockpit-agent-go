package checks

import (
	"context"
	"testing"
)

func TestGetServiceListFromWmi(t *testing.T) {
	c := &CheckWinService{}
	r, err := c.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	results := r.([]*resultWindowsServices)

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
