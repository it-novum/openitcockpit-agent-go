package checks

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
)

func TestGetChecksToExecute(t *testing.T) {
	cc := &config.CustomChecks{
		WorkerThreads: 8,
		CustomChecks: []*config.CustomCheck{
			{
				Name:      "check_1",
				Interval:  10,
				Enabled:   true,
				Timeout:   30,
				LastCheck: time.Now().Add(-60 * time.Second),
				NextCheck: time.Now().Add(-50 * time.Second),
			},
		},
	}

	checksToExecute := GetChecksToExecute(cc)
	fmt.Println(checksToExecute)

}

func TestRun(t *testing.T) {
	cc := &CustomCheckRunner{
		Result: make(chan interface{}),
		Checks: []*config.CustomCheck{
			{
				Name:      "check_1",
				Interval:  1,
				Enabled:   true,
				Timeout:   30,
				LastCheck: time.Now().Add(-60 * time.Second),
				NextCheck: time.Now().Add(-50 * time.Second),
				Command:   "echo 'hallo welt'",
			},
		},
	}
	cc.Run(context.Background())

	timeout := time.After(time.Second * 3)
	for {
		select {
		case <-timeout:
			fmt.Println("timeout")
			go func() {
				cc.Shutdown()
				close(cc.Result)
			}()
		case res, ok := <-cc.Result:
			if !ok {
				return
			}
			fmt.Println(res)
		}
	}

}
