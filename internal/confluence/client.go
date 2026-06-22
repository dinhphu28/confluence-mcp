// Package confluence is a minimal client for the Confluence Server/Data Center
// REST API.
package confluence

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client talks to a Confluence instance using a Personal Access Token.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// NewClient returns a Client for the given base URL and Personal Access Token.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		http:    http.DefaultClient,
	}
}

// do performs an authenticated request against an API path. A non-empty body is
// sent as JSON. It returns the raw response body.
func (c *Client) do(method, path, body string) (string, error) {
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("confluence error %d: %s", resp.StatusCode, string(respBody))
	}

	return string(respBody), nil
}

// get performs an authenticated GET and returns the raw response body.
func (c *Client) get(path string) (string, error) {
	return c.do(http.MethodGet, path, "")
}

// getBytes performs an authenticated GET and returns the raw response bytes.
// Used for binary resources such as attachment downloads.
func (c *Client) getBytes(path string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("confluence error %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}
