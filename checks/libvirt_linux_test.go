package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestChecksCheckLibvirt(t *testing.T) {
	check := &CheckLibvirt{}

	// First check run
	_, _ = check.Run(context.Background())

	time.Sleep(5 * time.Second)

	// Second run to get values - same as for diskio, netio and any other counter based checks...
	cr, _ := check.Run(context.Background())

	results, ok := cr.(map[string]*resultLibvirtDomain)
	if !ok {
		t.Fatal("False type")
	}

	js, _ := json.Marshal(results)
	fmt.Println(string(js))
}
