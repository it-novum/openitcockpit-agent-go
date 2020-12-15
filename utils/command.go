package utils

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/google/shlex"
)

// CommandResult to return the information
type CommandResult struct {
	Stdout string
	Stderr string
	RC     int
}

// Unified exit codes
const (
	Ok            = 0
	Warning       = 1
	Critical      = 2
	Unknown       = 3
	Timeout       = 124
	NotExecutable = 126
	NotFound      = 127
)

func runCommandJohannes(ctx context.Context, args []string, timeout time.Duration) {

}

func runCommand(commandStr string, timeout time.Duration) (*CommandResult, error) {
	result := &CommandResult{}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	args, err := shlex.Split(commandStr)
	if err != nil {
		result.RC = Unknown
		result.Stderr = err.Error()
		result.Stdout = err.Error()

		return result, err
	}

	outputBuf := &bytes.Buffer{}
	errorBuf := &bytes.Buffer{}

	c := exec.CommandContext(ctx, args[0], args[1:]...)
	c.Stdout = outputBuf
	c.Stderr = errorBuf

	// Do not hang forever
	// https://github.com/golang/go/issues/18874
	// https://github.com/golang/go/issues/22610
	go func() {
		<-ctx.Done()
		switch ctx.Err() {
		case context.DeadlineExceeded:
			if c.Process != nil {
				//Kill process because of timeout
				killProcess(c.Process)
			}
		case context.Canceled:
			//Process exited gracefully
			if c.Process != nil {
				c.Process.Kill()
			}
		}
	}()
	err = c.Run()

	if ctx.Err() == context.DeadlineExceeded {
		result.Stdout = fmt.Sprintf("Custom check %s timed out after %s seconds", strings.Join(args, " "), timeout.String())
		result.Stderr = fmt.Sprintf("Custom check %s timed out after %s seconds", strings.Join(args, " "), timeout.String())
		result.RC = Timeout
		return result, err
	}

	if err != nil && c.ProcessState == nil {
		if os.IsNotExist(err) {
			result.Stdout = fmt.Sprintf("No such file or directory: '%s'", strings.Join(args, " "))
			result.Stderr = fmt.Sprintf("No such file or directory: '%s'", strings.Join(args, " "))
			result.RC = NotFound
			return result, err
		}

		if os.IsPermission(err) {
			result.Stdout = fmt.Sprintf("File not executable: '%s'", strings.Join(args, " "))
			result.Stderr = fmt.Sprintf("File not executable: '%s'", strings.Join(args, " "))
			result.RC = NotExecutable
			return result, err
		}

		result.Stdout = fmt.Sprintf("Unknown error: %s Command: '%s'", err.Error(), strings.Join(args, " "))
		result.Stderr = fmt.Sprintf("Unknown error: %s Command: '%s'", err.Error(), strings.Join(args, " "))
		result.RC = Unknown
		return result, err
	}

	//No errors on command execution
	result.Stdout = outputBuf.String()
	result.Stderr = errorBuf.String()
	result.RC = Unknown

	state := c.ProcessState
	if status, ok := state.Sys().(syscall.WaitStatus); ok {
		result.RC = status.ExitStatus()
	}

	return result, nil
}
