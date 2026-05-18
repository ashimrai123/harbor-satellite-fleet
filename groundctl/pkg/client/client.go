package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Client is the Ground Control API client.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// CLIConfig holds persistent CLI configuration.
type CLIConfig struct {
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`
}

// NewClient creates a new Ground Control client.
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientFromConfig loads the client from the saved config file.
func NewClientFromConfig() (*Client, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("not logged in. Run 'groundctl login' first: %w", err)
	}
	return NewClient(cfg.BaseURL, cfg.Token), nil
}

// Login authenticates with Ground Control and stores the token.
func (c *Client) Login(username, password string) error {
	body := map[string]string{
		"username": username,
		"password": password,
	}

	resp, err := c.doRequest("POST", "/login", body)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	c.Token = result.Token

	// Persist config
	cfg := CLIConfig{
		BaseURL: c.BaseURL,
		Token:   c.Token,
	}
	return SaveConfig(cfg)
}

// doRequest performs an authenticated HTTP request.
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	return c.HTTPClient.Do(req)
}

// doGet is a convenience wrapper for GET requests that decodes JSON.
func (c *Client) doGet(path string, result interface{}) error {
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed (status %d): %s", resp.StatusCode, string(errBody))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// doPost is a convenience wrapper for POST requests.
func (c *Client) doPost(path string, body, result interface{}) (int, error) {
	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode >= 400 {
		errBody, _ := io.ReadAll(resp.Body)
		return statusCode, fmt.Errorf("request failed (status %d): %s", statusCode, string(errBody))
	}

	if result != nil {
		return statusCode, json.NewDecoder(resp.Body).Decode(result)
	}
	return statusCode, nil
}

// doDelete is a convenience wrapper for DELETE requests.
func (c *Client) doDelete(path string) error {
	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed (status %d): %s", resp.StatusCode, string(errBody))
	}

	return nil
}

// configDir returns the path to the groundctl config directory.
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".groundctl")
	return dir, os.MkdirAll(dir, 0o700)
}

// SaveConfig persists the client configuration.
func SaveConfig(cfg CLIConfig) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "config.json"), data, 0o600)
}

// LoadConfig loads the client configuration.
func LoadConfig() (CLIConfig, error) {
	dir, err := configDir()
	if err != nil {
		return CLIConfig{}, err
	}

	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		return CLIConfig{}, err
	}

	var cfg CLIConfig
	return cfg, json.Unmarshal(data, &cfg)
}
