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

var cfg *Config

func confluenceGET(path string) (string, error) {
	req, err := http.NewRequest("GET", cfg.Confluence.URL+path, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+cfg.Confluence.PAT)
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
	var err error

	cfg, err = LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
	}
}
