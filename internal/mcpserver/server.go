// Package mcpserver builds the MCP server and registers the Atlassian tool sets
// (Confluence today; Jira can be added the same way without touching existing
// product files).
package mcpserver

import (
	"github.com/mark3labs/mcp-go/server"
)

// New builds the base MCP server. Tools are added afterwards by the per-product
// Register* functions (e.g. RegisterConfluence).
func New(version string) *server.MCPServer {
	return server.NewMCPServer(
		"atlassian-mcp",
		version,
		server.WithToolCapabilities(false),
	)
}
