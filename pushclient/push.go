package pushclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/it-novum/openitcockpit-agent-go/config"
	"github.com/it-novum/openitcockpit-agent-go/utils"
	log "github.com/sirupsen/logrus"
)

func addressForIPPort(ipport string) string {
	return ipport[:strings.LastIndex(ipport, ":")]
}

func fetchSystemInformation() (string, string) {
	var (
		hostname  string
		ipaddress string
	)

	if name, err := os.Hostname(); err != nil {
		log.Errorln("Push Client: this server has no hostname")
	} else {
		hostname = name
	}

	conn, err := net.Dial("udp", "[2001:db8::1]:49151")
	if err != nil {
		log.Debugln("Push Client: creating udp6 connection for ip check failed")
	} else {
		ipaddress = addressForIPPort(conn.LocalAddr().String())
		_ = conn.Close()
	}

	conn, err = net.Dial("udp", "198.51.100.1:49151")
	if err != nil {
		log.Debugln("Push Client: creating udp connection for ip check failed")
	} else {
		ipaddress = addressForIPPort(conn.LocalAddr().String())
		_ = conn.Close()
	}

	return hostname, ipaddress
}

type authConfiguration struct {
	UUID     string `json:"uuid"`
	Password string `json:"password"`
}

type PushClient struct {
	StateInput chan []byte

	shutdown           chan struct{}
	wg                 sync.WaitGroup
	configuration      config.PushConfiguration
	authConfiguration  authConfiguration
	client             http.Client
	urlSubmitCheckData *url.URL
	urlRegisterAgent   *url.URL
	apiKeyHeader       string
	timeout            time.Duration
}

type registerAgentRequest struct {
	AgentUUID string `json:"agentuuid"`
	Password  string `json:"password"`
	Hostname  string `json:"hostname"`
	IPAddress string `json:"ipaddress"`
}

type registerAgentResponse struct {
	AgentUUID string `json:"agentuuid"`
	Password  string `json:"password"`
	Error     string `json:"error"`
}

type submitCheckDataRequest struct {
	CheckData *json.RawMessage `json:"checkdata"`
	AgentUUID string           `json:"agentuuid"`
	Password  string           `json:"password"`
}

type submitCheckDataResponse struct {
	ReceivedChecks int64  `json:"received_checks"`
	Error          string `json:"error"`
}

func (p *PushClient) saveAuthConfig() error {
	data, err := json.Marshal(&p.authConfiguration)
	if err != nil {
		return fmt.Errorf("could not write push client auth file: %s", err)
	}
	if err := ioutil.WriteFile(p.configuration.AuthFile, data, 0600); err != nil {
		return fmt.Errorf("could not write push client auth file: %s", err)
	}
	return nil
}

func (p *PushClient) readAuthConfig() error {
	if utils.FileExists(p.configuration.AuthFile) {
		data, err := ioutil.ReadFile(p.configuration.AuthFile)
		if err != nil {
			return fmt.Errorf("could not read push client auth file: %s", err)
		}
		if err := json.Unmarshal(data, &p.authConfiguration); err != nil {
			return fmt.Errorf("could not read push client auth file: %s", err)
		}
	}

	if p.authConfiguration.UUID == "" {
		p.authConfiguration.UUID = uuid.NewString()
		return p.saveAuthConfig()
	}
	return nil
}

func (p *PushClient) httpRequest(ctx context.Context, url *url.URL, sendJson interface{}, result interface{}) (int, error) {
	data, err := json.Marshal(sendJson)
	if err != nil {
		return 0, errors.Wrap(err, "could not serialize data for request")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewReader(data))
	if err != nil {
		return 0, errors.Wrap(err, "could not create request")
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", p.apiKeyHeader)

	res, err := p.client.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "request failed")
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, errors.Wrap(err, "reading response body from server was not successful")
	}
	log.Debugln("Push Client: Response status from server: ", res.StatusCode)
	if len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return 0, errors.Wrap(err, "could not unmarshal server response")
		}
	}

	return res.StatusCode, nil
}

func (p *PushClient) registerClient(ctx context.Context, state []byte) {
	log.Infoln("Push Client: register client at server")

	log.Debugln("Push Client: test for write permissions on auth configuration")
	if err := p.saveAuthConfig(); err != nil {
		log.Errorln("Push Client: unable to write client auth configuration: ", err)
		return
	}

	hostname, ipaddress := fetchSystemInformation()
	req := registerAgentRequest{
		AgentUUID: p.authConfiguration.UUID,
		Password:  p.authConfiguration.Password,
		Hostname:  hostname,
		IPAddress: ipaddress,
	}
	res := registerAgentResponse{}

	log.Debugln("Push Client: send request")
	status, err := p.httpRequest(ctx, p.urlRegisterAgent, &req, &res)
	if err != nil {
		log.Errorln("Push Client: ", err)
		return
	}
	switch status {
	case 405:
		log.Errorln("Push Client: authentication error (probably incorrect api key)")
		return
	case 403:
		log.Errorln("Push Client: this agent was already registered with a different password, you have to delete it in openITCOCKPIT and re-register it")
		return
	case 201:
		if res.AgentUUID != p.authConfiguration.UUID {
			log.Errorln("Push Client: unexpected agentuuid in server response during registration: ", res.AgentUUID)
			return
		}
		if res.Password == "" {
			log.Infoln("Push Client: Waiting for registration on the server")
			return
		}
		p.authConfiguration.Password = res.Password
		if err := p.saveAuthConfig(); err != nil {
			log.Errorln("Push Client: unable to write client auth configuration: ", err)
			p.authConfiguration.Password = ""
			return
		}
		log.Infoln("Push Client: server registration successful")
		p.submitCheckData(ctx, state)
		return
	case 200:
		if res.AgentUUID != p.authConfiguration.UUID || res.Password != p.authConfiguration.Password {
			log.Errorln("Push Client: server returned unexpected uuid or password for this agent: ", res.AgentUUID, ":", res.Password)
			return
		}
	default:
		if res.Error != "" {
			log.Errorln("Push Client: could not register client: ", res.Error)
		} else {
			log.Errorln("Push Client: unknown error during client registration, http status: ", status)
		}
		return
	}
}

// The interval of submitCheckData is defined by the check interval
// due to submitCheckData get's triggered when new check results are available
// custom checks get merged into the "normal" checl results so custom checks do not trigger this function.
// Only the check interval of the inbuild will trigger this.
func (p *PushClient) submitCheckData(ctx context.Context, state []byte) {
	log.Infoln("Push Client: send new state to server")

	if len(state) < 1 {
		state = []byte("{}")
	}

	checkData := json.RawMessage(state)

	req := submitCheckDataRequest{
		CheckData: &checkData,
		AgentUUID: p.authConfiguration.UUID,
		Password:  p.authConfiguration.Password,
	}
	res := submitCheckDataResponse{}

	status, err := p.httpRequest(ctx, p.urlSubmitCheckData, &req, &res)
	if err != nil {
		log.Errorln("Push client: ", err)
		return
	}

	switch status {
	case 405:
		log.Errorln("Push Client: authentication error (probably incorrect api key)")
		return
	case 200:
		log.Debugln("Push Client: submitted ", res.ReceivedChecks, " checks")
		return
	default:
		if res.Error != "" {
			log.Errorln("Push Client: could not send state to server: ", res.Error)
		} else {
			log.Errorln("Push Client: unknown error during submit checkdata, http status: ", status)
		}
		return
	}
}

func (p *PushClient) updateState(parent context.Context, state []byte) {
	log.Debugln("Push Client: new request")

	ctx, cancel := context.WithTimeout(parent, p.timeout)
	defer cancel()

	if p.authConfiguration.Password != "" {
		p.submitCheckData(ctx, state)
	} else {
		p.registerClient(ctx, state)
	}
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

	if err := p.readAuthConfig(); err != nil {
		return err
	}

	p.timeout = time.Duration(p.configuration.Timeout) * time.Second

	var (
		proxyURL *url.URL
		err      error
	)

	p.urlSubmitCheckData, err = url.Parse(p.configuration.URL)
	if err != nil {
		return err
	}
	p.urlSubmitCheckData.Path = path.Join(p.urlSubmitCheckData.Path, "agentconnector", "submit_checkdata.json")

	p.urlRegisterAgent, err = url.Parse(p.configuration.URL)
	if err != nil {
		return err
	}
	p.urlRegisterAgent.Path = path.Join(p.urlRegisterAgent.Path, "agentconnector", "register_agent.json")

	p.apiKeyHeader = fmt.Sprint("X-OITC-API ", p.configuration.Apikey)

	if p.configuration.Proxy != "" {
		proxyURL, err = url.Parse(p.configuration.Proxy)
		if err != nil {
			return err
		}
	} else {
		req := &http.Request{
			URL: p.urlSubmitCheckData,
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
				// received new check results
				p.updateState(ctx, newState)
			}
		}
	}()

	log.Debugln("Push Client: Start successful")
	return nil
}
