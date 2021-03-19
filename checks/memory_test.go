package checks

import (
	"context"
	"testing"
)

func TestCheckMemory(t *testing.T) {

	check := &CheckMem{}

	cr, err := check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	result, ok := cr.(*resultMemory)
	if !ok {
		t.Fatal("False type")
	}

	if result.Percent < 1.0 {
		t.Fatal("Used memory less than 1% ??")
	}

}
