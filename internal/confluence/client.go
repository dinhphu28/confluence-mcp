// Package confluence is a minimal client for the Confluence Server/Data Center
// REST API.
package confluence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
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

// bodyField builds the Confluence REST `body` payload for the given content
// representation (e.g. "storage" or "wiki").
func bodyField(representation, content string) map[string]any {
	return map[string]any{
		representation: map[string]any{
			"value":          content,
			"representation": representation,
		},
	}
}

// Search runs a CQL query and returns up to limit results.
func (c *Client) Search(cql string, limit int) (string, error) {
	path := fmt.Sprintf("/rest/api/content/search?limit=%d&expand=space,version&cql=%s",
		limit, url.QueryEscape(cql))
	return c.get(path)
}

// GetPage returns a single page by ID, including its storage-format body.
func (c *Client) GetPage(pageID string) (string, error) {
	path := "/rest/api/content/" + url.PathEscape(pageID) +
		"?expand=space,version,body.storage"
	return c.get(path)
}

// GetPageChildren lists the child pages directly under a page.
func (c *Client) GetPageChildren(pageID string, limit int) (string, error) {
	path := fmt.Sprintf("/rest/api/content/%s/child/page?limit=%d&expand=space,version",
		url.PathEscape(pageID), limit)
	return c.get(path)
}

// GetComments returns the comments on a page.
func (c *Client) GetComments(pageID string, limit int) (string, error) {
	path := fmt.Sprintf("/rest/api/content/%s/child/comment?limit=%d&expand=body.storage,version",
		url.PathEscape(pageID), limit)
	return c.get(path)
}

// CreatePage creates a new page. parentID is optional (empty for a top-level
// page). representation is the body format, e.g. "storage" or "wiki".
func (c *Client) CreatePage(spaceKey, title, content, parentID, representation string) (string, error) {
	payload := map[string]any{
		"type":  "page",
		"title": title,
		"space": map[string]any{"key": spaceKey},
		"body":  bodyField(representation, content),
	}
	if parentID != "" {
		payload["ancestors"] = []map[string]any{{"id": parentID}}
	}

	body, _ := json.Marshal(payload)
	return c.do(http.MethodPost, "/rest/api/content", string(body))
}

// UpdatePage updates an existing page. The current version and space are fetched
// automatically and the version is bumped. An empty title keeps the existing one.
func (c *Client) UpdatePage(pageID, content, title, representation string) (string, error) {
	raw, err := c.get("/rest/api/content/" + url.PathEscape(pageID) + "?expand=version,space")
	if err != nil {
		return "", err
	}

	var current struct {
		Title string `json:"title"`
		Space struct {
			Key string `json:"key"`
		} `json:"space"`
		Version struct {
			Number int `json:"number"`
		} `json:"version"`
	}
	if err := json.Unmarshal([]byte(raw), &current); err != nil {
		return "", fmt.Errorf("cannot parse current page: %w", err)
	}

	if title == "" {
		title = current.Title
	}

	payload := map[string]any{
		"id":      pageID,
		"type":    "page",
		"title":   title,
		"space":   map[string]any{"key": current.Space.Key},
		"version": map[string]any{"number": current.Version.Number + 1},
		"body":    bodyField(representation, content),
	}

	body, _ := json.Marshal(payload)
	return c.do(http.MethodPut, "/rest/api/content/"+url.PathEscape(pageID), string(body))
}

// AddComment posts a comment on a page.
func (c *Client) AddComment(pageID, content, representation string) (string, error) {
	payload := map[string]any{
		"type":      "comment",
		"container": map[string]any{"id": pageID, "type": "page"},
		"body":      bodyField(representation, content),
	}

	body, _ := json.Marshal(payload)
	return c.do(http.MethodPost, "/rest/api/content", string(body))
}

// DeletePage deletes a page by ID (moves it to the trash).
func (c *Client) DeletePage(pageID string) error {
	_, err := c.do(http.MethodDelete, "/rest/api/content/"+url.PathEscape(pageID), "")
	return err
}

// GetAttachments lists the attachments on a page, including each one's media
// type and version.
func (c *Client) GetAttachments(pageID string, limit int) (string, error) {
	path := fmt.Sprintf("/rest/api/content/%s/child/attachment?limit=%d&expand=version,metadata",
		url.PathEscape(pageID), limit)
	return c.get(path)
}

// Attachment is a downloaded attachment's bytes plus the metadata needed to
// present it.
type Attachment struct {
	Filename  string
	MediaType string
	Data      []byte
}

// DownloadAttachment fetches an attachment's bytes by its content ID.
func (c *Client) DownloadAttachment(attachmentID string) (*Attachment, error) {
	raw, err := c.get("/rest/api/content/" + url.PathEscape(attachmentID) + "?expand=metadata")
	if err != nil {
		return nil, err
	}

	var meta struct {
		Title    string `json:"title"`
		Metadata struct {
			MediaType string `json:"mediaType"`
		} `json:"metadata"`
		Links struct {
			Download string `json:"download"`
		} `json:"_links"`
	}
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return nil, fmt.Errorf("cannot parse attachment %s: %w", attachmentID, err)
	}
	if meta.Links.Download == "" {
		return nil, fmt.Errorf("attachment %s has no download link", attachmentID)
	}

	// The download link is relative to the instance base, not /rest/api.
	data, err := c.getBytes(meta.Links.Download)
	if err != nil {
		return nil, err
	}

	return &Attachment{
		Filename:  meta.Title,
		MediaType: meta.Metadata.MediaType,
		Data:      data,
	}, nil
}

// UploadAttachment uploads bytes as an attachment named filename on a page.
func (c *Client) UploadAttachment(pageID, filename string, data []byte) (string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err := fw.Write(data); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost,
		c.baseURL+"/rest/api/content/"+url.PathEscape(pageID)+"/child/attachment", &buf)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", w.FormDataContentType())
	// Required by Confluence for attachment uploads (XSRF check bypass).
	req.Header.Set("X-Atlassian-Token", "nocheck")

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
