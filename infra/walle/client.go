package walle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultTimeout = 15 * time.Second

// Client calls Walle OpenAPI endpoints.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a Walle API client. baseURL is e.g. http://ingress.9kfun.xyz/walle-api .
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		token:   strings.TrimSpace(token),
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// WithHTTPClient overrides the default HTTP client (mainly for tests).
func (c *Client) WithHTTPClient(httpClient *http.Client) *Client {
	if httpClient != nil {
		c.httpClient = httpClient
	}
	return c
}

// GetGameGroup fetches config for a single game group.
func (c *Client) GetGameGroup(ctx context.Context, group string) (*GameGroup, error) {
	name, err := ParseGroupName(group)
	if err != nil {
		return nil, err
	}
	groups, err := c.GetGameGroups(ctx, []string{name})
	if err != nil {
		return nil, fmt.Errorf("group %q: %w", name, err)
	}
	selected, err := SelectGameGroup(groups, name)
	if err != nil {
		return nil, fmt.Errorf("group %q: %w", name, err)
	}
	return selected, nil
}

// GetGameGroups fetches game group configs for the given group names.
func (c *Client) GetGameGroups(ctx context.Context, groups []string) ([]GameGroup, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c.baseURL == "" {
		return nil, fmt.Errorf("walle base url is empty")
	}
	if c.token == "" {
		return nil, fmt.Errorf("walle token is empty")
	}

	reqURL, err := c.gameGroupURL(groups)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s failed: %w", reqURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("read response from GET %s: %w", reqURL, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, httpStatusError(http.MethodGet, reqURL, resp.StatusCode, body)
	}

	var envelope Response
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("decode JSON from GET %s (HTTP %d): %w; body=%s", reqURL, resp.StatusCode, err, truncateBody(body))
	}
	if envelope.Status != "ok" {
		msg := strings.TrimSpace(envelope.Message)
		if msg == "" {
			msg = "unknown walle error"
		}
		return nil, fmt.Errorf("GET %s API status=%q message=%s", reqURL, envelope.Status, msg)
	}
	return envelope.Data, nil
}

func httpStatusError(method, reqURL string, statusCode int, body []byte) error {
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		msg = http.StatusText(statusCode)
	}
	return fmt.Errorf("%s %s returned HTTP %d: %s", method, reqURL, statusCode, truncateBody([]byte(msg)))
}

func truncateBody(body []byte) string {
	const maxLen = 512
	text := strings.TrimSpace(string(body))
	if text == "" {
		return "(empty body)"
	}
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

func (c *Client) gameGroupURL(groups []string) (string, error) {
	u, err := url.Parse(c.baseURL + "/openapi/game/group")
	if err != nil {
		return "", fmt.Errorf("invalid walle base url: %w", err)
	}
	q := u.Query()
	if len(groups) > 0 {
		q.Set("group", strings.Join(groups, ","))
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}
