package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"
)

var testCommands = struct {
	sleep         string
	ping          string
	pingOutput    string
	notExecutable string
}{}

func init() {
	if runtime.GOOS == "windows" {
		testCommands.sleep = `powershell.exe -windowstyle hidden -command "start-sleep 10"`
		testCommands.ping = "ping 127.0.0.1 -n 1"
		testCommands.pingOutput = "127.0.0.1"
		testCommands.notExecutable = `C:\\Windows\\System32\\drivers\\etc\\hosts`
	} else {
		testCommands.sleep = "sleep 10"
		testCommands.ping = "echo thisImageHasNoPingCommand"
		testCommands.pingOutput = "thisImageHasNoPingCommand"
		testCommands.notExecutable = "/etc/hosts"
	}
}

func TestRunPingCommand(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	timeout := 5 * time.Second
	result, err := RunCommand(ctx, CommandArgs{
		Command: testCommands.ping,
		Timeout: timeout,
	})
	if err != nil {
		fmt.Println(err.Error())
		t.Fatal("there was an error running ping")
	}

	if !strings.Contains(result.Stdout, testCommands.pingOutput) {
		t.Error("Unexpected output for ping...")
	}

	fmt.Println(result)
	js, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(js))
	cancel()
}

func TestCommandTimeout(t *testing.T) {
	timeout := 5 * time.Second
	result, err := RunCommand(context.Background(), CommandArgs{
		Command: testCommands.sleep,
		Timeout: timeout,
	})
	if err == nil {
		t.Fatal("there was no error")
	}

	if result.RC != 124 {
		t.Error("No Timeout??")
	}

	js, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(js))
}

func TestCommandCancel(t *testing.T) {
	timeout := 5 * time.Second

	done := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		RunCommand(ctx, CommandArgs{
			Command: testCommands.sleep,
			Timeout: timeout,
		})
		done <- struct{}{}
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("cancel did not work")
	}
}

func TestCommandNotFound(t *testing.T) {
	timeout := 5 * time.Second
	result, err := RunCommand(context.Background(), CommandArgs{
		Command: "foobar 123",
		Timeout: timeout,
	})
	if err == nil {
		t.Error("there was no error")
	}

	if result.Stdout[0:14] != "Unknown error:" || result.RC != 3 {
		t.Errorf("Unexpected output '%s' or return code: %d", result.Stdout, result.RC)
	}

	js, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(js))
}

func TestCommandNotFoundFromOs(t *testing.T) {
	timeout := 5 * time.Second
	result, err := RunCommand(context.Background(), CommandArgs{
		Command: "/foo/bar 123",
		Timeout: timeout,
	})
	if err == nil {
		t.Error("there was no error")
	}

	if result.RC != NotFound {
		t.Errorf("Unexpected return code: %d, error: %s", result.RC, err)
	}

	js, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(js))
}

func TestCommandNotExecutable(t *testing.T) {
	timeout := 5 * time.Second
	result, err := RunCommand(context.Background(), CommandArgs{
		Command: testCommands.notExecutable,
		Timeout: timeout,
	})
	if err == nil {
		t.Errorf("there was no error")
	}

	if result.RC != NotExecutable {
		t.Errorf("return code == %d, expected NotExecutable, error: %s", result.RC, err)
	}
}

func TestCommandShlex(t *testing.T) {
	timeout := 5 * time.Second
	_, err := RunCommand(context.Background(), CommandArgs{
		Command: `blubb "sdf`,
		Timeout: timeout,
	})
	if err == nil {
		t.Fatal("there was no error")
	}
	if err.Error() != "EOF found when expecting closing quote" {
		t.Fatalf("Unexpected error: %s", err)
	}
}
