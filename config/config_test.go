package config

import (
	"encoding/json"
	"fmt"
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

func TestAgentVersion1Config(t *testing.T) {
	c := &Configuration{}
	err := c.ReadConfig(agentVersion1ConfigBlank)
	if err != nil {
		t.Fatal(err)
	}

	if c.Checks.Interval != 30 {
		t.Error("Check interval has to be 30")
	}

	js, _ := json.MarshalIndent(c, "", "    ")
	fmt.Println(string(js))
}

func TestReadConfigFromFile(t *testing.T) {
	c := &Configuration{}
	err := c.ReadConfigFromFile("sdfdf")
	if err != nil {
		t.Fatal(err)
	}
}
