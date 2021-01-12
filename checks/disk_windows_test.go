package checks

import (
	"context"
	"encoding/json"
	"fmt"
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

	js, _ := json.Marshal(results)

	fmt.Println(string(js))

}
