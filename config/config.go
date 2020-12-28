package config

import (
	"fmt"
	"path"
	"strings"

	"github.com/it-novum/openitcockpit-agent-go/platformpaths"
	"github.com/spf13/viper"
)

// AgentVersion as the name says
const AgentVersion = "2.1.0"

// CustomCheck are external plugins and scripts which should be executed by the Agent
type CustomCheck struct {
	Name     string
	Interval int64  `mapstructure:"interval"`
	Enabled  bool   `mapstructure:"enabled"`
	Command  string `mapstructure:"command"`
	Timeout  int64  `mapstructure:"timeout"`
}

// Configuration with all sub configuration structs
type Configuration struct {

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

	Verbose            bool   `mapstructure:"verbose"`
	Debug              bool   `mapstructure:"debug"`
	ConfigUpdate       bool   `mapstructure:"config-update-mode"`
	CustomchecksConfig string `mapstructure:"customchecks"`

	// Default Checks

	CheckInterval   int64 `mapstructure:"interval"`
	Docker          bool  `mapstructure:"dockerstats"`
	Qemu            bool  `mapstructure:"qemustats"`
	CPU             bool  `mapstructure:"cpustats"`
	Processes       bool  `mapstructure:"processstats"`
	Netstats        bool  `mapstructure:"netstats"`
	NetIo           bool  `mapstructure:"netio"`
	Diskstats       bool  `mapstructure:"diskstats"`
	DiskIo          bool  `mapstructure:"diskio"`
	WindowsServices bool  `mapstructure:"winservices"`
	WindowsEventLog bool  `mapstructure:"wineventlog"`
	SystemdServices bool  `mapstructure:"systemdservices"`
	Alfresco        bool  `mapstructure:"alfrescostats"`

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

	OITC struct {
		Push         bool   `mapstructure:"enabled"`
		HostUUID     string `mapstructure:"hostuuid"`
		URL          string `mapstructure:"url"`
		Apikey       string `mapstructure:"apikey"`
		Proxy        string `mapstructure:"proxy"`
		PushInterval int64  `mapstructure:"interval"`
	} `mapstructure:"oitc"`

	// Default is the namespace workaround we need for the configuration file format
	Default *Configuration
}

var defaultValue = map[string]interface{}{
	"port":                 3333,
	"interval":             30,
	"qemustats":            true,
	"cpustats":             true,
	"processstats":         true,
	"netstats":             true,
	"netio":                true,
	"diskstats":            true,
	"diskio":               true,
	"winservices":          true,
	"wineventlog":          true,
	"systemdservices":      true,
	"alfrescostats":        true,
	"wineventlog-logtypes": "System,Application,Security",
	"customchecks":         path.Join(platformpaths.Get().ConfigPath(), "customchecks.cnf"),
	"certfile":             path.Join(platformpaths.Get().ConfigPath(), "agent.crt"),
	"keyfile":              path.Join(platformpaths.Get().ConfigPath(), "agent.key"),
	"autossl-folder":       platformpaths.Get().ConfigPath(),
	"autossl-csr-file":     path.Join(platformpaths.Get().ConfigPath(), "agent.csr"),
	"autossl-crt-file":     path.Join(platformpaths.Get().ConfigPath(), "agent.crt"),
	"autossl-key-file":     path.Join(platformpaths.Get().ConfigPath(), "agent.key"),
	"autossl-ca-file":      path.Join(platformpaths.Get().ConfigPath(), "server_ca.crt"),
}

var oitcDefaultvalue = map[string]interface{}{
	"interval": 60,
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

// Load configuration from default paths or configPath
func Load(configHint *LoadConfigHint) (*Configuration, error) {
	platformpath := platformpaths.Get()
	v := viper.New()
	setConfigurationDefaults(v)
	if configHint != nil {
		if configHint.ConfigFile != "" {
			v.SetConfigFile(configHint.ConfigFile)
		} else {
			v.AddConfigPath(configHint.SearchPath)
		}
	} else {
		v.AddConfigPath(platformpath.ConfigPath())
		v.AddConfigPath(".")
		v.SetConfigName("config.ini")
	}
	v.SetConfigType("ini")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	cfg := &Configuration{}
	cfg.Default = cfg
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadCustomChecks from specified config file
func LoadCustomChecks(configPath string) ([]*CustomCheck, error) {
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
