package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/it-novum/openitcockpit-agent-go/platformpaths"
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
customchecks = /etc/openitcockpit-agent/customchecks.ini
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
customchecks = C:\Program Files\it-novum\openitcockpit-agent\customchecks.ini
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

var agentVersion1ConfigEmpty = ""

var agentConfigWithCustomCheck string = `[default]
customchecks = "%s"
`

var agentConfigWithCustomCheck2 string = `[default]
interval = 5
customchecks = "%s"
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

var customChecksAgentEmptyConfig string = ``

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
`

var customChecksAgentVersion1ConfigMissingCommand string = `
[time_1]
  command = "C:\checks\check_time.exe"
  interval = 60
  timeout = 10
  enabled = true

[no_command_line_at_all]
timeout = 10
enabled = true
`

func saveTempConfig(config string, customchecks bool) string {
	filename := "config.ini"
	if customchecks {
		filename = "customchecks.ini"
	}
	tmpDir, err := os.MkdirTemp(os.TempDir(), "*-test")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(config), 0600); err != nil {
		panic(err)
	}
	return tmpDir
}

func saveTempConfigWithCC(config string, customchecks string) string {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "*-test")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "config.ini"), []byte(fmt.Sprintf(config, filepath.Join(tmpDir, "customchecks.ini"))), 0600); err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "customchecks.ini"), []byte(customchecks), 0600); err != nil {
		panic(err)
	}
	return tmpDir
}

func TestAgentVersion1BlankConfig(t *testing.T) {
	cfgdir := saveTempConfig(agentVersion1ConfigBlank, false)
	defer os.RemoveAll(cfgdir)

	configPath := filepath.Join(cfgdir, "config.ini")
	c, err := Load(context.Background(), configPath)
	if err != nil {
		t.Fatal(err)
	}

	if c.CheckInterval != 30 {
		t.Error("Check interval expect to be 30")
	}

	if c.Port != 3333 {
		t.Error("WebServer port expect to be 3333")
	}

	if c.CustomchecksFilePath != "/etc/openitcockpit-agent/customchecks.ini" {
		t.Error("WebServer port expect to be /etc/openitcockpit-agent/customchecks.ini")
	}

	if c.CPU != true {
		t.Error("Checks CPU expect to be true")
	}

	if c.JmxUser != "monitorRole" {
		t.Error("Alfresco JmxUser expect to be monitorRole")
	}

	if c.OITC.Push != false {
		t.Error("Push Mode expect to be false")
	}

	js, _ := json.MarshalIndent(c, "", "    ")
	fmt.Println(string(js))
}

func TestAgentVersion1EmptyConfig(t *testing.T) {
	cfgdir := saveTempConfig(agentVersion1ConfigEmpty, false)
	defer os.RemoveAll(cfgdir)

	configPath := filepath.Join(cfgdir, "config.ini")
	c, err := Load(context.Background(), configPath)
	if err != nil {
		t.Fatal(err)
	}

	if c.CheckInterval != 30 {
		t.Error("Check interval expect to be 30")
	}

	if c.Port != 3333 {
		t.Error("WebServer port expect to be 3333")
	}

	if c.CustomchecksFilePath != filepath.Join(platformpaths.Get().ConfigPath(), "customchecks.ini") {
		t.Error("WebServer port expect to be: ", filepath.Join(platformpaths.Get().ConfigPath(), "customchecks.ini"))
	}

	if c.CPU != true {
		t.Error("Checks CPU expect to be true")
	}

	if c.JmxUser != "" {
		t.Error("Alfresco JmxUser expect to be monitorRole")
	}

	if c.OITC.Push != false {
		t.Error("Push Mode expect to be false")
	}

	js, _ := json.MarshalIndent(c, "", "    ")
	fmt.Println(string(js))

}

func TestAgentVersion1Config(t *testing.T) {
	cfgdir := saveTempConfig(agentVersion1Config, false)
	defer os.RemoveAll(cfgdir)

	configPath := filepath.Join(cfgdir, "config.ini")
	c, err := Load(context.Background(), configPath)
	if err != nil {
		t.Fatal(err)
	}

	if c.CheckInterval != 60 {
		t.Error("Check interval expect to be 60")
	}

	if c.Port != 33333 {
		t.Error("WebServer port expect to be 33333")
	}

	if c.CertificateFile != "/foo/bar.cert" {
		t.Error("TLS CertificateFile expect to be /foo/bar.cert")
	}

	if c.KeyFile != "/foo/bar.key" {
		t.Error("TLS KeyFile expect to be /foo/bar.key")
	}

	if c.AutoSslCsrFile != "/etc/autossl/csr.csr" {
		t.Error("TLS AutoSslCsrFile expect to be /etc/autossl/csr.csr")
	}

	if c.AutoSslCrtFile != "/etc/autossl/crt.crt" {
		t.Error("TLS AutoSslCrtFile expect to be /etc/autossl/crt.crt")
	}

	if c.AutoSslKeyFile != "/etc/autossl/key.key" {
		t.Error("TLS AutoSslKeyFile expect to be /etc/autossl/key.key")
	}

	if c.AutoSslCaFile != "/etc/autossl/server_ca.ca" {
		t.Error("TLS AutoSslCaFile expect to be /etc/autossl/server_ca.ca")
	}

	if c.CustomchecksFilePath != "C:\\Program Files\\it-novum\\openitcockpit-agent\\customchecks.ini" {
		t.Error("WebServer port expect to be C:\\Program Files\\it-novum\\openitcockpit-agent\\customchecks.ini")
	}

	if c.CPU != false {
		t.Error("Checks CPU expect to be false")
	}

	if c.JmxUser != "oitc-agent" {
		t.Error("Alfresco JmxUser expect to be oitc-agent")
	}

	if c.OITC.HostUUID != "3a2d91e5-03f7-4d2c-b719-05bd69b312ee" {
		t.Error("Push HostUUID expect to be 3a2d91e5-03f7-4d2c-b719-05bd69b312ee")
	}

	if c.OITC.URL != "https://demo.openitcockpit.io" {
		t.Error("Push url expect to be https://demo.openitcockpit.io")
	}

	if c.OITC.Apikey != "aaaaabbbbbcccccdddddeeeeefffff" {
		t.Error("Push Apikey expect to be aaaaabbbbbcccccdddddeeeeefffff")
	}

	if c.OITC.Proxy != "proxy.example.org" {
		t.Error("Push HostUUID expect to be proxy.example.org")
	}

	if c.OITC.Push != true {
		t.Error("Push Mode expect to be true")
	}

	if c.BasicAuth != "username:pass:word" {
		t.Error("BasicAuth username expect to be 'username'")
	}

	if !strings.Contains(c.OITC.AuthFile, "auth.json") {
		t.Error("auth.json file not set")
	}

	js, _ := json.MarshalIndent(c, "", "    ")
	fmt.Println(string(js))

}

func TestReadConfigFromFile(t *testing.T) {
	dir, _ := os.Getwd()
	configPath := fmt.Sprintf("%s%s../example/config_example.ini", dir, string(os.PathSeparator))
	_, err := Load(context.Background(), configPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadCustomChecksConfigAgentVersion1(t *testing.T) {
	// Read custom checks config
	cfgdir := saveTempConfig(customChecksAgentVersion1Config, true)
	defer os.RemoveAll(cfgdir)

	ccc, err := unmarshalCustomChecks(filepath.Join(cfgdir, "customchecks.ini"))
	if err != nil {
		t.Fatal(err)
	}

	if len(ccc) != 3 {
		t.Error("This config is expected to have 3 enabled custom checks")
	}

	for _, customcheck := range ccc {
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
	cfgdir := saveTempConfig(customChecksAgentVersion1ConfigEmptyCommand, true)
	defer os.RemoveAll(cfgdir)

	_, err := unmarshalCustomChecks(filepath.Join(cfgdir, "customchecks.ini"))
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "missing command") {
		t.Fatal("unxpected error: ", err)
	}
}

func TestReadCustomChecksConfigAgentVersion1MissingCommandline(t *testing.T) {
	cfgdir := saveTempConfig(customChecksAgentVersion1ConfigMissingCommand, true)
	defer os.RemoveAll(cfgdir)

	_, err := unmarshalCustomChecks(filepath.Join(cfgdir, "customchecks.ini"))
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "missing command") {
		t.Fatal("unxpected error: ", err)
	}
}

func TestReadCustomChecksConfigEmpty(t *testing.T) {
	cfgdir := saveTempConfig(customChecksAgentEmptyConfig, true)
	defer os.RemoveAll(cfgdir)

	ccc, err := unmarshalCustomChecks(filepath.Join(cfgdir, "customchecks.ini"))
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

	if len(ccc) != 0 {
		t.Fatal("unexpected number of custom checks (0): ", len(ccc))
	}
}

func TestReadAgentConfigWithCC(t *testing.T) {
	cfgdir := saveTempConfigWithCC(agentConfigWithCustomCheck, customChecksAgentVersion1Config)
	defer os.RemoveAll(cfgdir)

	configPath := filepath.Join(cfgdir, "config.ini")
	c, err := Load(context.Background(), configPath)
	if err != nil {
		t.Fatal(err)
	}
	ccc := c.CustomCheckConfiguration

	if len(ccc) < 1 {
		t.Fatal("unexpected number of custom checks (>0): ", len(ccc))
	}
}

func TestReadAgentConfigWithCCAndNewConfig(t *testing.T) {
	cfgdir := saveTempConfigWithCC(agentConfigWithCustomCheck, customChecksAgentVersion1Config)
	defer os.RemoveAll(cfgdir)

	configPath := filepath.Join(cfgdir, "config.ini")
	c, err := Load(context.Background(), configPath)
	if err != nil {
		t.Fatal(err)
	}
	ccc := c.CustomCheckConfiguration

	if len(ccc) < 1 {
		t.Fatal("unexpected number of custom checks (>0): ", len(ccc))
	}

	if err := c.SaveCustomCheckConfiguration([]byte(customChecksAgentEmptyConfig)); err != nil {
		t.Fatal(err)
	}
	if err := c.SaveConfiguration([]byte(agentConfigWithCustomCheck2)); err != nil {
		t.Fatal(err)
	}
	c, err = Load(context.Background(), configPath)
	if err != nil {
		t.Fatal(err)
	}
	ccc = c.CustomCheckConfiguration

	if c.CheckInterval != 5 {
		t.Error("reload did not work, unexpected check interval (5): ", c.CheckInterval)
	}
	if len(ccc) != 0 {
		t.Error("reload did not work, unexpected number of custom checks (0): ", len(ccc))
	}
}
