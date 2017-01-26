package infra

import "github.com/deviceio/shared/protocol_v1"

// ClusterDeviceModel
type ClusterDeviceModel struct {
	id           string
	hostname     string
	architecture string
	platform     string
	isConnected  bool
	tags         []string
}

// NewClusterDeviceModelFromHandshake ...
func NewClusterDeviceModelFromHandshake(handshake *protocol_v1.Handshake, isConnected bool) *ClusterDeviceModel {
	return &ClusterDeviceModel{
		architecture: handshake.Architecture,
		hostname:     handshake.Hostname,
		id:           handshake.AgentID,
		isConnected:  isConnected,
		platform:     handshake.Platform,
		tags:         handshake.Tags,
	}
}

// ID ...
func (t *ClusterDeviceModel) ID() string {
	return t.id
}

// Hostname ...
func (t *ClusterDeviceModel) Hostname() string {
	return t.hostname
}

// Architecture ...
func (t *ClusterDeviceModel) Architecture() string {
	return t.architecture
}

// Platform ...
func (t *ClusterDeviceModel) Platform() string {
	return t.platform
}

// IsConnected ...
func (t *ClusterDeviceModel) IsConnected() bool {
	return t.isConnected
}

// Tags ...
func (t *ClusterDeviceModel) Tags() []string {
	return t.tags
}
