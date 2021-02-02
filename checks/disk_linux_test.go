package checks

import (
	"context"
	"strings"
	"testing"
)

func TestChecksCheckDisk(t *testing.T) {

	check := &CheckDisk{}

	cr, err := check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	results, ok := cr.([]*resultDisk)
	if !ok {
		t.Fatal("False type")

	}

	for _, result := range results {
		if strings.HasPrefix(result.Disk.Device, "/dev/sda") || strings.HasPrefix(result.Disk.Device, "/dev/nvme") {

			freeDiskSpacePercentage := 100.0 - result.Usage.Percent

			if freeDiskSpacePercentage <= 5.0 {
				t.Fatal("Equal or less than 5% free disk space available")
			}
		}
	}
}
