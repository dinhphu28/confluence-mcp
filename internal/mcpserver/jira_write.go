package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"dinhphu28/confluence-mcp/internal/jira"
)

func registerJiraWriteTools(s *server.MCPServer, client *jira.Client) {
	createIssueTool := mcp.NewTool(
		"jira_create_issue",
		mcp.WithDescription("Create a new Jira issue"),
		mcp.WithString("project_key", mcp.Required(), mcp.Description("Project key, e.g. DEV")),
		mcp.WithString("issue_type", mcp.Required(), mcp.Description("Issue type name, e.g. Task, Bug, Story")),
		mcp.WithString("summary", mcp.Required(), mcp.Description("Issue summary/title")),
		mcp.WithString("description", mcp.Description("Issue description (Jira wiki markup)")),
	)

	s.AddTool(createIssueTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectKey, err := request.RequireString("project_key")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		issueType, err := request.RequireString("issue_type")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		summary, err := request.RequireString("summary")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return jsonResult(client.CreateIssue(
			projectKey, issueType, summary,
			request.GetString("description", ""),
		))
	})

	addCommentTool := mcp.NewTool(
		"jira_add_comment",
		mcp.WithDescription("Add a comment to a Jira issue"),
		mcp.WithString("issue_key", mcp.Required(), mcp.Description("Issue key (e.g. DEV-123) or numeric ID")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Comment body (Jira wiki markup)")),
	)

	s.AddTool(addCommentTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		key, err := request.RequireString("issue_key")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		body, err := request.RequireString("body")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return jsonResult(client.AddComment(key, body))
	})
}
