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
	"github.com/it-novum/openitcockpit-agent-go/webserver"
	log "github.com/sirupsen/logrus"
)

type reloadConfig struct {
	Configuration            *config.Configuration
	CustomCheckConfiguration []*config.CustomCheck
	reloadDone               chan struct{}
}

type AgentInstance struct {
	ConfigurationPath string
	LogPath           string
	LogRotate         int
	Verbose           bool
	Debug             bool

	wg           sync.WaitGroup
	shutdown     chan struct{}
	reload       chan *reloadConfig
	configLoaded bool
	cccLoaded    bool

	stateInput            chan []byte
	checkResult           chan map[string]interface{}
	customCheckResultChan chan *checkrunner.CustomCheckResult

	customCheckResults map[string]interface{}

	logHandler         *loghandler.LogHandler
	webserver          *webserver.Server
	checkRunner        *checkrunner.CheckRunner
	customCheckHandler *checkrunner.CustomCheckHandler
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
		case a.stateInput <- data:
		case <-t.C:
			log.Errorln("Internal error: could not store check result: timeout")
		}
	}()
}

func (a *AgentInstance) doReload(ctx context.Context, cfg *reloadConfig) {
	if !a.configLoaded {
		// first load
		if cfg.Configuration.CustomchecksConfig != "" {
			go config.LoadCustomChecks(cfg.Configuration.CustomchecksConfig, func(ccc *config.CustomCheckConfiguration, err error) {
				if err != nil {
					if !a.cccLoaded {
						log.Fatalln(err)
					} else {
						log.Errorln("could not reload custom check configuration: ", err)
					}
				}
				if !a.cccLoaded {
					a.cccLoaded = true
				}
				a.ReloadCustomChecks(ccc.Checks)
			})
		}
		a.configLoaded = true
	}
	if a.stateInput == nil {
		a.stateInput = make(chan []byte)
	}
	if a.checkResult == nil {
		a.checkResult = make(chan map[string]interface{})
	}
	if a.webserver == nil {
		a.webserver = &webserver.Server{
			StateInput: a.stateInput,
		}
		a.webserver.Start(ctx)
	}
	a.webserver.Reload(cfg.Configuration)

	if a.checkRunner != nil {
		a.checkRunner.Shutdown()
	}
	a.checkRunner = &checkrunner.CheckRunner{
		Configuration: cfg.Configuration,
		Result:        a.checkResult,
	}
	if err := a.checkRunner.Start(ctx); err != nil {
		log.Fatalln(err)
	}
}

func (a *AgentInstance) doCustomCheckReload(ctx context.Context, ccc []*config.CustomCheck) {
	if a.customCheckHandler != nil {
		a.customCheckHandler.Shutdown()
	}
	a.customCheckHandler = &checkrunner.CustomCheckHandler{
		Configuration: ccc,
		ResultOutput:  a.customCheckResultChan,
	}
	a.customCheckHandler.Start(ctx)
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
}

func (a *AgentInstance) configLoad(cfg *config.Configuration, err error) {
	if err != nil {
		if !a.configLoaded {
			log.Fatalln(err)
		} else {
			log.Errorln("could not reload configuration: ", err)
			return
		}
	}
	a.Reload(cfg)
}

func (a *AgentInstance) Start(parent context.Context) {
	a.stateInput = make(chan []byte)
	a.checkResult = make(chan map[string]interface{})
	a.customCheckResultChan = make(chan *checkrunner.CustomCheckResult)
	a.customCheckResults = map[string]interface{}{}
	a.shutdown = make(chan struct{})
	a.reload = make(chan *reloadConfig)
	a.logHandler = &loghandler.LogHandler{
		Verbose:       a.Verbose,
		Debug:         a.Debug,
		LogPath:       a.LogPath,
		LogRotate:     a.LogRotate,
		DefaultWriter: os.Stderr,
	}

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		ctx, cancel := context.WithCancel(parent)
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
			case cfg := <-a.reload:
				if cfg.Configuration != nil {
					a.doReload(ctx, cfg)
				}
				if cfg.CustomCheckConfiguration != nil {
					a.doCustomCheckReload(ctx, cfg.CustomCheckConfiguration)
				}
				cfg.reloadDone <- struct{}{}
			case res := <-a.checkResult:
				a.processCheckResult(res)
			case res := <-a.customCheckResultChan:
				a.customCheckResults[res.Name] = res.Result
			}
		}
	}()

	config.Load(a.configLoad, &config.LoadConfigHint{
		ConfigFile: a.ConfigurationPath,
	})
}

func (a *AgentInstance) Reload(cfg *config.Configuration) {
	done := make(chan struct{})
	a.reload <- &reloadConfig{
		Configuration: cfg,
		reloadDone:    done,
	}
	<-done
}

func (a *AgentInstance) ReloadCustomChecks(ccc []*config.CustomCheck) {
	done := make(chan struct{})
	a.reload <- &reloadConfig{
		CustomCheckConfiguration: ccc,
		reloadDone:               done,
	}
	<-done
}

func (a *AgentInstance) Shutdown() {
	close(a.shutdown)
	a.wg.Wait()
}
