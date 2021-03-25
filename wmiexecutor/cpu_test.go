package wmiexecutor

import (
	"encoding/json"
	"testing"

	"github.com/it-novum/openitcockpit-agent-go/checks"
)

func TestCpuCheckToJsonAndBack(t *testing.T) {
	conf := &Configuration{
		verbose: false,
		debug:   false,
	}

	cpu := &CheckCpu{}
	cpu.Configure(conf)
	resultAsJson, err := cpu.RunQuery()

	if err != nil {
		t.Fatal(err)
	}

	var results []*checks.Win32_PerfFormattedData_PerfOS_Processor
	err = json.Unmarshal([]byte(resultAsJson), &results)

	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Fatal("Empty result")
	}
}
