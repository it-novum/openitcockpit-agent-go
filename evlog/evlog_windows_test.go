package evlog

import (
	"encoding/json"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestQuery(t *testing.T) {
	e := EventLog{
		LogChannel: "Application",
	}
	events, err := e.Query()
	if err != nil {
		t.Fatal(err)
	}

	js, err := json.MarshalIndent(events, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	log.Println(string(js))
}
