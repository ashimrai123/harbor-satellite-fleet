package client

import (
	"fmt"
	"time"
)

// Satellite represents a satellite in Ground Control.
type Satellite struct {
	ID                int32      `json:"ID"`
	Name              string     `json:"Name"`
	CreatedAt         time.Time  `json:"CreatedAt"`
	UpdatedAt         time.Time  `json:"UpdatedAt"`
	LastSeen          NullTime   `json:"LastSeen"`
	HeartbeatInterval NullString `json:"HeartbeatInterval"`
}

// SatelliteStatus represents the latest status of a satellite.
type SatelliteStatus struct {
	ID                 int32      `json:"ID"`
	SatelliteID        int32      `json:"SatelliteID"`
	Activity           string     `json:"Activity"`
	LatestStateDigest  NullString `json:"LatestStateDigest"`
	LatestConfigDigest NullString `json:"LatestConfigDigest"`
	CPUPercent         NullString `json:"CpuPercent"`
	MemoryUsedBytes    NullInt64  `json:"MemoryUsedBytes"`
	StorageUsedBytes   NullInt64  `json:"StorageUsedBytes"`
	LastSyncDurationMs NullInt64  `json:"LastSyncDurationMs"`
	ImageCount         NullInt32  `json:"ImageCount"`
	ReportedAt         time.Time  `json:"ReportedAt"`
	CreatedAt          time.Time  `json:"CreatedAt"`
}

// RegisterSatelliteRequest is the payload for registering a satellite.
type RegisterSatelliteRequest struct {
	Name       string   `json:"name"`
	Groups     []string `json:"groups,omitempty"`
	ConfigName string   `json:"config_name"`
}

// RegisterSatelliteResponse contains the ZTR token.
type RegisterSatelliteResponse struct {
	Token string `json:"token"`
}

// CachedArtifact represents a cached image on a satellite.
type CachedArtifact struct {
	ID        int32     `json:"ID"`
	Reference string    `json:"Reference"`
	SizeBytes int64     `json:"SizeBytes"`
	CreatedAt time.Time `json:"CreatedAt"`
}

// ListSatellites returns all satellites.
func (c *Client) ListSatellites() ([]Satellite, error) {
	var result []Satellite
	err := c.doGet("/api/satellites", &result)
	return result, err
}

// GetSatellite returns a satellite by name.
func (c *Client) GetSatellite(name string) (Satellite, error) {
	var result Satellite
	err := c.doGet(fmt.Sprintf("/api/satellites/%s", name), &result)
	return result, err
}

// RegisterSatellite registers a new satellite.
func (c *Client) RegisterSatellite(req RegisterSatelliteRequest) (RegisterSatelliteResponse, error) {
	var result RegisterSatelliteResponse
	_, err := c.doPost("/api/satellites", req, &result)
	return result, err
}

// DeleteSatellite deletes a satellite by name.
func (c *Client) DeleteSatellite(name string) error {
	return c.doDelete(fmt.Sprintf("/api/satellites/%s", name))
}

// GetSatelliteStatus returns the latest status for a satellite.
func (c *Client) GetSatelliteStatus(name string) (SatelliteStatus, error) {
	var result SatelliteStatus
	err := c.doGet(fmt.Sprintf("/api/satellites/%s/status", name), &result)
	return result, err
}

// GetActiveSatellites returns recently active satellites.
func (c *Client) GetActiveSatellites() ([]Satellite, error) {
	var result []Satellite
	err := c.doGet("/api/satellites/active", &result)
	return result, err
}

// GetStaleSatellites returns stale satellites.
func (c *Client) GetStaleSatellites() ([]Satellite, error) {
	var result []Satellite
	err := c.doGet("/api/satellites/stale", &result)
	return result, err
}

// GetCachedImages returns cached images for a satellite.
func (c *Client) GetCachedImages(name string) ([]CachedArtifact, error) {
	var result []CachedArtifact
	err := c.doGet(fmt.Sprintf("/api/satellites/%s/images", name), &result)
	return result, err
}
