// +build darwin linux

package checks

import (
	"context"
	"fmt"
	"testing"
)

func TestChecksCheckCpuPercentages(t *testing.T) {

	check := &CheckCpu{}

	result, err := check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(result)

}
