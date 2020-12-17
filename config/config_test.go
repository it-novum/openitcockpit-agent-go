package config

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

var agentVersion1ConfigBlank string = `[default]
interval = 30
# Port of the Agents build-in web server
port = 3333
address = 0.0.0.0
certfile =
keyfile =
try-autossl = true
autossl-csr-file = 
autossl-crt-file = 
autossl-key-file = 
autossl-ca-file = 
verbose = false
stacktrace = false
config-update-mode = false
auth =
customchecks = /etc/openitcockpit-agent/customchecks.cnf
temperature-fahrenheit = false
dockerstats = false
qemustats = false
cpustats = true
sensorstats = true
processstats = true
processstats-including-child-ids = false
netstats = true
diskstats = true
netio = true
diskio = true
winservices = true
systemdservices = true
wineventlog = true
wineventlog-logtypes = System, Application, Security
alfrescostats = false
alfresco-jmxuser = monitorRole
alfresco-jmxpassword = change_asap
alfresco-jmxaddress = 0.0.0.0
alfresco-jmxport = 50500
alfresco-jmxpath = /alfresco/jmxrmi
alfresco-jmxquery = 
alfresco-javapath = /usr/bin/java

[oitc]
enabled = false
hostuuid =
url = 
apikey =
proxy =
interval = 60
`

var agentVersion1Config string = `[default]
interval = 60
# Port of the Agents build-in web server
port = 33333
address = 127.0.0.1
certfile = /foo/bar.cert
keyfile = /foo/bar.key
try-autossl = true
autossl-csr-file =  /etc/autossl/csr.csr
autossl-crt-file = /etc/autossl/crt.crt
autossl-key-file = /etc/autossl/key.key
autossl-ca-file = /etc/autossl/server_ca.ca
verbose = false
stacktrace = false
config-update-mode = false
auth = username:pass:word
customchecks = C:\Program Files\it-novum\openitcockpit-agent\customchecks.cnf
temperature-fahrenheit = false
dockerstats = false
qemustats = false
cpustats = false
sensorstats = true
processstats = true
processstats-including-child-ids = false
netstats = true
diskstats = true
netio = true
diskio = true
winservices = true
systemdservices = true
wineventlog = true
wineventlog-logtypes = System, Application, Security
alfrescostats = false
alfresco-jmxuser = oitc-agent
alfresco-jmxpassword = change_asap
alfresco-jmxaddress = 0.0.0.0
alfresco-jmxport = 50500
alfresco-jmxpath = /alfresco/jmxrmi
alfresco-jmxquery = 
alfresco-javapath = /usr/bin/java

[oitc]
enabled = true
hostuuid = 3a2d91e5-03f7-4d2c-b719-05bd69b312ee
url = https://demo.openitcockpit.io
apikey = aaaaabbbbbcccccdddddeeeeefffff
proxy = proxy.example.org
interval = 90
`

var customChecksAgentVersion1ConfigStringMaxThreads string = `[default]
  max_worker_threads = abc
`

var customChecksAgentVersion1Config string = `[default]
  # max_worker_threads should be increased with increasing number of custom checks
  # but consider: each thread needs (a bit) memory
  max_worker_threads = 10

[time_1]
  command = "C:\checks\check_time.exe"
  interval = 60
  timeout = 10
  enabled = true

[check_Windows_Services_Status_OSS]
  command = powershell.exe -nologo -noprofile -File "C:\checks\check_Windows_Services_Status_OSS.ps1"
  interval = 15
  timeout = 10
  enabled = false

[check_ping]
  command = /usr/lib/nagios/plugins/check_ping -H 127.0.0.1 -w 100.0,20%% -c 500.0,60%% -p 5
  interval = 15
  timeout = 10
  enabled = true

[check_users]
  command = /usr/lib/nagios/plugins/check_users -w 3 -c 7
  interval = 15
  timeout = 10
  enabled = true
`

var customChecksAgentVersion1ConfigEmptyCommand string = `
[time_1]
  command = "C:\checks\check_time.exe"
  interval = 60
  timeout = 10
  enabled = true

[empty_command_line]
  command = 
  interval = 15
  timeout = 10
  enabled = false

  [no_command_line_at_all]
    timeout = 10
    enabled = true
`

func TestAgentVersion1BlankConfig(t *testing.T) {
	c := &Configuration{}
	err := c.ReadConfig(agentVersion1ConfigBlank)
	if err != nil {
		t.Fatal(err)
	}

	if c.Checks.Interval != 30 {
		t.Error("Check interval expect to be 30")
	}

	if c.WebServer.Port != 3333 {
		t.Error("WebServer port expect to be 3333")
	}

	if c.Checks.CustomchecksConfig != "/etc/openitcockpit-agent/customchecks.cnf" {
		t.Error("WebServer port expect to be /etc/openitcockpit-agent/customchecks.cnf")
	}

	if c.Checks.CPU != true {
		t.Error("Checks CPU expect to be true")
	}

	if c.Alfresco.JmxUser != "monitorRole" {
		t.Error("Alfresco JmxUser expect to be monitorRole")
	}

	if c.Mode.Push != false {
		t.Error("Push Mode expect to be false")
	}

	js, _ := json.MarshalIndent(c, "", "    ")
	fmt.Println(string(js))
}

func TestAgentVersion1Config(t *testing.T) {
	c := &Configuration{}
	err := c.ReadConfig(agentVersion1Config)
	if err != nil {
		t.Fatal(err)
	}

	if c.Checks.Interval != 60 {
		t.Error("Check interval expect to be 60")
	}

	if c.WebServer.Port != 33333 {
		t.Error("WebServer port expect to be 33333")
	}

	if c.TLS.CertificateFile != "/foo/bar.cert" {
		t.Error("TLS CertificateFile expect to be /foo/bar.cert")
	}

	if c.TLS.KeyFile != "/foo/bar.key" {
		t.Error("TLS KeyFile expect to be /foo/bar.key")
	}

	if c.TLS.AutoSslCsrFile != "/etc/autossl/csr.csr" {
		t.Error("TLS AutoSslCsrFile expect to be /etc/autossl/csr.csr")
	}

	if c.TLS.AutoSslCrtFile != "/etc/autossl/crt.crt" {
		t.Error("TLS AutoSslCrtFile expect to be /etc/autossl/crt.crt")
	}

	if c.TLS.AutoSslKeyFile != "/etc/autossl/key.key" {
		t.Error("TLS AutoSslKeyFile expect to be /etc/autossl/key.key")
	}

	if c.TLS.AutoSslCaFile != "/etc/autossl/server_ca.ca" {
		t.Error("TLS AutoSslCaFile expect to be /etc/autossl/server_ca.ca")
	}

	if c.Checks.CustomchecksConfig != "C:\\Program Files\\it-novum\\openitcockpit-agent\\customchecks.cnf" {
		t.Error("WebServer port expect to be C:\\Program Files\\it-novum\\openitcockpit-agent\\customchecks.cnf")
	}

	if c.Checks.CPU != false {
		t.Error("Checks CPU expect to be false")
	}

	if c.Alfresco.JmxUser != "oitc-agent" {
		t.Error("Alfresco JmxUser expect to be oitc-agent")
	}

	if c.Push.HostUUID != "3a2d91e5-03f7-4d2c-b719-05bd69b312ee" {
		t.Error("Push HostUUID expect to be 3a2d91e5-03f7-4d2c-b719-05bd69b312ee")
	}

	if c.Push.URL != "https://demo.openitcockpit.io" {
		t.Error("Push url expect to be https://demo.openitcockpit.io")
	}

	if c.Push.Apikey != "aaaaabbbbbcccccdddddeeeeefffff" {
		t.Error("Push Apikey expect to be aaaaabbbbbcccccdddddeeeeefffff")
	}

	if c.Push.Proxy != "proxy.example.org" {
		t.Error("Push HostUUID expect to be proxy.example.org")
	}

	if c.Push.Interval != 90 {
		t.Error("Push Interval expect to be 90")
	}

	if c.Mode.Push != true {
		t.Error("Push Mode expect to be true")
	}

	if c.BasicAuth.Username != "username" {
		t.Error("BasicAuth username expect to be 'username'")
	}

	if c.BasicAuth.Password != "pass:word" {
		t.Error("BasicAuth password expect to be 'pass:word'")
	}

	js, _ := json.MarshalIndent(c, "", "    ")
	fmt.Println(string(js))
}

func TestReadConfigFromFile(t *testing.T) {
	c := &Configuration{}
	dir, _ := os.Getwd()
	configPath := fmt.Sprintf("%s%s../config_example.ini", dir, string(os.PathSeparator))
	config, err := c.ReadConfigFromFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(config)
}

func TestReadCustomChecksConfigStringMaxThreads(t *testing.T) {
	//Read config.ini (required config.ini is an empty string)
	c := &Configuration{}
	err := c.ReadConfig("")
	if err != nil {
		t.Fatal(err)
	}
	err = c.ReadCustomChecksConfig(customChecksAgentVersion1ConfigStringMaxThreads)
	if err != nil {
		t.Fatal(err)
	}

	if c.CustomChecks.WorkerThreads != 8 {
		t.Error("CustomChecks.WorkerThreads expect to be 8")
	}

	if len(c.CustomChecks.CustomChecks) > 0 {
		t.Error("This config is expected to have no custom checks")
	}
}

func TestReadCustomChecksConfigAgentVersion1(t *testing.T) {
	//Read config.ini (required config.ini is an empty string)
	c := &Configuration{}
	err := c.ReadConfig("")
	if err != nil {
		t.Fatal(err)
	}

	// Read custom checks config
	err = c.ReadCustomChecksConfig(customChecksAgentVersion1Config)
	if err != nil {
		t.Fatal(err)
	}

	if c.CustomChecks.WorkerThreads != 10 {
		t.Error("CustomChecks.WorkerThreads expect to be 10")
	}

	if len(c.CustomChecks.CustomChecks) != 4 {
		t.Error("This config is expected to have 4 custom checks")
	}

	for _, customcheck := range c.CustomChecks.CustomChecks {
		if customcheck.Name == "time_1" {
			if customcheck.Enabled != true {
				t.Error("Custom check time_1 is expected to be enabled")
			}
		}

		if customcheck.Name == "check_Windows_Services_Status_OSS" {
			if customcheck.Enabled != false {
				t.Error("Custom check time_1 is expected to be disabled")
			}
		}
	}
}

func TestReadCustomChecksConfigAgentVersion1EmptyCommandline(t *testing.T) {
	//Read config.ini (required config.ini is an empty string)
	c := &Configuration{}
	err := c.ReadConfig("")
	if err != nil {
		t.Fatal(err)
	}

	// Read custom checks config
	err = c.ReadCustomChecksConfig(customChecksAgentVersion1ConfigEmptyCommand)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(c.CustomChecks.CustomChecks)
	if len(c.CustomChecks.CustomChecks) != 1 {
		t.Error("This config is expected to have 1 custom check")
	}
}
