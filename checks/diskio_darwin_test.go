package checks

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestChecksCheckDiskIO(t *testing.T) {

	check := &CheckDiskIo{}

	cr, err := check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	results, ok := cr.([]*resultDiskIo)
	if !ok {
		t.Fatal("False type")

	}

	var oldIops uint64 = 0
	for _, result := range results {
		if result.Device == "disk0" {
			fmt.Printf("Device [Check 1]: %s\n", result.Device)
			fmt.Printf("LoadPercent: %v\n", result.LoadPercent)
			fmt.Printf("TotalIops: %v\n", result.TotalIops)
			fmt.Printf("TotalAvgWait: %v\n", result.TotalAvgWait)
			oldIops = result.TotalIops
		}
	}

	time.Sleep(10 * time.Second)

	cr, err = check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	results, ok = cr.([]*resultDiskIo)
	if !ok {
		t.Fatal("False type")

	}

	var newIops uint64 = 0
	for _, result := range results {
		if result.Device == "disk0" {
			fmt.Printf("Device [Check 2]: %s\n", result.Device)
			fmt.Printf("LoadPercent: %v\n", result.LoadPercent)
			fmt.Printf("TotalIops: %v\n", result.TotalIops)
			fmt.Printf("TotalAvgWait: %v\n", result.TotalAvgWait)
			newIops = result.TotalIops
		}
	}

	if newIops <= oldIops {
		t.Fatal("No IOPS recorded")
	}

}
