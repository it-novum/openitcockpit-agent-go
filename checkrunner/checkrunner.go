package checkrunner

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/checks"
	"github.com/it-novum/openitcockpit-agent-go/config"
	log "github.com/sirupsen/logrus"
)

type CheckRunner struct {
	Configuration *config.Configuration
	Result        chan map[string]interface{}

	mtx      sync.Mutex
	wg       sync.WaitGroup
	shutdown chan struct{}
}

func (c *CheckRunner) Shutdown() {
	close(c.shutdown)
	c.wg.Wait()
}

func (c *CheckRunner) runChecks(parent context.Context, checks []checks.Check, timeout time.Duration) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	log.Infoln("Running ", len(checks), "checks")
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	done := make(chan struct{})

	results := make(map[string]interface{})
	go func() {
		for _, check := range checks {
			log.Debugln("Begin Check: ", check.Name())
			if result, err := check.Run(ctx); err != nil {
				log.Errorln("Check ", check.Name(), ": ", err)
			} else {
				results[check.Name()] = result
			}
			log.Debugln("Finish Check: ", check.Name())
		}
		// done maybe already to late
		select {
		case <-ctx.Done():
		case done <- struct{}{}:
		}
	}()

	select {
	case <-ctx.Done():
		err := ctx.Err()
		log.Errorln("Could not finish executing integrated checks: ", err)
		if errors.Is(err, context.DeadlineExceeded) {
			log.Errorln("Consider increasing check interval or disable unnecessary checks")
		}
		return
	case <-done:
	}

	select {
	case <-ctx.Done():
		err := ctx.Err()
		log.Errorln("Could not process integrated checks: ", err)
		if errors.Is(err, context.DeadlineExceeded) {
			log.Errorln("Consider increasing check interval or disable unnecessary checks")
		}
	case <-c.shutdown:
		log.Errorln("Check: canceled")
		return
	case c.Result <- results:
	}
}

// Start the check runner and returns immediatly (SHOULD NOT RUN IN GOROUTINE)
func (c *CheckRunner) Start(ctx context.Context) error {
	c.shutdown = make(chan struct{})

	checks, err := checks.ChecksForConfiguration(c.Configuration)
	if err != nil {
		return err
	}

	checkTimeout := time.Duration(c.Configuration.CheckInterval-1) * time.Second

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		ticker := time.NewTicker(time.Duration(c.Configuration.CheckInterval) * time.Second)
		defer ticker.Stop()

		go c.runChecks(ctx, checks, checkTimeout)
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-c.shutdown:
				if !ok {
					return
				}
			case <-ticker.C:
				go c.runChecks(ctx, checks, checkTimeout)
			}
		}
	}()

	return nil
}
