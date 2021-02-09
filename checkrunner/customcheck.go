package checkrunner

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/utils"
	log "github.com/sirupsen/logrus"
)

type CustomCheckExecutor struct {
	Configuration *config.CustomCheck
	ResultOutput  chan *CustomCheckResult

	wg       sync.WaitGroup
	shutdown chan struct{}
}

func (c *CustomCheckExecutor) Shutdown() {
	close(c.shutdown)
	c.wg.Wait()
}

func (c *CustomCheckExecutor) runCheck(ctx context.Context, timeout time.Duration) {
	log.Debugln("Begin CustomCheck: ", c.Configuration.Name)
	result, err := utils.RunCommand(ctx, utils.CommandArgs{
		Command:       c.Configuration.Command,
		Timeout:       timeout,
		Shell:         c.Configuration.Shell,
		PowershellExe: c.Configuration.PowershellExe,
	})
	if err != nil && result.RC == utils.Unknown {
		log.Infoln("Custom check '", c.Configuration.Name, "' error: ", err)
	}
	select {
	case c.ResultOutput <- &CustomCheckResult{
		Name:   c.Configuration.Name,
		Result: result,
	}:
	case <-time.After(time.Second * 5):
		log.Errorln("Internal error: timeout could not save custom check result")
	}
	log.Debugln("Finish CustomCheck: ", c.Configuration.Name)
}

func (c *CustomCheckExecutor) Start(parent context.Context) error {
	c.shutdown = make(chan struct{})
	timeout := time.Duration(c.Configuration.Timeout) * time.Second
	interval := time.Duration(c.Configuration.Interval) * time.Second

	if timeout > interval {
		return errors.New("custom check timeout must be lower or equal to interval")
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		ctx, cancel := context.WithCancel(parent)
		defer cancel()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		c.runCheck(ctx, timeout)
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-c.shutdown:
				if !ok {
					return
				}
			case <-ticker.C:
				c.runCheck(ctx, timeout)
			}
		}
	}()

	return nil
}
