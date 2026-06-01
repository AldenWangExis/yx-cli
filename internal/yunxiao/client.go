package yunxiao

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ClientConfig struct {
	BaseURL        string
	Token          string
	OrganizationID string
	Region         string
	HTTPClient     *http.Client
}

type Client struct {
	baseURL        string
	token          string
	organizationID string
	region         string
	httpClient     *http.Client
}

func NewClient(config ClientConfig) *Client {
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:        NormalizeBaseURL(config.BaseURL),
		token:          config.Token,
		organizationID: config.OrganizationID,
		region:         config.Region,
		httpClient:     httpClient,
	}
}

func (c *Client) OrganizationID() string {
	return c.organizationID
}

func (c *Client) IsCenter() bool {
	return c.region == "" || c.region == "center"
}

func (c *Client) DoJSON(ctx context.Context, method, path string, query url.Values, body []byte) ([]byte, error) {
	endpoint := c.baseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("x-yunxiao-token", c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("yunxiao request failed: %w", err)
	}
	defer resp.Body.Close()

	data, readErr := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, APIError{
			StatusCode: resp.StatusCode,
			RequestID:  resp.Header.Get("x-request-id"),
			Body:       redactSecret(string(data), c.token),
		}
	}
	if readErr != nil {
		return nil, fmt.Errorf("read response: %w", readErr)
	}
	return data, nil
}

func redactSecret(value, secret string) string {
	if secret == "" {
		return value
	}
	return strings.ReplaceAll(value, secret, "[REDACTED]")
}
