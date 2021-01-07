package checks

import (
	"context"
	"strings"
	"testing"
)

func TestChecksNet(t *testing.T) {

	check := &CheckNet{}

	cr, err := check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	results, ok := cr.(map[string]*resultNet)
	if !ok {
		t.Fatal("False type")

	}

	for name, nic := range results {
		if strings.HasPrefix("enp", name) || strings.HasPrefix("eth", name) {
			if nic.Speed < 10 && nic.Speed > 0 {
				t.Fatal("Network connection has less than 10 mbit?")
			}
		}

	}

}
