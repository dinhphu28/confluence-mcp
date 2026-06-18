# Confluence MCP

## Build

Linux:

```sh
make release VERSION=0.1.0            # -> dist/confluence-mcp_0.1.0_linux_amd64.tar.gz
```

Windows:

```sh
make release-windows VERSION=0.1.0    # -> dist/confluence-mcp_0.1.0_windows_amd64.zip
```

Both at once:

```sh
make release-all VERSION=0.1.0
```

Cross-compiling is pure Go (`CGO_ENABLED=0`), so the Windows binary can be
built from Linux/mac. `zip` is required for the Windows package.

## Installation

### Linux / macOS

Extract the tarball and run setup:

```sh
tar -xzf confluence-mcp_0.1.0_linux_amd64.tar.gz
cd confluence-mcp_0.1.0_linux_amd64
./confluence-mcp setup
```

The binary is installed to `~/.local/bin/confluence-mcp` and the config is
written to `~/.config/confluence-mcp/config.yaml`.

### Windows

Extract the zip and run setup from PowerShell or Command Prompt:

```powershell
Expand-Archive confluence-mcp_0.1.0_windows_amd64.zip
cd confluence-mcp_0.1.0_windows_amd64
.\confluence-mcp.exe setup
```

The binary is installed to `%USERPROFILE%\.local\bin\confluence-mcp.exe` and
the config is written to `%USERPROFILE%\.config\confluence-mcp\config.yaml`.

## opencode config

`setup` prints the exact snippet to add, using the installed path. For example:

```yaml
"confluence":
  {
    "type": "local",
    "command": ["/home/<YOUR_HOME>/.local/bin/confluence-mcp"],
  }
```

On Windows the command path will be the `.exe`, e.g.
`["C:\\Users\\<you>\\.local\\bin\\confluence-mcp.exe"]`.

## Confluence access token

Get a Personal Access Token from:

`https://your-confluence-server.com/plugins/personalaccesstokens/usertokens.action`
