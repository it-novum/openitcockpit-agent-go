package wmiexecutor

import (
	"bytes"
	"strings"
	"testing"
)

func TestExecuteMissingCommand(t *testing.T) {
	out := &bytes.Buffer{}

	argsToTest := [][]string{
		{"-c", "foobar"},
		{"--command", "foobar"},
	}

	for _, args := range argsToTest {
		r := New()
		r.cmd.SetArgs(args)
		r.cmd.SetOut(out)
		r.cmd.SetErr(out)
		err := r.Execute()
		if err == nil {
			t.Errorf("test failed \"%+v\"", args)
		} else if !strings.Contains(err.Error(), "No command 'foobar'") {
			t.Error("Unexpected error: ", err)
		}
	}
}
