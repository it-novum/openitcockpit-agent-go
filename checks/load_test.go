package checks

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestChecksCheckLoad(t *testing.T) {

	check := &CheckLoad{}

	_, err := check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	//Run it twice for windows
	cr, err := check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	result, ok := cr.(*resultLoad)
	if !ok {
		t.Fatal("False type")

	}

	if result.Load1 <= 0 && result.Load5 <= 0 && result.Load15 <= 0 {
		//CPU load of 0 is impossible...
		t.Fatal("CPU load of 0 is impossible")
	}

	fmt.Println(result)

}
