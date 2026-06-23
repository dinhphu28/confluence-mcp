package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"dinhphu28/confluence-mcp/internal/jira"
)

func registerJiraReadTools(s *server.MCPServer, client *jira.Client) {
	searchTool := mcp.NewTool(
		"jira_search",
		mcp.WithDescription("Search Jira issues with a JQL query"),
		mcp.WithString("jql", mcp.Required(), mcp.Description("JQL expression, e.g. 'project = DEV AND status = Open ORDER BY created DESC'")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of issues (default 25)")),
	)

	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		jql, err := request.RequireString("jql")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return jsonResult(client.Search(jql, request.GetInt("limit", 25)))
	})

	getIssueTool := mcp.NewTool(
		"jira_get_issue",
		mcp.WithDescription("Get a Jira issue by key or ID"),
		mcp.WithString("issue_key", mcp.Required(), mcp.Description("Issue key (e.g. DEV-123) or numeric ID")),
	)

	s.AddTool(getIssueTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		key, err := request.RequireString("issue_key")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return jsonResult(client.GetIssue(key))
	})

	getCommentsTool := mcp.NewTool(
		"jira_get_comments",
		mcp.WithDescription("Get the comments on a Jira issue"),
		mcp.WithString("issue_key", mcp.Required(), mcp.Description("Issue key (e.g. DEV-123) or numeric ID")),
	)

	s.AddTool(getCommentsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		key, err := request.RequireString("issue_key")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return jsonResult(client.GetComments(key))
	})
}
