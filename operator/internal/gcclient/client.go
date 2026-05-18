// Package gcclient provides a Ground Control API client for the operator.
// It reuses the same HTTP patterns as the groundctl CLI client.
package gcclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the Ground Control API client.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
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

// Login authenticates with Ground Control and stores the bearer token.
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
	return nil
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

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
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

func (c *Client) doGet(path string, result interface{}) error {
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s failed (%d): %s", path, resp.StatusCode, b)
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) doPost(path string, body, result interface{}) error {
	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("POST %s failed (%d): %s", path, resp.StatusCode, b)
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *Client) doDelete(path string) error {
	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DELETE %s failed (%d): %s", path, resp.StatusCode, b)
	}
	return nil
}

// --- Satellite API ---

// Satellite is the Ground Control satellite model.
type Satellite struct {
	ID   int32  `json:"ID"`
	Name string `json:"Name"`
}

// RegisterSatelliteRequest is the payload for registering a satellite.
type RegisterSatelliteRequest struct {
	Name       string   `json:"name"`
	Groups     []string `json:"groups,omitempty"`
	ConfigName string   `json:"config_name"`
}

// RegisterSatelliteResponse contains the ZTR bootstrap token.
type RegisterSatelliteResponse struct {
	Token string `json:"token"`
}

// GetSatellite returns a satellite by name.
func (c *Client) GetSatellite(name string) (Satellite, error) {
	var result Satellite
	return result, c.doGet(fmt.Sprintf("/api/satellites/%s", name), &result)
}

// RegisterSatellite registers a new satellite and returns its ZTR token.
func (c *Client) RegisterSatellite(req RegisterSatelliteRequest) (RegisterSatelliteResponse, error) {
	var result RegisterSatelliteResponse
	return result, c.doPost("/api/satellites", req, &result)
}

// DeleteSatellite deletes a satellite by name.
func (c *Client) DeleteSatellite(name string) error {
	return c.doDelete(fmt.Sprintf("/api/satellites/%s", name))
}

// --- Group API ---

// AddSatelliteToGroup adds a satellite to a group.
func (c *Client) AddSatelliteToGroup(satellite, group string) error {
	req := map[string]string{"satellite": satellite, "group": group}
	return c.doPost("/api/groups/satellite", req, nil)
}

// --- Config API ---

// SetSatelliteConfig assigns a config to a satellite.
func (c *Client) SetSatelliteConfig(satellite, configName string) error {
	req := map[string]string{"satellite": satellite, "config_name": configName}
	return c.doPost("/api/configs/satellite", req, nil)
}
