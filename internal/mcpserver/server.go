// Package mcpserver wires the Confluence client to the MCP tool surface.
package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"dinhphu28/confluence-mcp/internal/confluence"
)

// New builds an MCP server exposing the Confluence tools backed by client. When
// readOnly is true, only the read tools are registered.
func New(client *confluence.Client, version string, readOnly bool) *server.MCPServer {
	s := server.NewMCPServer(
		"confluence-mcp",
		version,
		server.WithToolCapabilities(false),
	)

	registerReadTools(s, client)
	if !readOnly {
		registerWriteTools(s, client)
	}
	return s
}

func registerReadTools(s *server.MCPServer, client *confluence.Client) {
	searchTool := mcp.NewTool(
		"confluence_search",
		mcp.WithDescription("Search Confluence content. Provide either a free-text 'query' "+
			"(matched against page text) or a raw 'cql' expression for full control."),
		mcp.WithString("query", mcp.Description("Search keyword (matched as CQL text ~ \"query\")")),
		mcp.WithString("cql", mcp.Description("Raw CQL expression, e.g. 'space = DEV AND label = api'. Overrides 'query' when set.")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results (default 10)")),
	)

	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query := request.GetString("query", "")
		cql := request.GetString("cql", "")
		limit := request.GetInt("limit", 10)

		if cql == "" {
			if query == "" {
				return mcp.NewToolResultError("either 'query' or 'cql' is required"), nil
			}
			cql = fmt.Sprintf(`text ~ "%s"`, query)
		}

		return jsonResult(client.Search(cql, limit))
	})

	getPageTool := mcp.NewTool(
		"confluence_get_page",
		mcp.WithDescription("Get a Confluence page by ID"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Confluence page ID")),
	)

	s.AddTool(getPageTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pageID, err := request.RequireString("page_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return jsonResult(client.GetPage(pageID))
	})

	getChildrenTool := mcp.NewTool(
		"confluence_get_page_children",
		mcp.WithDescription("List the child pages directly under a Confluence page"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Parent Confluence page ID")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of children (default 25)")),
	)

	s.AddTool(getChildrenTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pageID, err := request.RequireString("page_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return jsonResult(client.GetPageChildren(pageID, request.GetInt("limit", 25)))
	})

	getCommentsTool := mcp.NewTool(
		"confluence_get_comments",
		mcp.WithDescription("Get the comments on a Confluence page"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Confluence page ID")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of comments (default 25)")),
	)

	s.AddTool(getCommentsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pageID, err := request.RequireString("page_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return jsonResult(client.GetComments(pageID, request.GetInt("limit", 25)))
	})
}

func registerWriteTools(s *server.MCPServer, client *confluence.Client) {
	createPageTool := mcp.NewTool(
		"confluence_create_page",
		mcp.WithDescription("Create a new Confluence page"),
		mcp.WithString("space_key", mcp.Required(), mcp.Description("Key of the space to create the page in")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Page title")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Page body in the given representation (storage-format XHTML by default)")),
		mcp.WithString("parent_id", mcp.Description("Optional parent page ID to nest under")),
		mcp.WithString("representation", mcp.Description("Body format: 'storage' (default) or 'wiki'")),
	)

	s.AddTool(createPageTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		spaceKey, err := request.RequireString("space_key")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		title, err := request.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		content, err := request.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return jsonResult(client.CreatePage(
			spaceKey, title, content,
			request.GetString("parent_id", ""),
			request.GetString("representation", "storage"),
		))
	})

	updatePageTool := mcp.NewTool(
		"confluence_update_page",
		mcp.WithDescription("Update an existing Confluence page (version is bumped automatically)"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Confluence page ID")),
		mcp.WithString("content", mcp.Required(), mcp.Description("New page body in the given representation")),
		mcp.WithString("title", mcp.Description("New title (keeps the existing title if omitted)")),
		mcp.WithString("representation", mcp.Description("Body format: 'storage' (default) or 'wiki'")),
	)

	s.AddTool(updatePageTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pageID, err := request.RequireString("page_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		content, err := request.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return jsonResult(client.UpdatePage(
			pageID, content,
			request.GetString("title", ""),
			request.GetString("representation", "storage"),
		))
	})

	addCommentTool := mcp.NewTool(
		"confluence_add_comment",
		mcp.WithDescription("Add a comment to a Confluence page"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Confluence page ID to comment on")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Comment body in the given representation")),
		mcp.WithString("representation", mcp.Description("Body format: 'storage' (default) or 'wiki'")),
	)

	s.AddTool(addCommentTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pageID, err := request.RequireString("page_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		content, err := request.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return jsonResult(client.AddComment(
			pageID, content,
			request.GetString("representation", "storage"),
		))
	})

	deletePageTool := mcp.NewTool(
		"confluence_delete_page",
		mcp.WithDescription("Delete a Confluence page by ID (moves it to the trash)"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Confluence page ID to delete")),
	)

	s.AddTool(deletePageTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pageID, err := request.RequireString("page_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if err := client.DeletePage(pageID); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Deleted page %s", pageID)), nil
	})
}

// jsonResult turns a raw JSON response (or error) into a pretty-printed MCP
// tool result.
func jsonResult(raw string, err error) (*mcp.CallToolResult, error) {
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var pretty any
	_ = json.Unmarshal([]byte(raw), &pretty)
	b, _ := json.MarshalIndent(pretty, "", "  ")

	return mcp.NewToolResultText(string(b)), nil
}
