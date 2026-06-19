// Package setup implements the interactive `setup` command: it installs the
// binary, then creates or migrates the config file.
package setup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"dinhphu28/confluence-mcp/internal/config"
)

const defaultURL = "https://confluence.cads.live"

// binaryName returns the platform-appropriate executable name. On Windows the
// binary must carry the .exe extension to be runnable.
func binaryName() string {
	if runtime.GOOS == "windows" {
		return "confluence-mcp.exe"
	}
	return "confluence-mcp"
}

// installPath returns the full path the binary is installed to.
func installPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".local", "bin", binaryName()), nil
}

// Run installs the binary and ensures the config exists. On a fresh install it
// prompts for the URL and PAT; on re-run (e.g. after an upgrade) it reuses the
// existing config and only migrates the schema, without re-asking — unless
// reconfigure is true or a required field is missing.
func Run(reconfigure bool) error {
	if err := installSelf(); err != nil {
		return fmt.Errorf("install binary: %w", err)
	}

	cfg, found, err := config.Read()
	if err != nil {
		return fmt.Errorf("cannot read existing config: %w", err)
	}
	if cfg == nil {
		cfg = &config.Config{}
	}

	migrated := false
	if found {
		migrated = config.Migrate(cfg)
	}

	needPrompt := reconfigure || !found ||
		cfg.Confluence.URL == "" || cfg.Confluence.PAT == ""

	if needPrompt {
		fmt.Println("=== Confluence MCP Setup ===")
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)
		cfg.Confluence.URL = promptURL(reader, cfg.Confluence.URL)
		cfg.Confluence.PAT = promptPAT(reader, cfg.Confluence.PAT)
	}

	if cfg.Confluence.PAT == "" {
		return fmt.Errorf("personal access token is required")
	}

	cfg.Version = config.CurrentConfigVersion

	if err := config.Save(cfg); err != nil {
		return err
	}

	configPath, err := config.Path()
	if err != nil {
		return err
	}

	binPath, err := installPath()
	if err != nil {
		return err
	}

	fmt.Println()
	switch {
	case !found:
		fmt.Println("Created new config.")
	case needPrompt:
		fmt.Println("Updated config.")
	case migrated:
		fmt.Printf("Migrated existing config to version %d.\n", config.CurrentConfigVersion)
	default:
		fmt.Println("Reused existing config (binary upgraded, settings kept).")
	}

	fmt.Printf("Installed binary: %s\n", binPath)
	fmt.Printf("Config: %s\n", configPath)

	fmt.Println()
	fmt.Println("Add this to your opencode config:")
	fmt.Println()

	fmt.Println(`"confluence": {`)
	fmt.Println(`  "type": "local",`)
	fmt.Printf(`  "command": [%q]`+"\n", binPath)
	fmt.Println(`}`)

	return nil
}

// promptURL asks for the Confluence URL, defaulting to the current value (or the
// built-in default when there is none).
func promptURL(reader *bufio.Reader, current string) string {
	def := current
	if def == "" {
		def = defaultURL
	}

	fmt.Printf("Confluence URL [%s]: ", def)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return def
	}
	return input
}

// promptPAT asks for the Personal Access Token. When one already exists, an empty
// answer keeps it.
func promptPAT(reader *bufio.Reader, current string) string {
	label := "Confluence Personal Access Token"
	if current != "" {
		label += " [keep existing]"
	}

	fmt.Printf("%s: ", label)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return current
	}
	return input
}

func installSelf() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	targetPath, err := installPath()
	if err != nil {
		return err
	}

	// Already running from the install location; nothing to copy.
	if exePath == targetPath {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}

	src, err := os.Open(exePath)
	if err != nil {
		return err
	}
	defer src.Close()

	// Write to a temp file then atomically rename over the target. A plain
	// truncate-in-place fails with "text file busy" (ETXTBSY) when the existing
	// binary is currently running, e.g. during an upgrade; rename does not.
	tmpPath := targetPath + ".new"
	dst, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := dst.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}
