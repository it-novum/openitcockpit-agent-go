package wmiexecutor

import (
	"bytes"
	"os"
	"testing"
)

func TestEventlogCheckReadFromStdin(t *testing.T) {
	out := &bytes.Buffer{}

	stdin := `{"Age": 3600, "Logfiles": ["Application", "Security"]}`

	args := []string{
		"-c",
		"eventlog",
	}

	stdinp, w, _ := os.Pipe()

	go func() {
		_, _ = w.Write([]byte(stdin))
	}()

	os.Stdin = stdinp

	r := New()
	r.cmd.SetArgs(args)
	r.cmd.SetOut(out)
	r.cmd.SetErr(out)
	err := r.Execute()
	if err != nil {
		t.Errorf("test failed \"%+v\"", args)
	}

}
