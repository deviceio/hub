package infra

// ClusterDeviceModel
type ClusterDeviceModel struct {
	id           string
	hostname     string
	architecture string
	platform     string
	isConnected  bool
	tags         []string
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
