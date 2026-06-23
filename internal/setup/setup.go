// Package setup implements the interactive setup/login commands: it installs the
// binary and creates, migrates or re-authenticates the config file, per product
// (Confluence and Jira).
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

const defaultConfluenceURL = "https://confluence.cads.live"

// Mode selects how Product configures a single product.
type Mode int

const (
	// ModeSetup (re)configures URL + PAT and installs the binary.
	ModeSetup Mode = iota
	// ModeLogin only refreshes the PAT (keeping the URL), or falls back to a
	// full setup when the product is not configured yet. It does not install.
	ModeLogin
)

// productSetup binds a product's display info to the URL/PAT fields it edits in
// the config (Confluence and Jira are distinct structs with identical fields).
type productSetup struct {
	label      string
	defaultURL string
	required   bool
	url        *string
	pat        *string
}

func confluenceProduct(cfg *config.Config) productSetup {
	return productSetup{
		label:      "Confluence",
		defaultURL: defaultConfluenceURL,
		required:   true,
		url:        &cfg.Confluence.URL,
		pat:        &cfg.Confluence.PAT,
	}
}

func jiraProduct(cfg *config.Config) productSetup {
	return productSetup{
		label:      "Jira",
		defaultURL: "",
		required:   false,
		url:        &cfg.Jira.URL,
		pat:        &cfg.Jira.PAT,
	}
}

func productFor(name string, cfg *config.Config) (productSetup, error) {
	switch name {
	case "confluence":
		return confluenceProduct(cfg), nil
	case "jira":
		return jiraProduct(cfg), nil
	default:
		return productSetup{}, fmt.Errorf("unknown product %q (use 'confluence' or 'jira')", name)
	}
}

// loadForEdit reads and migrates the existing config (or returns an empty one).
func loadForEdit() (*config.Config, error) {
	cfg, found, err := config.Read()
	if err != nil {
		return nil, fmt.Errorf("cannot read existing config: %w", err)
	}
	if cfg == nil {
		cfg = &config.Config{}
	}
	if found {
		config.Migrate(cfg)
	}
	return cfg, nil
}

// Run sets up both products: it installs the binary and prompts only for the
// product(s) not yet configured. Confluence is required; Jira is optional and
// can be skipped by pressing Enter at its URL prompt.
func Run() error {
	if err := installSelf(); err != nil {
		return fmt.Errorf("install binary: %w", err)
	}

	cfg, err := loadForEdit()
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("=== Confluence MCP Setup ===")

	conf := confluenceProduct(cfg)
	if *conf.url == "" || *conf.pat == "" {
		fmt.Println()
		promptProduct(reader, conf)
	}
	if *conf.pat == "" {
		return fmt.Errorf("confluence personal access token is required")
	}

	jiraP := jiraProduct(cfg)
	if *jiraP.url == "" || *jiraP.pat == "" {
		fmt.Println()
		fmt.Println("--- Jira (optional; press Enter at the URL to skip) ---")
		promptProduct(reader, jiraP)
	}

	return finish(cfg, true)
}

// Product configures a single product in the given mode.
func Product(name string, mode Mode) error {
	if mode == ModeSetup {
		if err := installSelf(); err != nil {
			return fmt.Errorf("install binary: %w", err)
		}
	}

	cfg, err := loadForEdit()
	if err != nil {
		return err
	}

	p, err := productFor(name, cfg)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	switch mode {
	case ModeLogin:
		fmt.Printf("=== %s login ===\n\n", p.label)
		// Refresh only the token when already configured; otherwise behave like
		// setup and ask for the URL too.
		if *p.url == "" {
			*p.url = promptURL(reader, p.label, *p.url, p.defaultURL)
		} else {
			fmt.Printf("%s URL: %s\n", p.label, *p.url)
		}
		*p.pat = promptPAT(reader, "New "+p.label+" Personal Access Token", *p.pat)
	default: // ModeSetup
		fmt.Printf("=== %s setup ===\n\n", p.label)
		promptProduct(reader, p)
	}

	if *p.url == "" {
		return fmt.Errorf("%s url is required", strings.ToLower(p.label))
	}
	if *p.pat == "" {
		return fmt.Errorf("%s personal access token is required", strings.ToLower(p.label))
	}

	return finish(cfg, mode == ModeSetup)
}

// promptProduct asks for a product's URL then PAT, writing them through the
// productSetup pointers. For an optional product, an empty URL skips it.
func promptProduct(reader *bufio.Reader, p productSetup) {
	url := promptURL(reader, p.label, *p.url, p.defaultURL)
	if !p.required && url == "" {
		return // optional product skipped
	}
	*p.url = url
	*p.pat = promptPAT(reader, p.label+" Personal Access Token", *p.pat)
}

// promptURL asks for a URL, defaulting to the current value, or the product
// default, or empty (skip).
func promptURL(reader *bufio.Reader, label, current, def string) string {
	if current != "" {
		def = current
	}

	if def == "" {
		fmt.Printf("%s URL (Enter to skip): ", label)
	} else {
		fmt.Printf("%s URL [%s]: ", label, def)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return def
	}
	return input
}

// promptPAT asks for a Personal Access Token. When one already exists, an empty
// answer keeps it.
func promptPAT(reader *bufio.Reader, label, current string) string {
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

// finish stamps the schema version, saves the config, and prints the connection
// info (with the opencode snippet when withSnippet is true).
func finish(cfg *config.Config, withSnippet bool) error {
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
	fmt.Printf("Config: %s\n", configPath)
	fmt.Printf("Binary: %s\n", binPath)

	if withSnippet {
		fmt.Println()
		fmt.Println("Add this to your opencode config:")
		fmt.Println()
		fmt.Println(`"confluence": {`)
		fmt.Println(`  "type": "local",`)
		fmt.Printf(`  "command": [%q]`+"\n", binPath)
		fmt.Println(`}`)
	}

	return nil
}

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

	src, err := os.Open(exePath)
	if err != nil {
		return err
	}
	defer src.Close()

	return writeBinary(targetPath, src)
}

// writeBinary writes r to target as an executable. It writes to a temp file then
// atomically renames over the target: a plain truncate-in-place fails with "text
// file busy" (ETXTBSY) when the existing binary is currently running (e.g. during
// an upgrade or self-update); rename does not.
func writeBinary(target string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	tmpPath := target + ".new"
	dst, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}

	if _, err := io.Copy(dst, r); err != nil {
		dst.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := dst.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, target); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}
