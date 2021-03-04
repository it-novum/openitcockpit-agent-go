package winpsapi

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCreateToolhelp32Snapshot(t *testing.T) {
	procs, err := CreateToolhelp32Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	js, err := json.MarshalIndent(procs, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(js))
}

func TestProcessAPI(t *testing.T) {
	_, err := EnumProcesses()
	if err != nil {
		t.Fatal(err)
	}

	p, err := QueryProcessInformation(uint32(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}

	js, err := json.MarshalIndent(p, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(js))
}
