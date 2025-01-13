package checkrunner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	log "github.com/sirupsen/logrus"
)

type PrometheusCheckExecutor struct {
	Configuration *config.PrometheusExporter
	ResultOutput  chan *PrometheusExporterResult

	wg       sync.WaitGroup
	shutdown chan struct{}
}

func (c *PrometheusCheckExecutor) Shutdown() {
	close(c.shutdown)
	c.wg.Wait()
}

func (c *PrometheusCheckExecutor) runCheck(ctx context.Context, timeout time.Duration) {
	log.Debugln("Begin Prometheus Exporter: ", c.Configuration.Name)

	client := &http.Client{
		Timeout: timeout,
	}

	url := fmt.Sprintf("http://%s:%d%s", "localhost", c.Configuration.Port, c.Configuration.Path)
	resp, err := client.Get(url)
	if err != nil {
		log.Infoln("Prometheus Exporter '", c.Configuration.Name, "' error: ", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Infoln("Prometheus Exporter Error reading response body '", c.Configuration.Name, "' error: ", err)
		return
	}

	select {
	// Return custom check result to Agent Instance
	case c.ResultOutput <- &PrometheusExporterResult{
		Name:   c.Configuration.Name,
		Result: string(body),
	}:
	case <-time.After(time.Second * 5):
		log.Errorln("Internal error: timeout could not save Prometheus Exporter result")
	case <-c.shutdown:
		log.Errorln("Prometheus Exporte: canceled")
		return
	case <-ctx.Done():
		log.Errorln("Prometheus Exporte: canceled")
		return
	}
	log.Debugln("Finish Prometheus Exporte: ", c.Configuration.Name)
}

func (c *PrometheusCheckExecutor) Start(parent context.Context) error {
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
