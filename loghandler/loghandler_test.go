package loghandler

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	log "github.com/sirupsen/logrus"
)

func TestLogHandler(t *testing.T) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "*-test")
	if err != nil {
		log.Fatalln(err)
	}
	defer os.RemoveAll(tempDir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stderr := bytes.Buffer{}

	lh := &LogHandler{
		LogPath:       path.Join(tempDir, "agent.log"),
		LogRotate:     2,
		DefaultWriter: &stderr,
	}
	lh.Start(ctx)

	done := make(chan struct{})
	go func() {
		lh.Reload(&config.Configuration{
			Debug:   true,
			Verbose: true,
		})
		done <- struct{}{}
	}()

	select {
	case <-done:
		// ok
	case <-time.After(time.Second * 5):
		t.Fatal("timeout for reload of loghandler")
	}

	go func() {
		lh.Shutdown()
		done <- struct{}{}
	}()

	select {
	case <-done:
		// ok
	case <-time.After(time.Second * 5):
		t.Fatal("timeout for shutdown of loghandler")
	}
}

func TestLogHandlerRotate(t *testing.T) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "*-test")
	if err != nil {
		log.Fatalln(err)
	}
	defer os.RemoveAll(tempDir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stderr := bytes.Buffer{}

	lh := &LogHandler{
		LogPath:       path.Join(tempDir, "agent.log"),
		LogRotate:     2,
		DefaultWriter: &stderr,
	}
	midnight = func() time.Duration {
		return time.Second / 4
	}
	lh.Start(ctx)

	done := make(chan struct{})
	go func() {
		lh.Reload(&config.Configuration{
			Debug:   true,
			Verbose: true,
		})
		done <- struct{}{}
	}()

	select {
	case <-done:
		// ok
	case <-time.After(time.Second * 5):
		t.Fatal("timeout for reload of loghandler")
	}

	ticker := time.Tick(time.Second / 8)
	timeout := time.After(time.Second * 2)

outerfor:
	for {
		select {
		case <-timeout:
			t.Fatal("timeout for log rotate")
		case <-ticker:
			name1 := path.Join(tempDir, fmt.Sprintf("%s.%d", "agent.log", 1))
			name2 := path.Join(tempDir, fmt.Sprintf("%s.%d", "agent.log", 2))
			if _, err := os.Stat(name1); !os.IsNotExist(err) {
				if _, err := os.Stat(name2); !os.IsNotExist(err) {
					if _, err := os.Stat(path.Join(tempDir, "agent.log")); !os.IsNotExist(err) {
						break outerfor
					}
				}
			}
		}
	}

	go func() {
		lh.Shutdown()
		done <- struct{}{}
	}()

	select {
	case <-done:
		// ok
	case <-time.After(time.Second * 5):
		t.Fatal("timeout for shutdown of loghandler")
	}
}

func TestLogHandlerCancel(t *testing.T) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "*-test")
	if err != nil {
		log.Fatalln(err)
	}
	defer os.RemoveAll(tempDir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stderr := bytes.Buffer{}

	lh := &LogHandler{
		LogPath:       path.Join(tempDir, "agent.log"),
		LogRotate:     2,
		DefaultWriter: &stderr,
	}
	lh.Start(ctx)

	done := make(chan struct{})
	go func() {
		lh.Reload(&config.Configuration{
			Debug:   true,
			Verbose: true,
		})
		done <- struct{}{}
	}()

	select {
	case <-done:
		// ok
	case <-time.After(time.Second * 5):
		t.Fatal("timeout for reload of loghandler")
	}

	go func() {
		ticker := time.NewTicker(time.Second / 8)
		timeout := time.After(time.Second * 2)
		defer ticker.Stop()
		for {
			select {
			case <-timeout:
				return
			case <-ticker.C:
				if lh.logFile != nil {
					done <- struct{}{}
					return
				}
			}
		}
	}()

	select {
	case <-done:
		// ok
	case <-time.After(time.Second * 2):
		t.Fatal("timeout for loghandler load")
	}

	go func() {
		ticker := time.NewTicker(time.Second / 8)
		timeout := time.After(time.Second * 2)
		defer ticker.Stop()
		for {
			select {
			case <-timeout:
				return
			case <-ticker.C:
				if lh.logFile == nil {
					done <- struct{}{}
					return
				}
			}
		}
	}()

	cancel()

	select {
	case <-done:
		// ok
	case <-time.After(time.Second * 2):
		t.Fatal("timeout for cancel loghandler")
	}
}
