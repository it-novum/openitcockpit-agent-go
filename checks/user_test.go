package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestChecksCheckUser(t *testing.T) {

	check := &CheckUser{}

	cr, err := check.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	results, ok := cr.([]*resultUser)
	if !ok {
		t.Fatal("False type")
	}

	js, _ := json.Marshal(results)
	fmt.Println(string(js))

	if os.Getenv("CIBUILD") != "" {
		// results will always be empty on the Jenkins server
		return
	}

	if len(results) == 0 {
		t.Fatal("No logged in users - this is impossible. Who is running this test? ;)")
	}

}
