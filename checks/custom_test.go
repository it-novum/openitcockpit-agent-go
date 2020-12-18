package checks

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
)

func TestRun(t *testing.T) {
	cc := &CustomCheckRunner{
		Result: make(chan interface{}),
		Checks: []*config.CustomCheck{
			{
				Name:     "check_1",
				Interval: 1,
				Enabled:  true,
				Timeout:  30,
				Command:  "echo 'hallo welt'",
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
