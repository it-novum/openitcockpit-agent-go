package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
		if result.Device == "disk0" {
			fmt.Printf("Device [Check 1]: %s\n", result.Device)
			fmt.Printf("LoadPercent: %v\n", result.LoadPercent)
			fmt.Printf("TotalIopsPerSecond: %v\n", result.TotalIopsPerSecond)
			fmt.Printf("TotalAvgWait: %v\n", result.TotalAvgWait)
			oldIops = result.TotalIopsPerSecond
		}
	}

	time.Sleep(5 * time.Second)

	file, ioutilErr := ioutil.TempFile("", "makeSomeIops")
	if ioutilErr != nil {
		log.Fatal(ioutilErr)
	}
	defer os.Remove(file.Name())

	if _, err = file.WriteString("Make some iops on the disk, that the test does not fail"); err != nil {
		fmt.Println(err)
	}

	if err = file.Sync(); err != nil {
		fmt.Println(err)
	}

	if err = file.Close(); err != nil {
		fmt.Println(err)
	}

	time.Sleep(5 * time.Second)

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
		if result.Device == "disk0" {
			fmt.Printf("Device [Check 2]: %s\n", result.Device)
			fmt.Printf("LoadPercent: %v\n", result.LoadPercent)
			fmt.Printf("TotalIopsPerSecond: %v\n", result.TotalIopsPerSecond)
			fmt.Printf("TotalAvgWait: %v\n", result.TotalAvgWait)

			js, _ := json.Marshal(result)
			fmt.Println(string(js))
			newIops = result.TotalIopsPerSecond
		}
	}

	if newIops <= oldIops {
		t.Fatal("No IOPS recorded")
	}

}
