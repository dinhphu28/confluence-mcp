package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"dinhphu28/confluence-mcp/internal/config"
	"dinhphu28/confluence-mcp/internal/confluence"
	"dinhphu28/confluence-mcp/internal/jira"
	"dinhphu28/confluence-mcp/internal/mcpserver"
	"dinhphu28/confluence-mcp/internal/setup"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "setup":
			runCommand(setup.Run())

		case "confluence", "jira":
			runCommand(productCommand(os.Args[1], os.Args[2:]))

		case "--version":
			fmt.Println(version)
			return
		}
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	client := confluence.NewClient(cfg.Confluence.URL, cfg.Confluence.PAT)

	s := mcpserver.New(version)
	mcpserver.RegisterConfluence(s, client, cfg.Confluence.ReadOnly)

	if cfg.JiraEnabled() {
		jiraClient := jira.NewClient(cfg.Jira.URL, cfg.Jira.PAT)
		mcpserver.RegisterJira(s, jiraClient, cfg.Jira.ReadOnly)
	}

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
	}
}

// runCommand reports a CLI subcommand's outcome and exits — subcommands never
// fall through to starting the MCP server.
func runCommand(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

// productCommand dispatches `<product> setup|login`.
func productCommand(product string, args []string) error {
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}

	switch sub {
	case "setup":
		return setup.Product(product, setup.ModeSetup)
	case "login":
		return setup.Product(product, setup.ModeLogin)
	default:
		return fmt.Errorf("usage: confluence-mcp %s [setup|login]", product)
	}
}
