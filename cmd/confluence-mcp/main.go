package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"dinhphu28/confluence-mcp/internal/config"
	"dinhphu28/confluence-mcp/internal/confluence"
	"dinhphu28/confluence-mcp/internal/mcpserver"
	"dinhphu28/confluence-mcp/internal/setup"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "setup":
			reconfigure := len(os.Args) > 2 && os.Args[2] == "--reconfigure"
			if err := setup.Run(reconfigure); err != nil {
				fmt.Fprintf(os.Stderr, "setup error: %v\n", err)
				os.Exit(1)
			}
			return

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
	s := mcpserver.New(client, version, cfg.Confluence.ReadOnly)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
	}
}
