package cmd

import (
	"bytes"
	"runtime"
	"testing"
)

var testConfigPath string

func init() {
	if runtime.GOOS == "windows" {
		testConfigPath = `C:\\Windows\\System32\\drivers\\etc\\hosts`
	} else {
		testConfigPath = "/etc/hosts"
	}
}

func TestExecute(t *testing.T) {
	out := &bytes.Buffer{}

	okTests := [][]string{
		{"--help"},
		{"-h"},
		{"--config", testConfigPath},
		{"-c", testConfigPath},
		{"--verbose", "-c", testConfigPath},
		{"-v", "-c", testConfigPath},
	}
	for _, args := range okTests {
		initCommand()
		rootCmd.SetArgs(args)
		rootCmd.SetOut(out)
		rootCmd.SetErr(out)
		err := Execute()
		if err != nil {
			t.Errorf("test failed \"%+v\": %s", args, err)
		}
	}
}

func TestExecuteFail(t *testing.T) {
	out := &bytes.Buffer{}

	okTests := [][]string{
		{"--nonexisting"},
		{"--config", testConfigPath, "--addtiional"},
		{"-c", testConfigPath, "dfasdf"},
		{"-v", "-c", "someinvalidpath"},
	}
	for _, args := range okTests {
		initCommand()
		rootCmd.SetArgs(args)
		rootCmd.SetOut(out)
		rootCmd.SetErr(out)
		err := Execute()
		if err == nil {
			t.Errorf("test failed \"%+v\"", args)
		}
	}
}
