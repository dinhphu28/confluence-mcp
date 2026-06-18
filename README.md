# Confluence MCP

A small, zero-dependency MCP server (single Go binary) exposing read-only
Confluence Server/Data Center tools over stdio.

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

## Tools

All tools are read-only and authenticate with the configured Personal Access
Token.

### `confluence_search`

Search Confluence content.

| Param   | Required | Description                                                              |
| ------- | -------- | ------------------------------------------------------------------------ |
| `query` | no\*     | Free-text keyword, matched as CQL `text ~ "query"`.                      |
| `cql`   | no\*     | Raw CQL expression, e.g. `space = DEV AND label = api`. Overrides `query`. |
| `limit` | no       | Maximum number of results (default `10`).                                |

\* Provide either `query` or `cql`.

### `confluence_get_page`

Get a single page by ID (includes `body.storage`, space, and version).

| Param     | Required | Description           |
| --------- | -------- | --------------------- |
| `page_id` | yes      | Confluence page ID.   |

### `confluence_get_page_children`

List the child pages directly under a page.

| Param     | Required | Description                              |
| --------- | -------- | ---------------------------------------- |
| `page_id` | yes      | Parent Confluence page ID.               |
| `limit`   | no       | Maximum number of children (default `25`). |

### `confluence_get_comments`

Get the comments on a page (includes `body.storage`).

| Param     | Required | Description                              |
| --------- | -------- | ---------------------------------------- |
| `page_id` | yes      | Confluence page ID.                      |
| `limit`   | no       | Maximum number of comments (default `25`). |

## Confluence access token

Get a Personal Access Token from:

`https://your-confluence-server.com/plugins/personalaccesstokens/usertokens.action`
