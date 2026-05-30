package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agenthub/internal/hub"
)

type Config struct {
	BaseURL string
	Token   string
	Timeout time.Duration
}

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(cfg Config) *Client {
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if base == "" {
		base = "http://localhost:8080"
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Client{
		baseURL: base,
		token:   strings.TrimSpace(cfg.Token),
		http:    &http.Client{Timeout: timeout},
	}
}

type apiResp struct {
	Code   int             `json:"code"`
	ErrMsg string          `json:"errMsg"`
	Body   json.RawMessage `json:"body"`
}

func (c *Client) ListAgents(category string) ([]hub.AgentMeta, error) {
	path := "/api/hub/agents"
	if cat := strings.TrimSpace(category); cat != "" {
		path += "?category=" + url.QueryEscape(cat)
	}
	var out struct {
		Agents []hub.AgentMeta `json:"agents"`
		Total  int             `json:"total"`
	}
	if err := c.getJSON(path, &out); err != nil {
		return nil, err
	}
	return out.Agents, nil
}

func (c *Client) ListCategories() ([]string, error) {
	var out struct {
		Categories []string `json:"categories"`
	}
	if err := c.getJSON("/api/hub/categories", &out); err != nil {
		return nil, err
	}
	return out.Categories, nil
}

func (c *Client) GetAgent(agentName, version string) (hub.AgentDetail, error) {
	path := "/api/hub/agents/" + url.PathEscape(agentName)
	if v := strings.TrimSpace(version); v != "" {
		path += "?version=" + url.QueryEscape(v)
	}
	var out hub.AgentDetail
	if err := c.getJSON(path, &out); err != nil {
		return hub.AgentDetail{}, err
	}
	return out, nil
}

type UpdateMetaRequest struct {
	DisplayName string `json:"displayName,omitempty"`
	Summary     string `json:"summary,omitempty"`
	Category    string `json:"category,omitempty"`
}

func (c *Client) UpdateAgentMeta(agentName string, req UpdateMetaRequest) (hub.AgentMeta, error) {
	path := "/api/hub/agents/" + url.PathEscape(agentName)
	var out hub.AgentMeta
	if err := c.putJSON(path, req, &out); err != nil {
		return hub.AgentMeta{}, err
	}
	return out, nil
}

type UpdateFileRequest struct {
	Version string `json:"version"`
	Content string `json:"content"`
}

func (c *Client) UpdateFile(agentName, version, filePath string, content []byte) error {
	path := fmt.Sprintf("/api/hub/agents/%s/files/%s", url.PathEscape(agentName), escapePath(filePath))
	req := UpdateFileRequest{Version: version, Content: string(content)}
	return c.putJSON(path, req, nil)
}

func (c *Client) DeleteAgent(agentName string) error {
	path := "/api/hub/agents/" + url.PathEscape(agentName)
	return c.deleteJSON(path)
}

func (c *Client) GetPackageFile(agentName, version, path string) ([]byte, error) {
	u := fmt.Sprintf("/api/hub/agents/%s/files/%s", url.PathEscape(agentName), escapePath(path))
	if version != "" {
		u += "?version=" + url.QueryEscape(version)
	}
	var out struct {
		Content string `json:"content"`
	}
	if err := c.getJSON(u, &out); err != nil {
		return nil, err
	}
	return []byte(out.Content), nil
}

func (c *Client) DownloadAgent(agentName, version, dest string) error {
	u := fmt.Sprintf("%s/api/hub/agents/%s/download", c.baseURL, url.PathEscape(agentName))
	if version != "" {
		u += "?version=" + url.QueryEscape(version)
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return readAPIError(resp)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func (c *Client) UploadAgent(agentName, category, version, zipPath string) (hub.AgentMeta, error) {
	f, err := os.Open(zipPath)
	if err != nil {
		return hub.AgentMeta{}, err
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("agentName", agentName)
	if category != "" {
		_ = w.WriteField("category", category)
	}
	if version != "" {
		_ = w.WriteField("version", version)
	}
	part, err := w.CreateFormFile("file", filepath.Base(zipPath))
	if err != nil {
		return hub.AgentMeta{}, err
	}
	if _, err := io.Copy(part, f); err != nil {
		return hub.AgentMeta{}, err
	}
	if err := w.Close(); err != nil {
		return hub.AgentMeta{}, err
	}

	u := c.baseURL + "/api/hub/agents"
	req, err := http.NewRequest(http.MethodPost, u, &buf)
	if err != nil {
		return hub.AgentMeta{}, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	c.setAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return hub.AgentMeta{}, err
	}
	defer resp.Body.Close()

	var out struct {
		Agent hub.AgentMeta `json:"agent"`
	}
	if err := decodeAPI(resp, &out); err != nil {
		return hub.AgentMeta{}, err
	}
	return out.Agent, nil
}

func (c *Client) getJSON(path string, out any) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	return c.doJSON(req, out)
}

func (c *Client) putJSON(path string, body any, out any) error {
	var payload io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(b)
	}
	req, err := http.NewRequest(http.MethodPut, c.baseURL+path, payload)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c.setAuth(req)
	return c.doJSON(req, out)
}

func (c *Client) deleteJSON(path string) error {
	req, err := http.NewRequest(http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	c.setAuth(req)
	return c.doJSON(req, nil)
}

func (c *Client) doJSON(req *http.Request, out any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeAPI(resp, out)
}

func (c *Client) setAuth(req *http.Request) {
	if c.token == "" {
		return
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
}

func decodeAPI(resp *http.Response, out any) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var wrap apiResp
	if err := json.Unmarshal(body, &wrap); err != nil {
		if resp.StatusCode >= 400 {
			return fmt.Errorf("request failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}
		return fmt.Errorf("invalid response: %w", err)
	}
	if wrap.Code >= 400 || wrap.ErrMsg != "" {
		if wrap.ErrMsg != "" {
			return fmt.Errorf("%s", wrap.ErrMsg)
		}
		return fmt.Errorf("request failed: HTTP %d", wrap.Code)
	}
	if out == nil {
		return nil
	}
	if len(wrap.Body) == 0 || string(wrap.Body) == "null" {
		return nil
	}
	return json.Unmarshal(wrap.Body, out)
}

func readAPIError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var wrap apiResp
	if err := json.Unmarshal(body, &wrap); err == nil && wrap.ErrMsg != "" {
		return fmt.Errorf("%s", wrap.ErrMsg)
	}
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
}

func escapePath(p string) string {
	p = strings.TrimPrefix(p, "/")
	segments := strings.Split(p, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return strings.Join(segments, "/")
}
