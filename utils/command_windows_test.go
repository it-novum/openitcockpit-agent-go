package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

type testWindowsShell struct {
	command  string
	suffix   string
	args     string
	shell    string
	expected string
	exitCode int
}

var (
	testWS = []testWindowsShell{
		{
			command: `Write-Host $args[0]
			exit 1
			`,
			suffix:   "ps1",
			shell:    "powershell",
			args:     "Test",
			expected: "Test",
			exitCode: 1,
		},
		{
			command:  `Write-Host Test`,
			shell:    "powershell_command",
			expected: "Test",
			exitCode: 0,
		},
		{
			command: `
			WScript.Echo WScript.Arguments.item(0)
			`,
			suffix:   "vbs",
			shell:    "vbs",
			args:     "Test",
			expected: "Test",
			exitCode: 0,
		},
		{
			command: `
			echo %*
			exit 1
			`,
			suffix:   "bat",
			shell:    "bat",
			args:     "Test",
			expected: "Test",
			exitCode: 1,
		},
	}
)

func TestWindowsShells(t *testing.T) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "*-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	writeCommand := func(command, suffix string) string {
		res := path.Join(tempDir, fmt.Sprintf("testcommand.%s", suffix))
		if err := ioutil.WriteFile(res, []byte(command), 0666); err != nil {
			t.Fatal(err)
		}
		return res
	}

	for i, test := range testWS {
		cmd := test.command
		if test.suffix != "" {
			cmd = writeCommand(cmd, test.suffix)
		}
		res, err := RunCommand(context.Background(), CommandArgs{
			Command: cmd + " " + test.args,
			Timeout: time.Second * 2,
			Shell:   test.shell,
		})
		if strings.TrimSpace(res.Stdout) != test.expected {
			t.Error("test ", i, " expected output: ", test.expected, " got: ", res.Stdout)
		}
		if res.RC != test.exitCode {
			t.Error("test ", i, " unexpected exit code: ", res.RC)
		}
		if err != nil && res.RC == Unknown {
			t.Error(err)
		}
	}
}
