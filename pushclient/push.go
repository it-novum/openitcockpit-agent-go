package pushclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/it-novum/openitcockpit-agent-go/config"
	log "github.com/sirupsen/logrus"
)

type PushClient struct {
	StateInput chan []byte

	shutdown      chan struct{}
	wg            sync.WaitGroup
	configuration config.PushConfiguration
	client        http.Client
	url           *url.URL
	apiKeyHeader  string
	timeout       time.Duration

	state string
}

type pushData struct {
	CheckData string `json:"checkdata"`
	HostUUID  string `json:"hostuuid"`
}

func (p *PushClient) doRequest(parent context.Context) {
	log.Debugln("Push Client: new request")

	ctx, cancel := context.WithTimeout(parent, p.timeout)
	defer cancel()

	state := p.state
	if state == "" {
		log.Infoln("Push Client: No state to transfer")
		state = "{}"
	}

	data, err := json.Marshal(&pushData{
		CheckData: state,
		HostUUID:  p.configuration.HostUUID,
	})
	if err != nil {
		log.Errorln("Push Client: Could not serialize data for request: ", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.url.String(), bytes.NewReader(data))
	if err != nil {
		log.Errorln("Push Client: Could not create request: ", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", p.apiKeyHeader)

	res, err := p.client.Do(req)
	if err != nil {
		log.Errorln("Push Client: request error: ", err)
	}
	defer res.Body.Close()
	bodyStr := ""
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		bodyStr = ""
	} else {
		bodyStr = string(body)
	}

	if res.StatusCode != 200 {
		log.Errorln("Push Client: request status ", res.Status, ": ", bodyStr)
		return
	} else {
		log.Debugln("Push Client: request status ", res.Status, ": ", bodyStr)
	}

	log.Debugln("Push Client: request finished successful")
}

func (p *PushClient) Shutdown() {
	close(p.shutdown)
	p.wg.Wait()
}

// Run the server routine (should NOT be run in a go routine)
// You have to call Reload at least once to really start the webserver
func (p *PushClient) Start(ctx context.Context, cfg *config.Configuration) error {
	log.Debugln("Push Client: Starting")
	p.shutdown = make(chan struct{})
	p.configuration = *cfg.OITC

	//if p.configuration.PushInterval < 2 {
	//	return fmt.Errorf("Push Client: interval must be higher than 1")
	//}
	p.timeout = time.Duration(p.configuration.Timeout) * time.Second

	var (
		proxyURL *url.URL
		err      error
	)

	p.url, err = url.Parse(p.configuration.URL)
	p.url.Path = path.Join(p.url.Path, "agentconnector", "updateCheckdata.json")
	if err != nil {
		return err
	}
	p.apiKeyHeader = fmt.Sprint("X-OITC-API ", p.configuration.Apikey)

	if p.configuration.Proxy != "" {
		proxyURL, err = url.Parse(p.configuration.Proxy)
		if err != nil {
			return err
		}
	} else {
		req := &http.Request{
			URL: p.url,
		}
		proxyURL, err = http.ProxyFromEnvironment(req)
		if err != nil {
			return err
		}
	}

	transport := &http.Transport{}
	if proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	if !p.configuration.VerifyServerCertificate {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	p.client.Transport = transport

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case _, more := <-p.shutdown:
				if !more {
					return
				}
			case newState := <-p.StateInput:
				p.state = string(newState)
				p.doRequest(ctx)
			}
		}
	}()

	log.Debugln("Push Client: Start successful")
	return nil
}
