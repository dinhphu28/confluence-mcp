package confluence

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

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
