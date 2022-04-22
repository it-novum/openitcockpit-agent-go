module github.com/it-novum/openitcockpit-agent-go

go 1.15

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/andybalholm/crlf v0.0.0-20171020200849-670099aa064f
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/containerd/containerd v1.6.2 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2
	github.com/distatus/battery v0.10.0
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.14+incompatible
	github.com/elastic/beats/v7 v7.0.0-alpha2.0.20210222102351-e315d66b518a
	github.com/fsnotify/fsnotify v1.5.3 // indirect
	github.com/go-kit/kit v0.10.0 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/google/uuid v1.3.0
	github.com/gopherjs/gopherjs v0.0.0-20210202160940-bed99a852dfe // indirect
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/leoluk/perflib_exporter v0.1.0
	github.com/lufia/plan9stats v0.0.0-20220326011226-f1430873d8db // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/term v0.0.0-20210610120745-9d4ed1856297 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pkg/errors v0.9.1
	github.com/power-devops/perfstat v0.0.0-20220216144756-c35f1ee13d7c // indirect
	github.com/prometheus/common v0.34.0 // indirect
	github.com/prometheus/procfs v0.7.3
	github.com/shirou/gopsutil/v3 v3.22.3
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/spf13/cobra v1.4.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.11.0
	github.com/yusufpapurcu/wmi v1.2.2
	go.elastic.co/ecszap v1.0.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/net v0.0.0-20220421235706-1d1ef9303861 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20220422013727-9388b58f7150
	golang.org/x/text v0.3.7
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/genproto v0.0.0-20220421151946-72621c1f0bd3 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	howett.net/plist v1.0.0 // indirect
	libvirt.org/libvirt-go v7.4.0+incompatible
)

replace github.com/shirou/gopsutil/v3 v3.20.12 => github.com/it-novum/gopsutil/v3 v3.21.2-0.20210201093253-6e7f4ffe9947
