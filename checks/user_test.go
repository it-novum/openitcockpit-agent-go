package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestChecksCheckUser(t *testing.T) {

	check := &CheckUser{}

	cr, err := check.Run(context.Background())
	if err != nil {
		if strings.Contains(err.Error(), "/var/run/utmp: no such file or directory") {
			// https://github.com/shirou/gopsutil/issues/900
			fmt.Println(err)
			fmt.Println("The golang docker image does not have this directory. Are you using the golang docker image?")

			return
		} else {
			t.Fatal(err)
		}
	}

	results, ok := cr.([]*resultUser)
	if !ok {
		t.Fatal("False type")
	}

	js, _ := json.Marshal(results)
	fmt.Println(string(js))
}
