package checks

import (
	"context"
	"fmt"
	"testing"
)

func TestChecksCheckDiskIO(t *testing.T) {

	check := &CheckDiskIo{}
	config := check.DefaultConfiguration()
	check.Configure(config)

	results, err := check.Run(context.Background())
	fmt.Println(results)
	fmt.Println(err)

}