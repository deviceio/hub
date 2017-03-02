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

// AddOrUpdateDevice ...
func (t *ClusterService) AddOrUpdateDevice(model *DeviceInfoModel) {
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
