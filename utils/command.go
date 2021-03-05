package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/shlex"
	"golang.org/x/text/encoding/unicode"
)

// CommandResult to return the information
type CommandResult struct {
	Stdout                    string `json:"stdout"`
	RC                        int    `json:"rc"`
	ExecutionUnixTimestampSec int64  `json:"execution_unix_timestamp_sec"`
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

type CommandArgs struct {
	Command       string
	Timeout       time.Duration
	Shell         string
	PowershellExe string
}

var (
	powershellCommand = []string{
		"powershell.exe",
		"-NoProfile",
		"-ExecutionPolicy",
		"unrestricted",
		"-NonInteractive",
		"-NoLogo",
		"-OutputFormat",
		"Text",
		"-File",
	}
	powershellCommandEncoded = []string{
		"powershell.exe",
		"-NoProfile",
		"-ExecutionPolicy",
		"unrestricted",
		"-NonInteractive",
		"-NoLogo",
		"-OutputFormat",
		"Text",
		"-EncodedCommand",
	}
	cmdCommand = []string{
		"cmd.exe",
		"/q",
		"/c",
	}
	vbsCommand = []string{
		"cscript.exe",
	}
	vbsArgs = []string{
		"/Nologo",
	}
	findDoubleBackslash = regexp.MustCompile(`(?m)(\\\\)`)
	findBackslash       = regexp.MustCompile(`(?m)(\\)`)
)

func encodePowershell(command string) (string, error) {
	enc := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
	encoded, err := enc.String(command)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString([]byte(encoded)), nil
}

func parseCommand(command, shell, powershellExe string) ([]string, string, error) {
	if runtime.GOOS == "windows" {
		command = findDoubleBackslash.ReplaceAllString(command, "\\")
		command = findBackslash.ReplaceAllString(command, "\\\\")
		args, err := shlex.Split(command)
		if err != nil {
			return nil, "", err
		}

		if shell != "" && shell != "powershell_command" && FileNotExists(args[0]) {
			return nil, "", fmt.Errorf("file not found: %s", args[0])
		}

		switch shell {
		case "powershell":
			res := ConcatStringSlice(powershellCommand, args)
			if powershellExe != "" {
				res[0] = powershellExe
			}
			return res, "", nil
		case "powershell_command":
			encoded, err := encodePowershell(command)
			if err != nil {
				return nil, "", fmt.Errorf("could not encode powershell command '%s': %s", command, err)
			}
			res := ConcatStringSlice(powershellCommandEncoded, []string{encoded})
			if powershellExe != "" {
				res[0] = powershellExe
			}
			return res, "", nil
		case "bat":
			return ConcatStringSlice(cmdCommand, args), "", nil
		case "vbs":
			return ConcatStringSlice(vbsCommand, vbsArgs, args), "", nil
		case "":
			return args, "", nil
		default:
			return nil, "", fmt.Errorf("unknown shell: %s", shell)
		}
	} else {
		if shell == "" {
			args, err := shlex.Split(command)
			return args, "", err
		} else {
			return []string{shell}, command, nil
		}
	}
}

// RunCommand in shell style with timeout on every platform
func RunCommand(ctx context.Context, commandArgs CommandArgs) (*CommandResult, error) {
	result := &CommandResult{
		ExecutionUnixTimestampSec: time.Now().Unix(),
	}
	var wg sync.WaitGroup
	defer wg.Wait()

	ctxTimeout, cancel := context.WithTimeout(ctx, commandArgs.Timeout)
	defer cancel()

	args, stdin, err := parseCommand(commandArgs.Command, commandArgs.Shell, commandArgs.PowershellExe)
	if err != nil {
		result.RC = Unknown
		result.Stdout = err.Error()

		return result, err
	}

	outputBuf := &bytes.Buffer{}
	stdinBuf := bytes.NewBufferString(stdin)

	c := exec.CommandContext(ctxTimeout, args[0], args[1:]...)
	c.Stdout = outputBuf
	// there is a bug in powershell where powershell prints an xml with the shell contents to stderr
	if commandArgs.Shell == "powershell_command" {
		nulBuf := &bytes.Buffer{}
		c.Stderr = nulBuf
	} else {
		c.Stderr = outputBuf
	}
	c.Stdin = stdinBuf

	c.SysProcAttr = commandSysproc

	// Do not hang forever
	// https://github.com/golang/go/issues/18874
	// https://github.com/golang/go/issues/22610
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctxTimeout.Done()
		switch ctxTimeout.Err() {
		case context.DeadlineExceeded:
			if c.Process != nil {
				//Kill process because of timeout
				// nolint:errcheck
				killProcessGroup(c.Process)
			}
		case context.Canceled:
			//Process exited gracefully
			if c.Process != nil {
				// nolint:errcheck
				killProcessGroup(c.Process)
			}
		}
	}()
	err = c.Run()

	if ctxTimeout.Err() == context.DeadlineExceeded {
		result.Stdout = fmt.Sprintf("Custom check %s timed out after %s seconds", strings.Join(args, " "), commandArgs.Timeout.String())
		result.RC = Timeout
		return result, err
	}

	if err != nil && c.ProcessState == nil {
		rc := handleCommandError(args[0], err)
		switch rc {
		case NotFound:
			result.Stdout = fmt.Sprintf("No such file or directory: '%s'", strings.Join(args, " "))
		case NotExecutable:
			result.Stdout = fmt.Sprintf("File not executable: '%s'", strings.Join(args, " "))
		default:
			result.Stdout = fmt.Sprintf("Unknown error: %s Command: '%s'", err.Error(), strings.Join(args, " "))
		}
		result.RC = rc
		return result, err
	}

	//No errors on command execution
	result.Stdout = outputBuf.String()
	result.RC = Unknown

	state := c.ProcessState
	if status, ok := state.Sys().(syscall.WaitStatus); ok {
		result.RC = status.ExitStatus()
	}

	return result, nil
}
