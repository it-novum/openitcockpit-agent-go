package checks

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestChecksCheckDiskIO(t *testing.T) {

	check := &CheckDiskIo{}

	cr, err := check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	results, ok := cr.(map[string]*resultDiskIo)
	if !ok {
		t.Fatal("False type")

	}

	var oldIops uint64 = 0
	for _, result := range results {
		//fmt.Println(result)
		if strings.HasSuffix(result.Device, "da") {
			fmt.Printf("Device [Check 1]: %s\n", result.Device)
			fmt.Printf("LoadPercent: %v\n", result.LoadPercent)
			fmt.Printf("TotalIopsPerSecond: %v\n", result.TotalIopsPerSecond)
			fmt.Printf("TotalAvgWait: %v\n", result.TotalAvgWait)
			oldIops = result.TotalIops
		}
	}

	time.Sleep(10 * time.Second)

	cr, err = check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	results, ok = cr.(map[string]*resultDiskIo)
	if !ok {
		t.Fatal("False type")

	}

	var newIops uint64 = 0
	for _, result := range results {
		//fmt.Println(result)
		if strings.HasSuffix(result.Device, "da") {
			fmt.Printf("Device [Check 2]: %s\n", result.Device)
			fmt.Printf("LoadPercent: %v\n", result.LoadPercent)
			fmt.Printf("TotalIopsPerSecond: %v\n", result.TotalIopsPerSecond)
			fmt.Printf("TotalAvgWait: %v\n", result.TotalAvgWait)
			newIops = result.TotalIops
		}
	}

	if newIops <= oldIops {
		t.Fatal("No IOPS recorded")
	}
}
