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

var version = "dev"

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

// confluenceGETResult performs a GET and returns the response as a
// pretty-printed JSON tool result, or a tool error result.
func confluenceGETResult(path string) (*mcp.CallToolResult, error) {
	result, err := confluenceGET(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var pretty any
	_ = json.Unmarshal([]byte(result), &pretty)
	b, _ := json.MarshalIndent(pretty, "", "  ")

	return mcp.NewToolResultText(string(b)), nil
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "setup":
			if err := RunSetup(); err != nil {
				fmt.Fprintf(os.Stderr, "setup error: %v\n", err)
				os.Exit(1)
			}
			return

		case "--version":
			fmt.Println(version)
			return
		}
	}

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

		path := fmt.Sprintf("/rest/api/content/search?limit=%d&expand=space,version&cql=%s",
			limit, url.QueryEscape(cql))

		return confluenceGETResult(path)
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

		return confluenceGETResult(path)
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

		limit := request.GetInt("limit", 25)
		path := fmt.Sprintf("/rest/api/content/%s/child/page?limit=%d&expand=space,version",
			url.PathEscape(pageID), limit)

		return confluenceGETResult(path)
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

		limit := request.GetInt("limit", 25)
		path := fmt.Sprintf("/rest/api/content/%s/child/comment?limit=%d&expand=body.storage,version",
			url.PathEscape(pageID), limit)

		return confluenceGETResult(path)
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
	}
}
