package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestChecksCheckNtp(t *testing.T) {

	check := &CheckNtp{}

	cr, err := check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	result, ok := cr.(*resultNtp)

	if !ok {
		t.Fatal("False type")
	}

	time.Sleep(2 * time.Second)
	if result.SyncStatus == false {
		fmt.Printf("Server is not using an NTP server")

		if (time.Now().Unix() - result.Timestamp) < 1 {
			// we slept for 2 seconds both timestamps need be at least 1 second apart
			t.Fatal(" we slept for 2 seconds both timestamps need be at least 1 second apart")
		}

	} else {
		fmt.Printf("Server using an NTP server")
		if result.Offset == 0 {
			t.Fatal("NTP offset is never exactly 0")
		}
	}

	js, _ := json.Marshal(result)
	fmt.Println(string(js))

}
