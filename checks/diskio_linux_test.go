package checks

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
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
		if strings.HasSuffix(result.Device, "da") || strings.HasPrefix(result.Device, "nvme") {
			fmt.Printf("Device [Check 1]: %s\n", result.Device)
			fmt.Printf("LoadPercent: %v\n", result.LoadPercent)
			fmt.Printf("TotalIopsPerSecond: %v\n", result.TotalIopsPerSecond)
			fmt.Printf("TotalAvgWait: %v\n", result.TotalAvgWait)
			oldIops = oldIops + result.TotalIopsPerSecond
		}
	}

	file, err := os.CreateTemp(os.TempDir(), "makeSomeIops")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	if _, err := io.CopyN(file, rand.Reader, 8192); err != nil {
		t.Fatal(err)
	}
	if err = file.Sync(); err != nil {
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
	os.Remove(file.Name())

	// Hopefully something will do some IOPs in the meantime
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
		if strings.HasSuffix(result.Device, "da") || strings.HasPrefix(result.Device, "nvme") {
			fmt.Printf("Device [Check 2]: %s\n", result.Device)
			fmt.Printf("LoadPercent: %v\n", result.LoadPercent)
			fmt.Printf("TotalIopsPerSecond: %v\n", result.TotalIopsPerSecond)
			fmt.Printf("TotalAvgWait: %v\n", result.TotalAvgWait)
			newIops = newIops + result.TotalIopsPerSecond
		}
	}

	if newIops <= oldIops {
		t.Fatal("No IOPS recorded")
	}
}
