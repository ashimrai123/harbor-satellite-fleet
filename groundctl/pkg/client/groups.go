package client

import (
	"fmt"
	"time"
)

// Group represents a group in Ground Control.
type Group struct {
	ID          int32     `json:"ID"`
	GroupName   string    `json:"GroupName"`
	RegistryURL string   `json:"RegistryUrl"`
	Projects    []string  `json:"Projects,omitempty"`
	CreatedAt   time.Time `json:"CreatedAt"`
	UpdatedAt   time.Time `json:"UpdatedAt"`
}

// GroupArtifact represents an artifact in a group sync request.
type GroupArtifact struct {
	Repository string            `json:"repository"`
	Tag        []string          `json:"tag,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Type       string            `json:"type,omitempty"`
	Digest     string            `json:"digest,omitempty"`
	Deleted    bool              `json:"deleted,omitempty"`
}

// GroupSyncRequest is the payload for syncing a group.
type GroupSyncRequest struct {
	Group     string          `json:"group"`
	Registry  string          `json:"registry,omitempty"`
	Artifacts []GroupArtifact `json:"artifacts,omitempty"`
}

// SatelliteGroupRequest is the payload for adding/removing a satellite from a group.
type SatelliteGroupRequest struct {
	Satellite string `json:"satellite"`
	Group     string `json:"group"`
}

// ListGroups returns all groups.
func (c *Client) ListGroups() ([]Group, error) {
	var result []Group
	err := c.doGet("/api/groups", &result)
	return result, err
}

// GetGroup returns a group by name.
func (c *Client) GetGroup(name string) (Group, error) {
	var result Group
	err := c.doGet(fmt.Sprintf("/api/groups/%s", name), &result)
	return result, err
}

// SyncGroup creates or updates a group with its artifact list.
func (c *Client) SyncGroup(req GroupSyncRequest) (Group, error) {
	var result Group
	_, err := c.doPost("/api/groups/sync", req, &result)
	return result, err
}

// DeleteGroup deletes a group by name (requires admin).
func (c *Client) DeleteGroup(name string) error {
	return c.doDelete(fmt.Sprintf("/api/groups/%s", name))
}

// ListGroupSatellites returns satellites in a group.
func (c *Client) ListGroupSatellites(name string) ([]Satellite, error) {
	var result []Satellite
	err := c.doGet(fmt.Sprintf("/api/groups/%s/satellites", name), &result)
	return result, err
}

// AddSatelliteToGroup adds a satellite to a group.
func (c *Client) AddSatelliteToGroup(satellite, group string) error {
	req := SatelliteGroupRequest{Satellite: satellite, Group: group}
	_, err := c.doPost("/api/groups/satellite", req, nil)
	return err
}

// RemoveSatelliteFromGroup removes a satellite from a group.
func (c *Client) RemoveSatelliteFromGroup(satellite, group string) error {
	req := SatelliteGroupRequest{Satellite: satellite, Group: group}
	return c.doDeleteWithBody("/api/groups/satellite", req)
}
