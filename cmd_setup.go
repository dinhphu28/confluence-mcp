package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func RunSetup() error {
	if err := installSelf(); err != nil {
		return fmt.Errorf("install binary: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== Confluence MCP Setup ===")
	fmt.Println()

	fmt.Print("Confluence URL [https://confluence.cads.live]: ")
	urlInput, _ := reader.ReadString('\n')
	confluenceURL := strings.TrimSpace(urlInput)

	if confluenceURL == "" {
		confluenceURL = "https://confluence.cads.live"
	}

	fmt.Print("Confluence Personal Access Token: ")
	patInput, _ := reader.ReadString('\n')
	pat := strings.TrimSpace(patInput)

	if pat == "" {
		return fmt.Errorf("personal access token is required")
	}

	cfg := Config{
		Confluence: ConfluenceConfig{
			URL: confluenceURL,
			PAT: pat,
		},
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".config", "confluence-mcp")
	configPath := filepath.Join(configDir, "config.yaml")

	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("Installed binary: %s\n",
		filepath.Join(homeDir, ".local", "bin", "confluence-mcp"))
	fmt.Printf("Config written: %s\n", configPath)

	fmt.Println()
	fmt.Println("Add this to your opencode config:")
	fmt.Println()

	fmt.Println(`"confluence": {`)
	fmt.Println(`  "type": "local",`)
	fmt.Printf(`  "command": ["%s"]`+"\n",
		filepath.Join(homeDir, ".local", "bin", "confluence-mcp"))
	fmt.Println(`}`)

	return nil
}

func installSelf() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	targetDir := filepath.Join(homeDir, ".local", "bin")

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}

	targetPath := filepath.Join(targetDir, "confluence-mcp")

	src, err := os.Open(exePath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(
		targetPath,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0o755,
	)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
