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
	results, ok := cr.Result.([]*resultDiskIo)
	if !ok {
		t.Fatal("False type")

	}

	for _, result := range results {
		fmt.Println(result)
		fmt.Println(result.LoadPercent)
		fmt.Println(result.TotalIops)
		fmt.Println(result.TotalAvgWait)
	}

	time.Sleep(2 * time.Second)

	cr, err = check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	results, ok = cr.Result.([]*resultDiskIo)
	if !ok {
		t.Fatal("False type")

	}

	for _, result := range results {
		fmt.Println(result)
		fmt.Println(result.LoadPercent)
		fmt.Println(result.TotalIops)
		fmt.Println(result.TotalAvgWait)
	}

}
