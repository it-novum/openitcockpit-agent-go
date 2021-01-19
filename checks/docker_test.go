package checks

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestChecksCheckDocker(t *testing.T) {

	check := &CheckDocker{}

	cr, err := check.Run(context.Background())
	if err != nil {
		if strings.Contains(err.Error(), "error during connect: This error may") {
			fmt.Println("Docker not installed or running on this system ???")
			t.SkipNow()
		}
		t.Fatal(err)

	}

	results, ok := cr.([]*resultDocker)
	if !ok {
		t.Fatal("False type")

	}

	if len(results) == 0 {
		fmt.Println("No running docker containers found")
	}

	if len(results) > 0 {
		for _, result := range results {
			if result.MemoryUsed == 0.0 {
				t.Fatal("Container memory usage is 0.0 - thats suspect!")
			}
		}
	}

}
