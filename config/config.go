package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/it-novum/openitcockpit-agent-go/basiclog"
	"github.com/it-novum/openitcockpit-agent-go/platformpaths"
	"github.com/it-novum/openitcockpit-agent-go/utils"
	"github.com/spf13/viper"
)

// AgentVersion as the name says
const AgentVersion = "3.0.12"

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
	Shell         string `mapstructure:"shell"`
	PowershellExe string `mapstructure:"powershell_exe"`
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

type PrometheusConfiguration struct {
	Enable            bool   `mapstructure:"enabled"`
	ExportersFilePath string `mapstructure:"exporters"`
}

type PrometheusExporter struct {
	Name     string `mapstructure:"-"`
	Enabled  bool   `mapstructure:"enabled"`
	Method   string `mapstructure:"method"` //http or https
	Port     int64  `mapstructure:"port"`   // 9100
	Path     string `mapstructure:"path"`   // /metrics
	Interval int64  `mapstructure:"interval"`
	Timeout  int64  `mapstructure:"timeout"`
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

	// EnablePPROF for debugging memory leaks with the go tool pprof command
	EnablePPROF bool `mapstructure:"enable-dev-pprof"`

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
	Sensors         bool  `mapstructure:"sensorstats"`
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
	Ntp             bool  `mapstructure:"ntp"`

	// Alfresco

	JmxUser     string `mapstructure:"alfresco-jmxuser"`
	JmxPassword string `mapstructure:"alfresco-jmxpassword"`
	JmxAddress  string `mapstructure:"alfresco-jmxaddress"`
	JmxPort     int64  `mapstructure:"alfresco-jmxport"`
	JmxPath     string `mapstructure:"alfresco-jmxpath"`
	JmxQuery    string `mapstructure:"alfresco-jmxquery"`
	JavaPath    string `mapstructure:"alfresco-javapath"`

	// WindowsEventLog with all event log types to monitor

	WindowsEventLogTypes  []string `mapstructure:"wineventlog-logtypes"`
	WindowsEventLogAge    int64    `mapstructure:"wineventlog-age"`    // WMI Version
	WindowsEventLogCache  int64    `mapstructure:"wineventlog-cache"`  // JD Version
	WindowsEventLogMethod string   `mapstructure:"wineventlog-method"` // WMI or PowerShell

	// Push Mode

	OITC *PushConfiguration `json:"oitc"`

	// Default is the namespace workaround we need for the configuration file format
	Default *Configuration `json:"-"`

	CustomCheckConfiguration []*CustomCheck `json:"customchecks_configuration" mapstructure:"-"`

	// Prometheus Exporter / Proxy
	Prometheus            *PrometheusConfiguration `json:"prometheus"`
	PrometheusExporterConfiguration []*PrometheusExporter    `json:"prometheus_exporter_configuration" mapstructure:"-"`
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
	"ntp":                  true,
	"wineventlog-logtypes": "System,Application",
	"wineventlog-age":      3600,
	"wineventlog-cache":    3600,
	"wineventlog-method":   "WMI",
	"customchecks":         filepath.Join(platformpaths.Get().ConfigPath(), "customchecks.ini"),
	"autossl-folder":       platformpaths.Get().ConfigPath(),
	"autossl-csr-file":     filepath.Join(platformpaths.Get().ConfigPath(), "agent.csr"),
	"autossl-crt-file":     filepath.Join(platformpaths.Get().ConfigPath(), "agent.crt"),
	"autossl-key-file":     filepath.Join(platformpaths.Get().ConfigPath(), "agent.key"),
	"autossl-ca-file":      filepath.Join(platformpaths.Get().ConfigPath(), "server_ca.crt"),
}

var oitcDefaultvalue = map[string]interface{}{
	"authfile": filepath.Join(platformpaths.Get().ConfigPath(), "auth.json"),
}

var prometheusDefaultvalue = map[string]interface{}{
	"enabled":   false,
	"exporters": filepath.Join(platformpaths.Get().ConfigPath(), "prometheus_exporters.ini"),
}

func setConfigurationDefaults(v *viper.Viper) {
	for key, value := range defaultValue {
		v.SetDefault("default."+key, value)
	}

	for key, value := range oitcDefaultvalue {
		v.SetDefault("oitc."+key, value)
	}

	for key, value := range prometheusDefaultvalue {
		v.SetDefault("prometheus."+key, value)
	}
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
				logger, _ := basiclog.New()
				logger.Errorln("Configuration: could not load custom checks: ", err)
			} else {
				cfg.CustomCheckConfiguration = ccc
			}
		} else {
			logger, _ := basiclog.New()
			logger.Errorln("Configuration: custom check configuration does not exist: ", cfg.CustomchecksFilePath)
		}
	}

	// we have to set at least an empty array if we don't load any configuration
	if cfg.CustomCheckConfiguration == nil {
		cfg.CustomCheckConfiguration = []*CustomCheck{}
	}

	// Parse Prometheus Exporter configuration
	if cfg.Prometheus.ExportersFilePath != "" && cfg.Prometheus.Enable {
		if utils.FileExists(cfg.Prometheus.ExportersFilePath) {
			if promExporters, err := unmarshalPrometheusExporters(cfg.Prometheus.ExportersFilePath); err != nil {
				logger, _ := basiclog.New()
				logger.Errorln("Configuration: could not load prometheus exporter: ", err)
			} else {
				cfg.PrometheusExporterConfiguration = promExporters
			}
		} else {
			logger, _ := basiclog.New()
			logger.Errorln("Configuration: Prometheus exporter configuration does not exist: ", cfg.CustomchecksFilePath)
		}
	}

	// we have to set at least an empty array if we don't load any exporter configuration
	if cfg.PrometheusExporterConfiguration == nil {
		cfg.PrometheusExporterConfiguration = []*PrometheusExporter{}
	}

	return cfg, nil
}

// Load configuration from default paths or configPath. The reload func must be short lived or start a go routine.
func Load(ctx context.Context, configPath string) (*Configuration, error) {
	v := viper.New()
	setConfigurationDefaults(v)

	v.SetConfigFile(configPath)

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
			if check.Enabled {
				checks = append(checks, check)
			}
		}
	}

	return checks, nil
}

func unmarshalPrometheusExporters(configPath string) ([]*PrometheusExporter, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("ini")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := map[string]*PrometheusExporter{}

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	exporters := make([]*PrometheusExporter, 0)
	for name, check := range cfg {
		if name != "default" {
			check.Name = name
			if check.Timeout <= 0 {
				check.Timeout = 15
			}

			if strings.TrimSpace(check.Method) == "https" {
				check.Method = "https"
			} else {
				check.Method = "http"
			}

			if strings.TrimSpace(check.Path) == "" {
				return nil, fmt.Errorf("missing path for prometheus exporter: %s", check.Name)
			}
			if check.Enabled {
				exporters = append(exporters, check)
			}
		}
	}

	return exporters, nil
}

func (c *Configuration) SaveConfiguration(config []byte) error {
	if err := os.WriteFile(c.ConfigurationPath, config, 0600); err != nil {
		return err
	}
	return nil
}

func (c *Configuration) ReadConfigurationFile() ([]byte, error) {
	return os.ReadFile(c.ConfigurationPath)
}

func (c *Configuration) SaveCustomCheckConfiguration(config []byte) error {
	if err := os.WriteFile(c.CustomchecksFilePath, config, 0600); err != nil {
		return err
	}
	return nil
}

func (c *Configuration) ReadCustomCheckConfiguration() []byte {
	data, err := os.ReadFile(c.CustomchecksFilePath)
	if err != nil {
		data = []byte{}
	}
	return data
}

func (c *Configuration) ReadPrometheusExporterConfiguration() []byte {
	data, err := os.ReadFile(c.Prometheus.ExportersFilePath)
	if err != nil {
		data = []byte{}
	}
	return data
}

func (c *Configuration) SavePrometheusExporterConfiguration(config []byte) error {
	if err := os.WriteFile(c.Prometheus.ExportersFilePath, config, 0600); err != nil {
		return err
	}
	return nil
}