// Package mcpserver exposes the HRIS as a read-only MCP server, so an AI
// assistant can answer questions about attendance and leave.
//
// It talks to the existing REST API rather than the database directly. That
// costs one HTTP hop but keeps a single source of truth: role checks, filters
// and report logic stay in the API, and this package cannot drift away from
// them or quietly bypass a permission the API enforces.
package mcpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var ErrUnauthorized = errors.New("kredensial HRIS ditolak")

// Client is a thin authenticated wrapper over the HRIS REST API. Access tokens
// last 15 minutes, so it logs in lazily and retries once on a 401 rather than
// refreshing on a timer.
type Client struct {
	baseURL  string
	email    string
	password string
	http     *http.Client

	mu    sync.Mutex
	token string
}

func NewClient(baseURL, email, password string) *Client {
	return &Client{
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		email:    email,
		password: password,
		http:     &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *Client) login() error {
	body, err := json.Marshal(map[string]string{"email": c.email, "password": c.password})
	if err != nil {
		return err
	}
	resp, err := c.http.Post(c.baseURL+"/api/v1/auth/login", "application/json", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("tidak bisa menghubungi API HRIS di %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login gagal (HTTP %d)", resp.StatusCode)
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	c.token = out.AccessToken
	return nil
}

// get calls a read endpoint, logging in first when there is no token yet and
// once more if the token expired mid-session.
func (c *Client) get(path string, params url.Values, out any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token == "" {
		if err := c.login(); err != nil {
			return err
		}
	}

	body, status, err := c.do(path, params)
	if err != nil {
		return err
	}
	if status == http.StatusUnauthorized {
		if err := c.login(); err != nil {
			return err
		}
		if body, status, err = c.do(path, params); err != nil {
			return err
		}
	}
	switch {
	case status == http.StatusForbidden:
		return fmt.Errorf("akun MCP tidak punya akses ke %s -- pakai akun HR atau superadmin", path)
	case status != http.StatusOK:
		return fmt.Errorf("permintaan ke %s gagal (HTTP %d): %s", path, status, strings.TrimSpace(string(body)))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(body, out)
}

func (c *Client) do(path string, params url.Values) ([]byte, int, error) {
	u := c.baseURL + "/api/v1" + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("tidak bisa menghubungi API HRIS di %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close() //nolint:errcheck
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}
