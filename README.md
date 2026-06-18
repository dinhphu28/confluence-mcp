# Confluence MCP

## Build

```sh
make release VERSION=0.1.0
```

Output will be in `dist/confluence-mcp_0.1.0_linux_amd64.tar.gz`

## Installation

Extract the tarball:

```sh
tar -xzf confluence-mcp_0.1.0_linux_amd64.tar.gz
cd confluence-mcp_0.1.0_linux_amd64
```

Then run the setup command:

```sh
./confluence-mcp setup
```

To use with opencode, add the config of mcp to the opencode config file:

```yaml
"confluence":
  {
    "type": "local",
    "command": ["/home/<YOUR_HOME>/.local/bin/confluence-mcp"],
  }
```

To get Confluence access token. Go to

`https://your-confluence-server.com/plugins/personalaccesstokens/usertokens.action`

