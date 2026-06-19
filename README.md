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

Extract the zip, then open **PowerShell as Administrator** and run setup:

```powershell
Expand-Archive confluence-mcp_0.1.0_windows_amd64.zip
cd confluence-mcp_0.1.0_windows_amd64
.\confluence-mcp.exe setup
```

The binary is installed to `%USERPROFILE%\.local\bin\confluence-mcp.exe` and
the config is written to `%USERPROFILE%\.config\confluence-mcp\config.yaml`.

## Upgrading

To upgrade, extract the new release and run `setup` again:

```sh
./confluence-mcp setup
```

`setup` is safe to re-run. When a config already exists it **reuses your URL and
token** — it does not re-ask — and only:

- reinstalls the binary, and
- migrates the config to the current schema if needed.

You'll see one of `Reused existing config`, `Migrated existing config to
version N`, or (first run) `Created new config`.

To deliberately change the URL or token, force the prompts:

```sh
./confluence-mcp setup --reconfigure
```

At the prompts, pressing Enter keeps the current value (the existing token is
preserved on a blank line).

### Config versioning

The config file carries a `config_version` field so upgrades can migrate it
automatically instead of asking you to reconfigure:

```yaml
config_version: 1
confluence:
  url: https://your-confluence-server.com
  pat: <your-token>
```

- Configs written by older releases (without `config_version`) are migrated on
  the next `setup`.
- The server refuses to start against a config whose `config_version` is *newer*
  than the binary supports, telling you to upgrade — so a downgrade can't
  silently misread a newer file.

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
