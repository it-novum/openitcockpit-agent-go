module github.com/it-novum/openitcockpit-agent-go

go 1.15

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.5.1 // indirect
	github.com/alecthomas/units v0.0.0-20201120081800-1786d5ef83d4 // indirect
	github.com/andybalholm/crlf v0.0.0-20171020200849-670099aa064f
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/containerd/containerd v1.5.11 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2
	github.com/distatus/battery v0.10.0
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.14+incompatible
	github.com/elastic/beats/v7 v7.0.0-alpha2.0.20210222102351-e315d66b518a
	github.com/godbus/dbus/v5 v5.0.6 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/google/uuid v1.3.0
	github.com/gopherjs/gopherjs v0.0.0-20210202160940-bed99a852dfe // indirect
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/leoluk/perflib_exporter v0.1.0
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/moby/term v0.0.0-20210610120745-9d4ed1856297 // indirect
	github.com/pelletier/go-toml v1.9.3 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0 // indirect
	github.com/prometheus/procfs v0.7.3
	github.com/shirou/gopsutil/v3 v3.21.1
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/spf13/afero v1.5.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	github.com/yusufpapurcu/wmi v1.2.2
	go.elastic.co/ecszap v1.0.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20220207234003-57398862261d
	golang.org/x/text v0.3.7
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/genproto v0.0.0-20220208230804-65c12eb4c068 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	howett.net/plist v0.0.0-20201203080718-1454fab16a06 // indirect
	libvirt.org/libvirt-go v7.0.0+incompatible
)

replace github.com/shirou/gopsutil/v3 v3.20.12 => github.com/it-novum/gopsutil/v3 v3.21.2-0.20210201093253-6e7f4ffe9947
