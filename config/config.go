package config

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"gopkg.in/ini.v1"
)

// AgentVersion as the name says
const AgentVersion = "2.1.0"

// Mode - Push or Pull
type Mode struct {
	Push bool `key:"enabled"`
}

// WebServer will hold the configuration options for the WebServer
type WebServer struct {
	Address string `key:"address"`
	Port    int64  `key:"port"`
}

// BasicAuth configuration options for the WebServer
type BasicAuth struct {
	Username string
	Password string
}

// TLS will hold the configuration options for the SSL
type TLS struct {
	AutoSslEnabled  bool   `key:"try-autossl"`
	CertificateFile string `key:"certfile"`
	KeyFile         string `key:"keyfile"`
	AutoSslFolder   string `key:"autossl-folder"`
	AutoSslCsrFile  string `key:"autossl-csr-file"`
	AutoSslCrtFile  string `key:"autossl-crt-file"`
	AutoSslKeyFile  string `key:"autossl-key-file"`
	AutoSslCaFile   string `key:"autossl-ca-file"`
}

// Push configuration of the Agent is running in push mode
type Push struct {
	HostUUID string `key:"hostuuid"`
	URL      string `key:"url"`
	Apikey   string `key:"apikey"`
	Proxy    string `key:"proxy"`
	Interval int64  `key:"interval"`
}

// Alfresco configuration
type Alfresco struct {
	JmxUser     string `key:"alfresco-jmxuser"`
	JmxPassword string `key:"alfresco-jmxpassword"`
	JmxAddress  string `key:"alfresco-jmxaddress"`
	JmxPort     int64  `key:"alfresco-jmxport"`
	JmxPath     string `key:"alfresco-jmxpath"`
	JmxQuery    string `key:"alfresco-jmxquery"`
	JavaPath    string `key:"alfresco-javapath"`
}

// WindowsEventLog with all event log types to monitor
type WindowsEventLog struct {
	types []string `key:"wineventlog-logtypes"`
}

// Checks which should be enabled and executed by the Agent
type Checks struct {
	Interval           int64  `key:"interval"`
	Docker             bool   `key:"dockerstats"`
	Qemu               bool   `key:"qemustats"`
	CPU                bool   `key:"cpustats"`
	Processes          bool   `key:"processstats"`
	Netstats           bool   `key:"netstats"`
	NetIo              bool   `key:"netio"`
	Diskstats          bool   `key:"diskstats"`
	DiskIo             bool   `key:"diskio"`
	WindowsServices    bool   `key:"winservices"`
	WindowsEventLog    bool   `key:"wineventlog"`
	SystemdServices    bool   `key:"systemdservices"`
	Alfresco           bool   `key:"alfrescostats"`
	CustomchecksConfig string `key:"customchecks"`
}

// Configuration with all sub configuration structs
type Configuration struct {
	Mode            *Mode
	WebServer       *WebServer
	BasicAuth       *BasicAuth
	TLS             *TLS
	Push            *Push
	Alfresco        *Alfresco
	WindowsEventLog *WindowsEventLog
	Checks          *Checks
}

func iterateFieldValues(obj interface{}, f func(reflect.StructField, reflect.Value)) {
	fields := reflect.TypeOf(obj)
	values := reflect.ValueOf(obj)
	if fields.Kind() == reflect.Ptr {
		fields = fields.Elem()
		values = values.Elem()
	}

	num := fields.NumField()
	for i := 0; i < num; i++ {
		field := fields.Field(i)
		value := values.Field(i)

		f(field, value)
	}
}

func mapConfig(obj interface{}, sectionName string, cfg *ini.File) {
	section := cfg.Section(sectionName)
	if section == nil {
		//log
		return
	}

	iterateFieldValues(obj, func(f reflect.StructField, v reflect.Value) {
		keyName := f.Tag.Get("key")
		if keyName == "" {
			return
		}

		key := section.Key(keyName)
		if key == nil {
			return
		}

		switch v.Kind() {
		case reflect.Int64:
			newValue, err := key.Int64()
			if err != nil {
				// log
			} else {
				v.SetInt(newValue)
			}
		case reflect.Float32, reflect.Float64:
		case reflect.String:
			newValue := key.String()
			v.SetString(newValue)
		case reflect.Bool:
			newValue, err := key.Bool()
			if err != nil {
				// log
			} else {
				v.SetBool(newValue)
			}
		default:
			panic(fmt.Errorf("can't assign value to a non-number type"))
		}
	})
}

// ReadConfig parse the given configuration string into the Configuration struct
func (c *Configuration) ReadConfig(config string) error {
	// Create all structs and set the default values
	c.Mode = &Mode{
		Push: false,
	}
	c.WebServer = &WebServer{
		Address: "0.0.0.0",
		Port:    3333,
	}
	c.BasicAuth = &BasicAuth{}
	c.TLS = &TLS{
		AutoSslEnabled: true,
	}
	c.Push = &Push{}
	c.Alfresco = &Alfresco{}
	c.WindowsEventLog = &WindowsEventLog{
		types: []string{"System", "Application", "Security"},
	}
	c.Checks = &Checks{
		Interval:           30,
		CustomchecksConfig: "/etc/openitcockpit-agent/customchecks.cnf",
		Docker:             false,
		Qemu:               false,
		CPU:                true,
		Processes:          true,
		Netstats:           true,
		NetIo:              true,
		Diskstats:          true,
		DiskIo:             true,
		WindowsServices:    true,
		WindowsEventLog:    true,
		SystemdServices:    true,
	}

	cfg, err := ini.Load([]byte(config))
	if err != nil {
		return err
	}

	mapConfig(c.Checks, "default", cfg)
	mapConfig(c.WebServer, "default", cfg)
	mapConfig(c.TLS, "default", cfg)
	mapConfig(c.Push, "oitc", cfg)
	mapConfig(c.Mode, "oitc", cfg)
	mapConfig(c.Alfresco, "default", cfg)

	if cfg.Section("default").Key("auth").String() != "" {
		auth := cfg.Section("default").Key("auth").String()
		authResult := strings.SplitN(auth, ":", 2)
		if len(authResult) == 2 {
			c.BasicAuth.Username = authResult[0]
			c.BasicAuth.Password = authResult[1]
		}
	}

	return nil

}

// ReadConfigFromFile reads the content of the passed ini file to pass it to ReadConfig()
func (c *Configuration) ReadConfigFromFile(path string) (config string, err error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
