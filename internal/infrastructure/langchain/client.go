package langchain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
}

func NewClient(timeout time.Duration) *Client {
	transport := &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
		IdleConnTimeout:     90 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
	}
	return &Client{
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}
}

type executeRequest struct {
	Input      string                 `json:"input"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	SessionID  string                 `json:"session_id,omitempty"`
}

type ExecuteResult struct {
	StatusCode int
	Body       []byte
	Duration   time.Duration
}

func (c *Client) Execute(ctx context.Context, baseURL, agentID, apiKey, userMessage, sessionID string, params map[string]interface{}) (*ExecuteResult, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("langchain base URL is required")
	}
	if agentID == "" {
		return nil, fmt.Errorf("agentID is required")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("langchain API key is required")
	}

	url := buildExecuteURL(baseURL, agentID)
	reqPayload := executeRequest{
		Input:      userMessage,
		Parameters: params,
		SessionID:  sessionID,
	}

	body, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &ExecuteResult{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Duration:   time.Since(start),
	}, nil
}

func trimTrailingSlash(u string) string {
	if len(u) > 0 && u[len(u)-1] == '/' {
		return u[:len(u)-1]
	}
	return u
}

func buildExecuteURL(baseURL, agentID string) string {
	u := trimTrailingSlash(baseURL)

	if strings.Contains(u, "/agents/") {
		// Already full path, just ensure no trailing slash then append execute if needed.
		if !strings.HasSuffix(u, "/execute") {
			u = u + "/execute"
		}
		return u
	}

	// If base already includes /api/v*, keep it, otherwise default to /api/v1
	if strings.Contains(u, "/api/v") {
		return fmt.Sprintf("%s/agents/%s/execute", u, agentID)
	}
	return fmt.Sprintf("%s/api/v1/agents/%s/execute", u, agentID)
}
