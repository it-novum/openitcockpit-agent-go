package checks

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/utils"
)

// CustomCheckRunner runs custom checks
type CustomCheckRunner struct {
	// Result channel for check results
	// Do not close before Shutdown completes
	Result   chan interface{}
	Checks   []*config.CustomCheck
	shutdown chan struct{}
	wg       sync.WaitGroup
}

// Run the custom checks in background (DO NOT RUN IN GO ROUTINE)
func (c *CustomCheckRunner) Run(parentCtx context.Context) {
	c.wg.Add(1)
	c.shutdown = make(chan struct{})

	go func() {
		defer c.wg.Done()

		ctx, cancel := context.WithCancel(parentCtx)
		defer cancel()

		checkPipe := make(chan *config.CustomCheck)
		for _, loopCheck := range c.Checks {
			if loopCheck.Enabled == true {
				c.wg.Add(1)
				go func() {
					check := <-checkPipe
					defer c.wg.Done()
					ticker := time.NewTicker(time.Duration(check.Interval) * time.Second)
					defer ticker.Stop()
					for {
						select {
						case <-ctx.Done():
							return
						case <-ticker.C:
							//ausführen
							result, err := utils.RunCommand(ctx, check.Command, time.Duration(check.Timeout)*time.Second)
							if err != nil {
								log.Println(err)
							}
							c.Result <- result
						}
					}
				}()
			}
			checkPipe <- loopCheck
		}

		select {
		case <-ctx.Done():
		case <-c.shutdown:
			cancel()
		}
	}()
}

// Shutdown custom check runner, waits for completion
func (c *CustomCheckRunner) Shutdown() {
	c.shutdown <- struct{}{}
	c.wg.Wait()
}
