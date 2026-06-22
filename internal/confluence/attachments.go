package confluence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

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
