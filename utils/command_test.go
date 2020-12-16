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

func TestRunPingCommand(t *testing.T) {
	command := "ping 127.0.0.1 -c 1"
	expected := "PING 127.0.0.1 (127.0.0.1)"
	if runtime.GOOS == "windows" {
		command = "ping 127.0.0.1 -n 1"
		expected = "127.0.0.1"
	}

	ctx, cancel := context.WithCancel(context.Background())
	timeout := 5 * time.Second
	result, err := RunCommand(ctx, command, timeout)
	if err != nil {
		t.Fatal("there was an error running ping")
	}

	if !strings.Contains(result.Stdout, expected) {
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
	result, err := RunCommand(context.Background(), "sleep 10", timeout)
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
		RunCommand(ctx, "sleep 10", timeout)
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
	result, err := RunCommand(context.Background(), "foobar 123", timeout)
	if err == nil {
		t.Fatal("there was no error")
	}

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
	timeout := 5 * time.Second
	result, err := RunCommand(context.Background(), "/foo/bar 123", timeout)
	if err == nil {
		t.Fatal("there was no error")
	}

	if result.RC != 127 {
		t.Error("Unexpected return code")
	}

	js, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(js))
}

func TestCommandNotExecutable(t *testing.T) {
	filename := "/etc/hosts"
	if runtime.GOOS == "windows" {
		filename = `C:\Windows\System32\drivers\etc\hosts`
	}

	timeout := 5 * time.Second
	result, err := RunCommand(context.Background(), filename, timeout)
	if err == nil {
		t.Fatal("there was no error")
	}

	if result.RC != NotExecutable {
		t.Fatal("return code != NotExecutable")
	}
}

func TestCommandShlex(t *testing.T) {
	timeout := 5 * time.Second
	_, err := RunCommand(context.Background(), `blubb "sdf`, timeout)
	if err == nil {
		t.Fatal("there was no error")
	}
	if err.Error() != "EOF found when expecting closing quote" {
		t.Fatalf("Unexpected error: %s", err)
	}
}
