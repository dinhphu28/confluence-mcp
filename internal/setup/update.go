package setup

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
)

const githubRepo = "dinhphu28/atlassian-mcp"

type release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

// Update checks GitHub for a newer release and replaces the installed binary in
// place. When checkOnly is true it only reports availability without installing.
func Update(currentVersion string, checkOnly bool) error {
	rel, err := latestRelease()
	if err != nil {
		return err
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")

	fmt.Printf("Installed: %s\n", currentVersion)
	fmt.Printf("Latest:    %s\n", rel.TagName)

	if !isNewer(latest, current) {
		fmt.Println("Already up to date.")
		return nil
	}

	if checkOnly {
		fmt.Printf("Update available: %s -> %s. Run 'atlassian-mcp update' to install.\n", current, latest)
		return nil
	}

	name, assetURL, err := assetForPlatform(rel)
	if err != nil {
		return err
	}

	fmt.Printf("Downloading %s ...\n", name)
	data, err := download(assetURL)
	if err != nil {
		return err
	}

	bin, err := extractBinary(name, data)
	if err != nil {
		return err
	}

	target, err := installPath()
	if err != nil {
		return err
	}
	if err := writeBinary(target, bytes.NewReader(bin)); err != nil {
		return err
	}

	fmt.Printf("\nUpdated to %s at %s\n", latest, target)
	fmt.Println("Restart/reconnect the MCP server to use the new version.")
	return nil
}

// latestRelease fetches the latest published release from GitHub.
func latestRelease() (*release, error) {
	url := "https://api.github.com/repos/" + githubRepo + "/releases/latest"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "atlassian-mcp")
	req.Header.Set("Accept", "application/vnd.github+json")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no published release found for %s", githubRepo)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("github error %d: %s", resp.StatusCode, string(body))
	}

	var rel release
	if err := json.Unmarshal(body, &rel); err != nil {
		return nil, fmt.Errorf("cannot parse release: %w", err)
	}
	return &rel, nil
}

// assetForPlatform returns the release asset matching the running OS/arch.
func assetForPlatform(rel *release) (name, url string, err error) {
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	suffix := fmt.Sprintf("_%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)

	for _, a := range rel.Assets {
		if strings.HasSuffix(a.Name, suffix) {
			return a.Name, a.URL, nil
		}
	}
	return "", "", fmt.Errorf("release %s has no asset for %s/%s", rel.TagName, runtime.GOOS, runtime.GOARCH)
}

// download fetches the bytes at url (following redirects).
func download(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "atlassian-mcp")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download error %d", resp.StatusCode)
	}
	return data, nil
}

// extractBinary returns the binary bytes from a release archive (.tar.gz or
// .zip). The archive contains <dir>/<binary>; the entry is matched by base name.
func extractBinary(assetName string, data []byte) ([]byte, error) {
	if strings.HasSuffix(assetName, ".zip") {
		return extractFromZip(data)
	}
	return extractFromTarGz(data)
}

func extractFromTarGz(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag == tar.TypeReg && path.Base(hdr.Name) == binaryName() {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("binary %q not found in archive", binaryName())
}

func extractFromZip(data []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	for _, f := range zr.File {
		if path.Base(f.Name) == binaryName() {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("binary %q not found in archive", binaryName())
}

// isNewer reports whether latest is a newer version than current. A "dev" build
// is always considered older. Versions are compared as dotted numeric fields.
func isNewer(latest, current string) bool {
	if current == "dev" || current == "" {
		return true
	}

	lf := versionFields(latest)
	cf := versionFields(current)
	for i := 0; i < len(lf) || i < len(cf); i++ {
		var l, c int
		if i < len(lf) {
			l = lf[i]
		}
		if i < len(cf) {
			c = cf[i]
		}
		if l != c {
			return l > c
		}
	}
	return false
}

func versionFields(v string) []int {
	parts := strings.Split(v, ".")
	out := make([]int, len(parts))
	for i, p := range parts {
		// Stop a field at the first non-numeric run (e.g. "1-rc2" -> 1).
		n := 0
		for n < len(p) && p[n] >= '0' && p[n] <= '9' {
			n++
		}
		out[i], _ = strconv.Atoi(p[:n])
	}
	return out
}
