package cmd

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func dynamicPort() int64 {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		log.Fatal(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	port := int64(l.Addr().(*net.TCPAddr).Port)
	l.Close()
	return port
}

const exampleConfig = `[default]
port = "%d"
customchecks =
`

type testPlatformPath struct {
	tempPath   string
	configPath string
	logPath    string
}

func (p *testPlatformPath) Init() error {
	return nil
}

func (p *testPlatformPath) LogPath() string {
	return p.logPath
}

func (p *testPlatformPath) ConfigPath() string {
	return p.configPath
}

func (p *testPlatformPath) AdditionalData() map[string]string {
	return map[string]string{}
}

func (p *testPlatformPath) close() {
	os.RemoveAll(p.tempPath)
}

func newTestPath(t *testing.T, invalidLogDir bool) *testPlatformPath {
	var (
		err     error
		ok      bool
		tempDir = os.TempDir()
		tpp     = &testPlatformPath{}
	)

	tpp.tempPath, err = os.MkdirTemp(tempDir, "*-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if !ok {
			os.RemoveAll(tpp.tempPath)
		}
	}()
	if invalidLogDir {
		tpp.logPath = filepath.Join(tpp.tempPath, "nonexistent", "agent.log")
	} else {
		tpp.logPath = filepath.Join(tpp.tempPath, "agent.log")
	}

	tpp.configPath = filepath.Join(tpp.tempPath, "config.ini")
	err = os.WriteFile(tpp.configPath, []byte(fmt.Sprintf(exampleConfig, dynamicPort())), 0600)
	if err != nil {
		t.Fatal(err)
	}
	ok = true
	return tpp
}

func TestExecute(t *testing.T) {
	out := &bytes.Buffer{}
	tpp := newTestPath(t, false)
	defer tpp.close()

	okTests := [][]string{
		{"--help"},
		{"-h"},
		{"--config", tpp.ConfigPath()},
		{"-c", tpp.ConfigPath()},
		{"--verbose"},
		{"-v"},
		{},
	}
	for _, args := range okTests {
		r := New()
		r.platformPath = tpp
		r.cmd.SetArgs(args)
		r.cmd.SetOut(out)
		r.cmd.SetErr(out)
		done := make(chan struct{})
		go func() {
			err := r.Execute()
			if err != nil {
				t.Errorf("test failed \"%+v\": %s", args, err)
			}
			done <- struct{}{}
		}()
		time.Sleep(time.Second * 2)
		r.Shutdown()
		<-done
	}
}

func TestExecuteFail(t *testing.T) {
	out := &bytes.Buffer{}
	tpp := newTestPath(t, false)
	defer tpp.close()

	failTests := [][]string{
		{"unknown flag: --nonexisting", "--nonexisting"},
		{"unknown flag: --addtiional", "--config", tpp.ConfigPath(), "--addtiional"},
		{"unknown command \"dfasdf\"", "-c", tpp.ConfigPath(), "dfasdf"},
		{"--config \"someinvalidpath\" does not exist", "-v", "-c", "someinvalidpath"},
	}
	for _, tData := range failTests {
		args := tData[1:]
		r := New()
		r.platformPath = tpp
		r.cmd.SetArgs(args)
		r.cmd.SetOut(out)
		r.cmd.SetErr(out)
		err := r.Execute()
		if err == nil {
			t.Errorf("test failed \"%+v\"", args)
		} else {
			if !strings.Contains(err.Error(), tData[0]) {
				t.Errorf("Unexpected error: %s, expecting: %s", err, tData[0])
			}
		}
	}
}

func TestExecuteLogPathNotExists(t *testing.T) {
	out := &bytes.Buffer{}
	tpp := newTestPath(t, true)
	defer tpp.close()

	args := []string{"-c", tpp.configPath}
	r := New()
	r.platformPath = tpp
	r.cmd.SetArgs(args)
	r.cmd.SetOut(out)
	r.cmd.SetErr(out)
	err := r.Execute()
	if err == nil {
		t.Errorf("test failed \"%+v\"", args)
	} else if !strings.Contains(err.Error(), "Could not open/create log file") {
		t.Error("Unexpected error: ", err)
	}
}

func TestExecuteMissingPlatformPaths(t *testing.T) {
	out := &bytes.Buffer{}
	tpp := newTestPath(t, false)
	tpp.logPath = ""
	defer tpp.close()

	args := []string{}
	r := New()
	r.platformPath = tpp
	r.cmd.SetArgs(args)
	r.cmd.SetOut(out)
	r.cmd.SetErr(out)
	err := r.Execute()
	if err == nil {
		t.Errorf("test failed \"%+v\"", args)
	} else if !strings.Contains(err.Error(), "No log file path given") {
		t.Error("Unexpected error: ", err)
	}

	tpp.configPath = ""
	r = New()
	r.platformPath = tpp
	r.cmd.SetArgs(args)
	r.cmd.SetOut(out)
	r.cmd.SetErr(out)
	err = r.Execute()
	if err == nil {
		t.Errorf("test failed \"%+v\"", args)
	} else if !strings.Contains(err.Error(), "No config.ini path given") {
		t.Error("Unexpected error: ", err)
	}
}
