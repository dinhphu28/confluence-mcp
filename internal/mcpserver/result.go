package mcpserver

import (
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

// jsonResult turns a raw JSON response (or error) into a pretty-printed MCP
// tool result. Shared by all product tool sets.
func jsonResult(raw string, err error) (*mcp.CallToolResult, error) {
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var pretty any
	_ = json.Unmarshal([]byte(raw), &pretty)
	b, _ := json.MarshalIndent(pretty, "", "  ")

	return mcp.NewToolResultText(string(b)), nil
}
