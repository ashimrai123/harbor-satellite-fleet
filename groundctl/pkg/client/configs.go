package client

import (
	"encoding/json"
	"fmt"
	"time"
)

// Config represents a satellite configuration in Ground Control.
type Config struct {
	ID          int32           `json:"ID"`
	ConfigName  string          `json:"ConfigName"`
	RegistryURL string          `json:"RegistryUrl"`
	ConfigData  json.RawMessage `json:"Config,omitempty"`
	CreatedAt   time.Time       `json:"CreatedAt"`
	UpdatedAt   time.Time       `json:"UpdatedAt"`
}

// CreateConfigRequest is the payload for creating a config.
type CreateConfigRequest struct {
	ConfigName string          `json:"config_name"`
	Registry   string          `json:"registry,omitempty"`
	ConfigData json.RawMessage `json:"config,omitempty"`
}

// SatelliteConfigRequest is the payload for assigning a config to a satellite.
type SatelliteConfigRequest struct {
	Satellite  string `json:"satellite"`
	ConfigName string `json:"config_name"`
}

// ListConfigs returns all configs.
func (c *Client) ListConfigs() ([]Config, error) {
	var result []Config
	err := c.doGet("/api/configs", &result)
	return result, err
}

// GetConfig returns a config by name.
func (c *Client) GetConfig(name string) (Config, error) {
	var result Config
	err := c.doGet(fmt.Sprintf("/api/configs/%s", name), &result)
	return result, err
}

// CreateConfig creates a new config.
func (c *Client) CreateConfig(req CreateConfigRequest) error {
	_, err := c.doPost("/api/configs", req, nil)
	return err
}

// DeleteConfig deletes a config by name.
func (c *Client) DeleteConfig(name string) error {
	return c.doDelete(fmt.Sprintf("/api/configs/%s", name))
}

// SetSatelliteConfig assigns a config to a satellite.
func (c *Client) SetSatelliteConfig(satellite, configName string) error {
	req := SatelliteConfigRequest{Satellite: satellite, ConfigName: configName}
	_, err := c.doPost("/api/configs/satellite", req, nil)
	return err
}
