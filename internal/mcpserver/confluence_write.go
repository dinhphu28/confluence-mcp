package mcpserver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"dinhphu28/atlassian-mcp/internal/confluence"
)

func registerConfluenceWriteTools(s *server.MCPServer, client *confluence.Client) {
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

	uploadAttachmentTool := mcp.NewTool(
		"confluence_upload_attachment",
		mcp.WithDescription("Upload a local file as an attachment on a Confluence page"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Page ID to attach the file to")),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Absolute path to the local file to upload")),
	)

	s.AddTool(uploadAttachmentTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pageID, err := request.RequireString("page_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		filePath, err := request.RequireString("file_path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return jsonResult(client.UploadAttachment(pageID, filepath.Base(filePath), data))
	})

	addLabelTool := mcp.NewTool(
		"confluence_add_label",
		mcp.WithDescription("Add a label to a Confluence page"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Confluence page ID")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Label name")),
	)

	s.AddTool(addLabelTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pageID, err := request.RequireString("page_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return jsonResult(client.AddLabel(pageID, name))
	})

	movePageTool := mcp.NewTool(
		"confluence_move_page",
		mcp.WithDescription("Move a Confluence page under a new parent (title and content preserved)"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("ID of the page to move")),
		mcp.WithString("target_parent_id", mcp.Required(), mcp.Description("ID of the new parent page")),
	)

	s.AddTool(movePageTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pageID, err := request.RequireString("page_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		targetParentID, err := request.RequireString("target_parent_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return jsonResult(client.MovePage(pageID, targetParentID))
	})

	replyToCommentTool := mcp.NewTool(
		"confluence_reply_to_comment",
		mcp.WithDescription("Reply to an existing Confluence comment"),
		mcp.WithString("parent_comment_id", mcp.Required(), mcp.Description("ID of the comment to reply to")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Reply body in the given representation")),
		mcp.WithString("representation", mcp.Description("Body format: 'storage' (default) or 'wiki'")),
	)

	s.AddTool(replyToCommentTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		parentCommentID, err := request.RequireString("parent_comment_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		content, err := request.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return jsonResult(client.ReplyToComment(
			parentCommentID, content,
			request.GetString("representation", "storage"),
		))
	})
}
