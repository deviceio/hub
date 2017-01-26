package domain

import (
	"net/http"

	"github.com/deviceio/hub/cluster"
)

// ClusterService ...
type ClusterService struct {
	hub           *Hub
	deviceQuery   cluster.DeviceQuery
	deviceCommand cluster.DeviceCommand
}

// NewClusterService ...
func NewClusterService(hub *Hub, opts *ClusterOptions) *ClusterService {
	return &ClusterService{
		hub:           hub,
		deviceQuery:   opts.DeviceQuery,
		deviceCommand: opts.DeviceCommand,
	}
}

// AuthDevice ...
func (t *ClusterService) AuthDevice(model cluster.DeviceModel) bool {
	if model.ID() == "" {
		return false
	}

	if model.Architecture() == "" {
		return false
	}

	if model.Hostname() == "" {
		return false
	}

	if model.Platform() == "" {
		return false
	}

	return true
}

// AddOrUpdateDevice ...
func (t *ClusterService) AddOrUpdateDevice(model cluster.DeviceModel) {
	t.deviceCommand.AddOrUpdate(model)
}

// DeviceExists ...
func (t *ClusterService) DeviceExists(deviceid string) bool {
	return t.deviceQuery.Exists(deviceid)
}

// ShouldProxyRequest ...
func (t *ClusterService) ShouldProxyRequest(deviceid string) bool {
	return false
}

// ProxyRequest ...
func (t *ClusterService) ProxyRequest(deviceid string, rw http.ResponseWriter, r *http.Request) {

}

// IsDeviceConnected ...
func (t *ClusterService) IsDeviceConnected(deviceid string) bool {
	return t.deviceQuery.Connected(deviceid)
}
