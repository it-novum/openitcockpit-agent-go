# openITCOCKPIT Monitoring Agent 3.0
Cross-Platform Monitoring Agent for openITCOCKPIT written in Go

## Table of contents
* [Supported operating systems](#supported-operating-systems)
* [Installation](#installation)
  - [Debian and Ubuntu](#debian-and-ubuntu)
  - [Red Hat Linux / CentOS / openSUSE](#red-hat-linux--centos--opensuse)
  - [Arch Linux](#arch-linux)
  - [Windows](#windows)
  - [macOS](#macos)
* [Supported Platforms](#supported-platforms)
* [Full documentation](#full-documentation)
* [License](#license)

## Supported operating systems

* Microsoft Windows Server 2012
* Microsoft Windows 8 or newer
* Apple macOS 10.14 Mojave or newer (Intel / Apple Silicon)
* Linux (Everything from Debian 6.0 (Squeeze) / CentOS 6.6 and newer should work fine)

Please notice: Due to old versions of PowerShell on Windows 7 / Windows Server 2008 R2 you need to add add the required Firewall rules manually to Windows Firewall.
Windows 7 / Windows Server 2008 R2 is official not supported by the Agent - even if it probably works.

## Requirements
* openITCOCKPIT Version >= 4.2

## Installation

Please visit the [release page](https://github.com/it-novum/openitcockpit-agent-go/releases) to download the latest or older versions.

### Debian and Ubuntu

#### Using the repository

```
curl https://packages.openitcockpit.io/repokey.txt | sudo apt-key add

sudo echo "deb https://packages.openitcockpit.io/openitcockpit-agent/deb/stable deb main" > /etc/apt/sources.list.d/openitcockpit-agent.list
sudo apt-get update

sudo apt-get install openitcockpit-agent
```

#### Manually
Install
```
sudo apt-get install ./openitcockpit-agent_3.x.x_amd64.deb
```

Uninstall
```
sudo apt-get purge openitcockpit-agent
```

### Red Hat Linux / CentOS / openSUSE

#### Using the repository

```
cat <<EOT > /etc/yum.repos.d/openitcockpit-agent.repo
[openitcockpit-agent]
name=openITCOCKPIT Monitoring Agent
baseurl=https://packages.openitcockpit.io/openitcockpit-agent/rpm/stable/$basearch/
enabled=1
gpgcheck=1
gpgkey=https://packages.openitcockpit.io/repokey.txt
EOT

yum-config-manager --enable openitcockpit-agent

yum install openitcockpit-agent
```

#### Manually
Install
```
rpm -i openitcockpit-agent-3.x.x-x.x86_64.rpm
```

Uninstall
```
rpm -e openitcockpit-agent
```

### Arch Linux
Install
```
sudo pacman -U openitcockpit-agent-3.x.x-x-x86_64.pkg.tar.zst
```

Uninstall
```
sudo pacman -R openitcockpit-agent
```

### Windows
Install

**GUI**

Install with double clicking the msi installer file.

![openITCOCKPIT Monitoring Agent MSI installer](/docs/images/msi_installer_new.png)

**CLI**

Automated install

```
msiexec.exe /i openitcockpit-agent*.msi INSTALLDIR="C:\Program Files\it-novum\openitcockpit-agent\" /qn
```

Uninstall

Please use the Windows built-in graphical software manager to uninstall.

### macOS

**GUI**

Install with double clicking the pkg installer file.

![openITCOCKPIT Monitoring Agent PKG installer](/docs/images/pkg_install_macos3.png)

**CLI**

Install
```
sudo installer -pkg openitcockpit-agent-3.x.x-darwin-amd64.pkg -target / -verbose
```

Uninstall
```
sudo installer -pkg openitcockpit-agent-uninstaller-3.x.x-darwin-amd64.pkg -target / -verbose
```

## Supported Platforms

| Platform              | Windows                | Linux | macOS |
|-----------------------|------------------------|-------|-------|
| 64 bit (amd64)        | ✅                      | ✅     | ✅     |
| 32 bit (i386)         | ✅                      | ✅     | -     |
| arm64 / Apple Silicon | Use the 32 bit version | ✅     | ✅     |


Please see to Wiki how to [cross compile binaries](https://github.com/it-novum/openitcockpit-agent-go/wiki/Build-binary#cross-compile) for different operating systems and CPU architectures.

## Full documentation
Do you want to build own binaries, learn more about cross compiling or how to start hacking the Agent?

Please see the [full documentation](https://github.com/it-novum/openitcockpit-agent-go/wiki).

## License
```
Copyright 2021 it-novum GmbH

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
