package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	baseURL = os.Getenv("CONFLUENCE_URL")
	token   = os.Getenv("CONFLUENCE_PAT")
)

func confluenceGET(path string) (string, error) {
	req, err := http.NewRequest("GET", baseURL+path, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("confluence error %d: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}

func main() {
	if baseURL == "" || token == "" {
		panic("CONFLUENCE_URL and CONFLUENCE_PAT are required")
	}

	s := server.NewMCPServer(
		"confluence-mcp",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	searchTool := mcp.NewTool(
		"confluence_search",
		mcp.WithDescription("Search Confluence pages using CQL text search"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search keyword")),
	)

	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		cql := fmt.Sprintf(`text ~ "%s"`, query)
		path := "/rest/api/content/search?limit=10&expand=space,version&cql=" + url.QueryEscape(cql)

		result, err := confluenceGET(path)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var pretty any
		_ = json.Unmarshal([]byte(result), &pretty)
		b, _ := json.MarshalIndent(pretty, "", "  ")

		return mcp.NewToolResultText(string(b)), nil
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

		path := "/rest/api/content/" + url.PathEscape(pageID) +
			"?expand=space,version,body.storage"

		result, err := confluenceGET(path)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var pretty any
		_ = json.Unmarshal([]byte(result), &pretty)
		b, _ := json.MarshalIndent(pretty, "", "  ")

		return mcp.NewToolResultText(string(b)), nil
	})

	if err := server.ServeStdio(s); err != nil {
		panic(err)
	}
}
