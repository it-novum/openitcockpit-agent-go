package checkrunner

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestCheckRunner(t *testing.T) {
	c := &CheckRunner{
		Configuration: &config.Configuration{
			CheckInterval: 30,
		},
		Result: make(chan map[string]interface{}),
	}
	err := c.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	select {
	case res := <-c.Result:
		b, err := json.MarshalIndent(res, "", "    ")
		if err != nil {
			t.Fatal(err)
		} else {
			log.Infoln(string(b))
		}
	case <-time.After(time.Second * 30):
		t.Fatal("timeout waiting for results")
	}

	c.Shutdown()
}

func TestCheckRunnerCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := &CheckRunner{
		Configuration: &config.Configuration{
			CheckInterval: 30,
		},
		Result: make(chan map[string]interface{}),
	}
	err := c.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cancel()

	done := make(chan struct{})

	go func() {
		c.wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-done:
		return
	case <-c.Result:
		t.Fatal("did not expect any result")
	case <-time.After(time.Second * 30):
		t.Fatal("timeout waiting for results")
	}
}
