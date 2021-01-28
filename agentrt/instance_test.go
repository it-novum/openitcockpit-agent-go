package agentrt

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"
	"time"
)

const exampleConfig = `[default]
customchecks = "%s"
`

const exampleCCConfigWin = `[check1]
shell = powershell_command
command = "Write-Host Test"
`

const exampleCCConfigNix = `[check1]
command = "echo 'hallo welt'"
`

func writeTestConfig(t *testing.T, tempDir string) {
	cfgPath := path.Join(tempDir, "config.cnf")
	cccPath := path.Join(tempDir, "customchecks.cnf")
	if err := ioutil.WriteFile(cfgPath, []byte(fmt.Sprintf(exampleConfig, cccPath)), 0600); err != nil {
		t.Fatal(err)
	}
	cccConfig := exampleCCConfigNix
	if runtime.GOOS == "windows" {
		cccConfig = exampleCCConfigWin
	}
	if err := ioutil.WriteFile(cccPath, []byte(cccConfig), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestAgentReload(t *testing.T) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "*-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	writeTestConfig(t, tempDir)

	rt := &AgentInstance{
		ConfigurationPath: path.Join(tempDir, "config.cnf"),
		LogPath:           path.Join(tempDir, "agent.log"),
		LogRotate:         3,
		Debug:             true,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rt.Start(ctx)
	for i := 0; i < 10; i++ {
		t.Log("Reload ", i)
		rt.Reload()
	}

	rt.Shutdown()
}

func TestAgentCancel(t *testing.T) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "*-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	writeTestConfig(t, tempDir)

	rt := &AgentInstance{
		ConfigurationPath: path.Join(tempDir, "config.cnf"),
		LogPath:           path.Join(tempDir, "agent.log"),
		LogRotate:         3,
		Debug:             true,
	}

	ctx, cancel := context.WithCancel(context.Background())

	rt.Start(ctx)
	rt.Reload()
	cancel()

	ticker := time.NewTicker(time.Microsecond * 200)
	defer ticker.Stop()
	timeout := time.NewTimer(time.Second * 10)
	defer timeout.Stop()
	select {
	case <-timeout.C:
		t.Fatal("timeout waiting for shutdown")
	case <-ticker.C:
		if rt.logHandler == nil {
			return
		}
	}
}
