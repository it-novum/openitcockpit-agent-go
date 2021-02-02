module github.com/it-novum/openitcockpit-agent-go

go 1.15

require (
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d
	github.com/alecthomas/units v0.0.0-20201120081800-1786d5ef83d4 // indirect
	github.com/containerd/containerd v1.4.3 // indirect
	github.com/coreos/go-systemd/v22 v22.1.0
	github.com/distatus/battery v0.10.0
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v20.10.2+incompatible
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/google/uuid v1.2.0
	github.com/gorilla/mux v1.8.0
	github.com/leoluk/perflib_exporter v0.1.0
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/prometheus/client_golang v1.9.0 // indirect
	github.com/prometheus/common v0.15.0
	github.com/prometheus/procfs v0.3.0
	github.com/shirou/gopsutil/v3 v3.20.12
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/afero v1.5.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777 // indirect
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a // indirect
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c
	golang.org/x/text v0.3.5
	google.golang.org/genproto v0.0.0-20210126160654-44e461bb6506 // indirect
	google.golang.org/grpc v1.35.0 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gotest.tools/v3 v3.0.3 // indirect
	howett.net/plist v0.0.0-20201203080718-1454fab16a06 // indirect
	libvirt.org/libvirt-go v7.0.0+incompatible
)

replace github.com/shirou/gopsutil/v3 v3.20.12 => github.com/it-novum/gopsutil/v3 v3.21.2-0.20210201093253-6e7f4ffe9947
