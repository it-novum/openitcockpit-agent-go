# openITCOCKPIT Monitoring Agent 3.0
Cross-Platform Monitoring Agent for openITCOCKPIT written in Go

1. [Installation](#Installation)
2. [Usage](#Usage)
3. [Join Development](#Join-Development)

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

Start hacking :)

## Notes
- https://github.com/kata-containers/govmm