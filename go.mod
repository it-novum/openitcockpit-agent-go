module github.com/it-novum/openitcockpit-agent-go

go 1.15

require (
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/andybalholm/crlf v0.0.0-20171020200849-670099aa064f
	github.com/coreos/go-systemd/v22 v22.5.0
	github.com/distatus/battery v0.11.0
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker v24.0.6+incompatible
	github.com/docker/go-units v0.5.0 // indirect
	github.com/elastic/beats/v7 v7.0.0-alpha2.0.20210222102351-e315d66b518a
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/google/uuid v1.3.1
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/leoluk/perflib_exporter v0.2.1
	github.com/lufia/plan9stats v0.0.0-20230326075908-cb1d2100619a // indirect
	github.com/moby/term v0.0.0-20220808134915-39b0c02b01ae // indirect
	github.com/opencontainers/image-spec v1.1.0-rc5 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/power-devops/perfstat v0.0.0-20221212215047-62379fc7944b // indirect
	github.com/prometheus/procfs v0.11.1
	github.com/shirou/gopsutil/v3 v3.23.8
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.7.0
	github.com/spf13/viper v1.16.0
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.3
	go.elastic.co/ecszap v1.0.0 // indirect
	go.uber.org/goleak v1.2.1 // indirect
	golang.org/x/sys v0.12.0
	golang.org/x/text v0.13.0
	golang.org/x/tools v0.13.0 // indirect
	gotest.tools/v3 v3.4.0 // indirect
	libvirt.org/libvirt-go v7.4.0+incompatible
)

replace github.com/shirou/gopsutil/v3 v3.20.12 => github.com/it-novum/gopsutil/v3 v3.21.2-0.20210201093253-6e7f4ffe9947
