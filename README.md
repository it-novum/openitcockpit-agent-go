# openITCOCKPIT Monitoring Agent 3.0
Cross-Platform Monitoring Agent for openITCOCKPIT written in Go

1. [Installation](#Installation)
2. [Requirements](#Requirements)
3. [Usage](#Usage)
4. [Join Development](#Join-Development)

## Requirements

- Windows Server 2012 or newer
- Windows 8 or newer
- macOS 10.14 Mojave or newer
- Linux

## Join Development

Do you want to modify the source code of the openITCOCKPIT Monitoring Agent? If yes follow this guide to getting started. 

Please make sure you have install [Golang](https://golang.org/) >= 1.15.6 and [Visual Studio Code](https://code.visualstudio.com/) installed.

1. Clone this repository
```
git clone https://github.com/it-novum/openitcockpit-agent.git
```

2. Run Visual Studio Code and make sure that you have installed the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go)
![Install Go extension for VS Code](docs/images/vscode_golang_ext.jpg)

3. Install Go tools
Press `ctrl` + `shift` + `P` (Windows and Linux) or `cmd` + `shift` + `P` on macOS and select `Go: Install/Update Tools`
![Install Go tools](docs/images/vscode_install_go_tools.png)

Select all tools and confirm with `Ok`
![Select and install Go tools](docs/images/vscode_install_all_go_tools.png)

The installation is completed, as soon as you see `All tools successfully installed. You are ready to Go :).` in the VS Code terminal

4. Setup VS Code
Press `ctrl` + `shift` + `P` (Windows and Linux) or `cmd` + `shift` + `P` on macOS and type `settings json` and select `Preferences: Open Settings (JSON)`.
Add the following settings to your JSON.
```JS
    "go.testTimeout": "90s",
    "go.useLanguageServer": true,
    "go.lintTool": "golangci-lint",
    "[go]": {
        "editor.formatOnSave": true,
        "editor.codeActionsOnSave": {
            "source.organizeImports": true,
        },
        // Optional: Disable snippets, as they conflict with completion ranking.
        "editor.snippetSuggestions": "none",
    },
    "[go.mod]": {
        "editor.formatOnSave": true,
        "editor.codeActionsOnSave": {
            "source.organizeImports": true,
        },
    },
    "gopls": {
        // Add parameter placeholders when completing a function.
        "usePlaceholders": true,

        // If true, enable additional analyses with staticcheck.
        // Warning: This will significantly increase memory usage.
        "staticcheck": false,
    }
```
> Source: https://github.com/golang/tools/blob/master/gopls/doc/vscode.md

5. Debug Launch Configuration

Run -> Open Configurations
```JS
{
    // Use IntelliSense to learn about possible attributes.
    // Hover to view descriptions of existing attributes.
    // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}",
            "env": {},
            "args": ["-c", ".\\config.cnf", "--disable-logfile", "--debug"]
        }
    ]
}
```

Create a new file in workspace folder -> "config.ini"
```ini
[default]
customchecks = ./customchecks.ini
```

Create a new file in workspace folder -> "customchecks.ini" (Windows)

```ini
[check_Windows_Services_Status_OSS]
command = echo 'hello world'
interval = 15
timeout = 10
enabled = false
```

Create a new file in workspace folder -> "customchecks.ini" (Linux/Mac)

```ini
[check_echo]
command = echo 'hello world'
interval = 15
timeout = 10
enabled = false
```


Start hacking :)

## Notes
- https://github.com/kata-containers/govmm
- https://github.com/digitalocean/go-qemu
- https://github.com/cha87de/kvmtop
- https://github.com/0xef53/go-qmp
- https://github.com/0xef53/kvmrun
