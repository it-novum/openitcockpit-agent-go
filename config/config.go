package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/it-novum/openitcockpit-agent-go/platformpaths"
	"github.com/it-novum/openitcockpit-agent-go/utils"
	"github.com/prometheus/common/log"
	"github.com/spf13/viper"
)

// AgentVersion as the name says
const AgentVersion = "3.0.0"

// CustomCheck are external plugins and scripts which should be executed by the Agent
type CustomCheck struct {
	Name     string `mapstructure:"-"`
	Interval int64  `mapstructure:"interval"`
	Enabled  bool   `mapstructure:"enabled"`
	Command  string `mapstructure:"command"`
	Timeout  int64  `mapstructure:"timeout"`
	// Linux/Darwin = if set run shell and pipe command into it
	// Windows => powershell -> start powershell, command must be path to powershell file
	// Windows => powershell_command -> start powershell, run command directly
	// Windows => bat -> start cmd call, command must be path to bat file
	// Windows => vbs -> wscript, command must be path to vbs file
	// if not set the command will be just executed as it is
	Shell string `mapstructure:"shell"`
}

type PushConfiguration struct {
	Push                    bool   `mapstructure:"enabled"`
	HostUUID                string `mapstructure:"hostuuid"`
	URL                     string `mapstructure:"url"`
	Apikey                  string `mapstructure:"apikey"`
	Proxy                   string `mapstructure:"proxy"`
	Timeout                 int64  `mapstructure:"timeout"`
	VerifyServerCertificate bool   `mapstructure:"verify-server-certificate"`
	EnableWebserver         bool   `mapstructure:"enable-webserver"`
	// Stores authentication information generated by push client
	AuthFile string `mapstructure:"authfile"`
}

// Configuration with all sub configuration structs
type Configuration struct {
	ConfigurationPath string `json:"-" mapstructure:"-"`
	viper             *viper.Viper

	// TLS

	AutoSslEnabled  bool   `mapstructure:"try-autossl"`
	CertificateFile string `mapstructure:"certfile"`
	KeyFile         string `mapstructure:"keyfile"`
	AutoSslFolder   string `mapstructure:"autossl-folder"`
	AutoSslCsrFile  string `mapstructure:"autossl-csr-file"`
	AutoSslCrtFile  string `mapstructure:"autossl-crt-file"`
	AutoSslKeyFile  string `mapstructure:"autossl-key-file"`
	AutoSslCaFile   string `mapstructure:"autossl-ca-file"`

	// Webserver

	Address   string `mapstructure:"address"`
	Port      int64  `mapstructure:"port"`
	BasicAuth string `mapstructure:"auth"`

	// Config Misc

	ConfigUpdate         bool   `mapstructure:"config-update-mode"`
	CustomchecksFilePath string `mapstructure:"customchecks"`

	// Default Checks

	CheckInterval   int64 `mapstructure:"interval"`
	Docker          bool  `mapstructure:"dockerstats"`
	Qemu            bool  `mapstructure:"qemustats"`
	CPU             bool  `mapstructure:"cpustats"`
	Load            bool  `mapstructure:"load"`
	Memory          bool  `mapstructure:"memory"`
	Processes       bool  `mapstructure:"processstats"`
	Netstats        bool  `mapstructure:"netstats"`
	NetIo           bool  `mapstructure:"netio"`
	Sensors         bool  `mapstructure:"sensors"`
	Diskstats       bool  `mapstructure:"diskstats"`
	DiskIo          bool  `mapstructure:"diskio"`
	Swap            bool  `mapstructure:"swap"`
	User            bool  `mapstructure:"userstats"`
	WindowsServices bool  `mapstructure:"winservices"`
	WindowsEventLog bool  `mapstructure:"wineventlog"`
	SystemdServices bool  `mapstructure:"systemdservices"`
	LaunchdServices bool  `mapstructure:"launchdservices"`
	Alfresco        bool  `mapstructure:"alfrescostats"`
	Libvirt         bool  `mapstructure:"libvirt"`

	// Alfresco

	JmxUser     string `mapstructure:"alfresco-jmxuser"`
	JmxPassword string `mapstructure:"alfresco-jmxpassword"`
	JmxAddress  string `mapstructure:"alfresco-jmxaddress"`
	JmxPort     int64  `mapstructure:"alfresco-jmxport"`
	JmxPath     string `mapstructure:"alfresco-jmxpath"`
	JmxQuery    string `mapstructure:"alfresco-jmxquery"`
	JavaPath    string `mapstructure:"alfresco-javapath"`

	// WindowsEventLog with all event log types to monitor

	WindowsEventLogTypes []string `mapstructure:"wineventlog-logtypes"`

	// Push Mode

	OITC *PushConfiguration `json:"oitc"`

	// Default is the namespace workaround we need for the configuration file format
	Default *Configuration `json:"-"`

	CustomCheckConfiguration []*CustomCheck `json:"customchecks_configuration" mapstructure:"-"`
}

var defaultValue = map[string]interface{}{
	"port":                 3333,
	"interval":             30,
	"qemustats":            true,
	"cpustats":             true,
	"load":                 true,
	"memory":               true,
	"processstats":         true,
	"netstats":             true,
	"netio":                true,
	"sensors":              true,
	"diskstats":            true,
	"diskio":               true,
	"swap":                 true,
	"userstats":            true,
	"winservices":          true,
	"wineventlog":          true,
	"systemdservices":      true,
	"alfrescostats":        true,
	"libvirt":              true,
	"wineventlog-logtypes": "System,Application,Security",
	"customchecks":         path.Join(platformpaths.Get().ConfigPath(), "customchecks.cnf"),
	"autossl-folder":       platformpaths.Get().ConfigPath(),
	"autossl-csr-file":     path.Join(platformpaths.Get().ConfigPath(), "agent.csr"),
	"autossl-crt-file":     path.Join(platformpaths.Get().ConfigPath(), "agent.crt"),
	"autossl-key-file":     path.Join(platformpaths.Get().ConfigPath(), "agent.key"),
	"autossl-ca-file":      path.Join(platformpaths.Get().ConfigPath(), "server_ca.crt"),
}

var oitcDefaultvalue = map[string]interface{}{
	"authfile": path.Join(platformpaths.Get().ConfigPath(), "auth.cnf"),
}

func setConfigurationDefaults(v *viper.Viper) {
	for key, value := range defaultValue {
		v.SetDefault("default."+key, value)
	}

	for key, value := range oitcDefaultvalue {
		v.SetDefault("oitc."+key, value)
	}
}

// LoadConfigHint for Load func
type LoadConfigHint struct {
	SearchPath string
	ConfigFile string
}

func unmarshalConfiguration(v *viper.Viper) (*Configuration, error) {
	cfg := &Configuration{}
	cfg.Default = cfg
	cfg.OITC = &PushConfiguration{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}
	if cfg.OITC.Push && cfg.OITC.Timeout == 0 {
		if cfg.CheckInterval <= 1 {
			cfg.OITC.Timeout = 1
		} else {
			cfg.OITC.Timeout = cfg.CheckInterval - 1
		}
	}
	cfg.ConfigurationPath = v.ConfigFileUsed()
	cfg.viper = v

	if cfg.CustomchecksFilePath != "" {
		if utils.FileExists(cfg.CustomchecksFilePath) {
			if ccc, err := unmarshalCustomChecks(cfg.CustomchecksFilePath); err != nil {
				log.Errorln("Configuration: could not load custom checks: ", err)
			} else {
				cfg.CustomCheckConfiguration = ccc
			}
		} else {
			log.Errorln("Configuration: custom check configuration does not exist: ", cfg.CustomchecksFilePath)
		}
	}

	// we have to set at least an empty array if we don't load any configuration
	if cfg.CustomCheckConfiguration == nil {
		cfg.CustomCheckConfiguration = []*CustomCheck{}
	}

	return cfg, nil
}

// Load configuration from default paths or configPath. The reload func must be short lived or start a go routine.
func Load(ctx context.Context, configHint *LoadConfigHint) (*Configuration, error) {
	platformpath := platformpaths.Get()
	v := viper.New()
	setConfigurationDefaults(v)
	if configHint != nil {
		if configHint.ConfigFile != "" {
			v.SetConfigFile(configHint.ConfigFile)
		} else {
			v.SetConfigFile(path.Join(configHint.SearchPath, "config.cnf"))
		}
	} else {
		v.SetConfigFile(path.Join(platformpath.ConfigPath(), "config.cnf"))
	}
	v.SetConfigType("ini")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	return unmarshalConfiguration(v)
}

func unmarshalCustomChecks(configPath string) ([]*CustomCheck, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("ini")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := map[string]*CustomCheck{}

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	checks := make([]*CustomCheck, 0)
	for name, check := range cfg {
		if name != "default" {
			check.Name = name
			if check.Interval <= 0 {
				check.Interval = 60
			}
			if check.Timeout <= 0 {
				check.Timeout = 15
			}
			if strings.TrimSpace(check.Command) == "" {
				return nil, fmt.Errorf("missing command in custom check: %s", check.Name)
			}
			checks = append(checks, check)
		}
	}

	return checks, nil
}

func (c *Configuration) SaveConfiguration(config []byte) error {
	if err := ioutil.WriteFile(c.ConfigurationPath, config, 0600); err != nil {
		return err
	}
	return nil
}

func (c *Configuration) ReadConfigurationFile() ([]byte, error) {
	return ioutil.ReadFile(c.ConfigurationPath)
}

func (c *Configuration) SaveCustomCheckConfiguration(config []byte) error {
	if err := ioutil.WriteFile(c.CustomchecksFilePath, config, 0600); err != nil {
		return err
	}
	return nil
}

func (c *Configuration) ReadCustomCheckConfiguration() []byte {
	data, err := ioutil.ReadFile(c.CustomchecksFilePath)
	if err != nil {
		data = []byte{}
	}
	return data
}
