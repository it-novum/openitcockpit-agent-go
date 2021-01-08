package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/it-novum/openitcockpit-agent-go/platformpaths"
	"github.com/spf13/viper"
)

// AgentVersion as the name says
const AgentVersion = "3.0.0"

// CustomCheck are external plugins and scripts which should be executed by the Agent
type CustomCheck struct {
	Name     string
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

// CustomCheckConfiguration stores only custom check commands
type CustomCheckConfiguration struct {
	ConfigurationPath string
	viper             *viper.Viper

	Checks []*CustomCheck
}

// Configuration with all sub configuration structs
type Configuration struct {
	ConfigurationPath string
	viper             *viper.Viper
	reload            ReloadFunc

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
	LaunchdServices bool  `mapstructure:"launchdservices"`
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

// Reload the agent with same configuration async
func (c *Configuration) Reload() {
	go func() {
		done := make(chan struct{})
		go func() {
			c.reload(c, nil)
			done <- struct{}{}
		}()
		t := time.NewTimer(time.Second * 30)
		defer t.Stop()
		select {
		case <-done:
			return
		case <-t.C:
			log.Fatalln("Internal error: timeout for configuration reload")
		}
	}()
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

// ReloadFunc will be called after the configuration has been loaded or changed
type ReloadFunc func(*Configuration, error)

// ReloadCustomCheckFunc will be called after the custom check configuration has been loaded or changed
type ReloadCustomCheckFunc func(*CustomCheckConfiguration, error)

// LoadConfigHint for Load func
type LoadConfigHint struct {
	SearchPath string
	ConfigFile string
}

func unmarshalConfiguration(v *viper.Viper) (*Configuration, error) {
	cfg := &Configuration{}
	cfg.Default = cfg
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	cfg.ConfigurationPath = v.ConfigFileUsed()
	cfg.viper = v
	return cfg, nil
}

// Load configuration from default paths or configPath. The reload func must be short lived or start a go routine.
func Load(reload ReloadFunc, configHint *LoadConfigHint) {
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
		reload(nil, err)
		return
	}

	v.OnConfigChange(func(in fsnotify.Event) {
		cfg, err := unmarshalConfiguration(v)
		cfg.reload = reload
		reload(cfg, err)
	})

	cfg, err := unmarshalConfiguration(v)
	cfg.reload = reload
	reload(cfg, err)
	if err == nil {
		go v.WatchConfig()
	}
}

func unmarshalCustomChecks(v *viper.Viper) (*CustomCheckConfiguration, error) {
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

	return &CustomCheckConfiguration{
		ConfigurationPath: v.ConfigFileUsed(),
		viper:             v,
		Checks:            checks,
	}, nil
}

// LoadCustomChecks from specified config file. The reload func must be short lived or start a go routine.
func LoadCustomChecks(configPath string, reload ReloadCustomCheckFunc) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("ini")

	if err := v.ReadInConfig(); err != nil {
		reload(nil, err)
		return
	}

	v.OnConfigChange(func(in fsnotify.Event) {
		ccConfig, err := unmarshalCustomChecks(v)
		reload(ccConfig, err)
	})

	ccConfig, err := unmarshalCustomChecks(v)
	reload(ccConfig, err)
	if err == nil {
		go v.WatchConfig()
	}
}

// SaveConfiguration to disk
func SaveConfiguration(oldConfiguration *Configuration, config []byte) error {
	if err := ioutil.WriteFile(oldConfiguration.ConfigurationPath, config, 0666); err != nil {
		return err
	}
	// reload will be done automatically by WatchConfig

	return nil
}
