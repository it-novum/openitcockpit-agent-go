package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestCheckSwap(t *testing.T) {

	check := &CheckSwap{}

	cr, err := check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	result, ok := cr.(*resultSwap)
	if !ok {
		t.Fatal("False type")
	}

	js, _ := json.Marshal(result)
	fmt.Println(string(js))

}
