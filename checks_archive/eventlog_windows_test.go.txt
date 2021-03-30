package checks

import (
	"context"
	"encoding/json"
	"log"
	"testing"

	"github.com/it-novum/openitcockpit-agent-go/config"
)

func TestEventLog(t *testing.T) {
	c := CheckWindowsEventLog{}
	_, err := c.Configure(&config.Configuration{
		WindowsEventLog: true,
		WindowsEventLogTypes: []string{
			"Application",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := c.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	js, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	log.Println(string(js))
}
