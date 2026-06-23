package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// AttachmentMeta describes an attachment on an issue.
type AttachmentMeta struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	MimeType string `json:"mimeType"`
	Content  string `json:"content"` // download URL
	Size     int    `json:"size"`
}

// Attachment is a downloaded attachment's bytes plus presentation metadata.
type Attachment struct {
	Filename  string
	MediaType string
	Data      []byte
}

// IssueAttachments lists the attachments on an issue.
func (c *Client) IssueAttachments(key string) ([]AttachmentMeta, error) {
	raw, err := c.get("/rest/api/2/issue/" + url.PathEscape(key) + "?fields=attachment")
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Fields struct {
			Attachment []AttachmentMeta `json:"attachment"`
		} `json:"fields"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("cannot parse attachments for issue %s: %w", key, err)
	}
	return parsed.Fields.Attachment, nil
}

// DownloadAttachment fetches an attachment's bytes from its content URL.
func (c *Client) DownloadAttachment(meta AttachmentMeta) (*Attachment, error) {
	data, err := c.getBytesURL(meta.Content)
	if err != nil {
		return nil, err
	}
	return &Attachment{
		Filename:  meta.Filename,
		MediaType: meta.MimeType,
		Data:      data,
	}, nil
}

// getBytesURL performs an authenticated GET against an absolute (or root-relative)
// URL and returns the raw bytes. Attachment content URLs are absolute.
func (c *Client) getBytesURL(rawURL string) ([]byte, error) {
	target := rawURL
	if !strings.HasPrefix(rawURL, "http") {
		target = c.baseURL + rawURL
	}

	req, err := http.NewRequest(http.MethodGet, target, nil)
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
		return nil, fmt.Errorf("jira error %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}
