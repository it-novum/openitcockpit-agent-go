package utils

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestRunPingCommand(t *testing.T) {
	nanosec := 5 * 1000 * 1000 * 1000 // convert seconds into nanoseconds
	timeout := time.Duration(nanosec)
	result, err := runCommand("/tmp/test.sh", timeout)

	if result.Stdout[0:27] != "PING 127.0.0.1 (127.0.0.1):" {
		t.Error("Unexpected output for ping...")
	}

	fmt.Println(result)
	js, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(js))
}

func TestCommandTimeout(t *testing.T) {
	nanosec := 5 * 1000 * 1000 * 1000 // convert seconds into nanoseconds
	timeout := time.Duration(nanosec)
	result, err := runCommand("sleep 10", timeout)

	if result.RC != 124 {
		t.Error("No Timeout??")
	}

	js, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(js))
}

func TestCommandNotFound(t *testing.T) {
	nanosec := 5 * 1000 * 1000 * 1000 // convert seconds into nanoseconds
	timeout := time.Duration(nanosec)
	result, err := runCommand("foobar 123", timeout)

	if result.Stderr[0:14] != "Unknown error:" || result.RC != 3 {
		t.Error("Unexpected output or return code")
	}

	js, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(js))
}

func TestCommandNotFoundFromOs(t *testing.T) {
	nanosec := 5 * 1000 * 1000 * 1000 // convert seconds into nanoseconds
	timeout := time.Duration(nanosec)
	result, err := runCommand("/foo/bar 123", timeout)

	if result.RC != 127 {
		t.Error("Unexpected return code")
	}

	js, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(js))
}
