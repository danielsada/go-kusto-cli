package client

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

const queryPath = "/v2/rest/query"

// MaxResponseSize is the maximum response body size (100 MB).
const MaxResponseSize = 100 * 1024 * 1024

// Client communicates with the Kusto REST API v2.
type Client struct {
	HTTPClient *http.Client
	ClusterURL string
	token      string
}

// New creates a new Kusto REST API client.
func New(clusterURL, token string) *Client {
	url := strings.TrimRight(clusterURL, "/")
	return &Client{
		ClusterURL: url,
		token:      token,
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
	}
}

type queryRequest struct {
	DB  string `json:"db"`
	CSL string `json:"csl"`
}

// Query executes a KQL query and returns the raw v2 JSON response frames.
func (c *Client) Query(ctx context.Context, database, query string) ([]Frame, error) {
	body, err := json.Marshal(queryRequest{DB: database, CSL: query})
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.ClusterURL+queryPath, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	limitedReader := io.LimitReader(resp.Body, MaxResponseSize+1)
	respBody, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if int64(len(respBody)) > MaxResponseSize {
		return nil, fmt.Errorf("response body exceeds maximum size of %d bytes", MaxResponseSize)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 500))
	}

	frames, err := ParseFrames(respBody)
	if err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return frames, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
