package agentrt

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/checkrunner"
	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/loghandler"
	"github.com/it-novum/openitcockpit-agent-go/pushclient"
	"github.com/it-novum/openitcockpit-agent-go/webserver"
	log "github.com/sirupsen/logrus"
)

type AgentInstance struct {
	ConfigurationPath string
	LogPath           string
	LogRotate         int
	Verbose           bool
	Debug             bool

	wg       sync.WaitGroup
	shutdown chan struct{}
	reload   chan chan struct{}

	stateWebserver        chan []byte
	statePushClient       chan []byte
	checkResult           chan map[string]interface{}
	customCheckResultChan chan *checkrunner.CustomCheckResult

	customCheckResults map[string]interface{}

	logHandler         *loghandler.LogHandler
	webserver          *webserver.Server
	checkRunner        *checkrunner.CheckRunner
	customCheckHandler *checkrunner.CustomCheckHandler
	pushClient         *pushclient.PushClient
}

func (a *AgentInstance) processCheckResult(result map[string]interface{}) {
	for k, v := range a.customCheckResults {
		result[k] = v
	}

	data, err := json.Marshal(result)
	if err != nil {
		log.Errorln("Internal error: could not serialize check result: ", err)
		errorResult := map[string]string{
			"error": err.Error(),
		}
		data, err = json.Marshal(errorResult)
		if err != nil {
			log.Fatalln("Internal error: could also not serialize error result: ", err)
		}
	}
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		t := time.NewTimer(time.Second * 10)
		defer t.Stop()

		// we may have to give the webserver some time to think about it
		select {
		case a.stateWebserver <- data:
		case <-t.C:
			log.Errorln("Internal error: could not store check result: timeout")
		}
	}()
	if a.statePushClient != nil {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()

			t := time.NewTimer(time.Second * 10)
			defer t.Stop()

			// we may have to give the push client some time to think about it
			select {
			case a.statePushClient <- data:
			case <-t.C:
				log.Errorln("Internal error: could not store check result for push client: timeout")
			}
		}()
	}
}

func (a *AgentInstance) doReload(ctx context.Context, cfg *config.Configuration) {
	if a.stateWebserver == nil {
		a.stateWebserver = make(chan []byte)
	}
	if a.checkResult == nil {
		a.checkResult = make(chan map[string]interface{})
	}
	if a.webserver == nil {
		a.webserver = &webserver.Server{
			StateInput: a.stateWebserver,
			Reloader:   a,
		}
		a.webserver.Start(ctx)
	}
	a.webserver.Reload(cfg)

	if a.checkRunner != nil {
		a.checkRunner.Shutdown()
	}
	a.checkRunner = &checkrunner.CheckRunner{
		Configuration: cfg,
		Result:        a.checkResult,
	}
	if err := a.checkRunner.Start(ctx); err != nil {
		log.Fatalln(err)
	}
	if a.statePushClient != nil {
		a.pushClient.Shutdown()
		a.pushClient = nil
	}
	if cfg.OITC.Push {
		a.statePushClient = make(chan []byte)
		a.pushClient = &pushclient.PushClient{
			StateInput: a.statePushClient,
		}
		if err := a.pushClient.Start(ctx, cfg); err != nil {
			log.Fatalln("Could not load push client: ", err)
		}
	}
	a.doCustomCheckReload(ctx, cfg.CustomCheckConfiguration)
}

func (a *AgentInstance) doCustomCheckReload(ctx context.Context, ccc []*config.CustomCheck) {
	if a.customCheckHandler != nil {
		a.customCheckHandler.Shutdown()
		a.customCheckHandler = nil
	}
	if ccc != nil && len(ccc) > 0 {
		a.customCheckHandler = &checkrunner.CustomCheckHandler{
			Configuration: ccc,
			ResultOutput:  a.customCheckResultChan,
		}
		a.customCheckHandler.Start(ctx)
	}
}

func (a *AgentInstance) stop() {
	if a.logHandler != nil {
		a.logHandler.Shutdown()
		a.logHandler = nil
	}
	if a.webserver != nil {
		a.webserver.Shutdown()
		a.webserver = nil
	}
	if a.customCheckHandler != nil {
		a.customCheckHandler.Shutdown()
		a.customCheckHandler = nil
	}
	if a.checkRunner != nil {
		a.checkRunner.Shutdown()
		a.checkRunner = nil
	}
	if a.pushClient != nil {
		a.pushClient.Shutdown()
		a.pushClient = nil
	}
}

func (a *AgentInstance) Start(parent context.Context) {
	a.stateWebserver = make(chan []byte)
	a.checkResult = make(chan map[string]interface{})
	a.customCheckResultChan = make(chan *checkrunner.CustomCheckResult)
	a.customCheckResults = map[string]interface{}{}
	a.shutdown = make(chan struct{})
	a.reload = make(chan chan struct{})
	a.logHandler = &loghandler.LogHandler{
		Verbose:       a.Verbose,
		Debug:         a.Debug,
		LogPath:       a.LogPath,
		LogRotate:     a.LogRotate,
		DefaultWriter: os.Stderr,
	}

	ctx, cancel := context.WithCancel(parent)
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		defer cancel()

		a.logHandler.Start(ctx)

		defer a.stop()

		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-a.shutdown:
				if !ok {
					return
				}
			case done := <-a.reload:
				cfg, err := config.Load(ctx, &config.LoadConfigHint{
					ConfigFile: a.ConfigurationPath,
				})
				if err != nil {
					log.Fatalln("could not load configuration: ", err)
				}
				a.doReload(ctx, cfg)
				done <- struct{}{}
			case res := <-a.checkResult:
				a.processCheckResult(res)
			case res := <-a.customCheckResultChan:
				a.customCheckResults[res.Name] = res.Result
			}
		}
	}()

	a.Reload()
}

func (a *AgentInstance) Reload() {
	done := make(chan struct{})

	a.reload <- (done)
	<-done
}

func (a *AgentInstance) Shutdown() {
	close(a.shutdown)
	a.wg.Wait()
}
