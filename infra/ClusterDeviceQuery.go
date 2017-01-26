package infra

import (
	"github.com/deviceio/hub/infra/data"
	"github.com/deviceio/shared/logging"
)

// ClusterDeviceQuery ...
type ClusterDeviceQuery struct {
	logger logging.Logger
}

// NewClusterDeviceQuery ...
func NewClusterDeviceQuery(logger logging.Logger) *ClusterDeviceQuery {
	return &ClusterDeviceQuery{
		logger: logger,
	}
}

// DeviceExists ...
func (t *ClusterDeviceQuery) Exists(deviceid string) bool {
	resp, err := data.Table("Device").Filter(data.Filter{
		"id": deviceid,
	}).Count().Run(data.Session)

	if err != nil {
		t.logger.Error(err.Error())
		return false
	}

	var count int

	resp.One(&count)
	resp.Close()

	if count == 1 {
		return true
	}

	return false
}

// DeviceConnected ...
func (t *ClusterDeviceQuery) Connected(deviceid string) bool {
	resp, err := data.Table("Device").Filter(data.Filter{
		"id":           deviceid,
		"is_connected": true,
	}).Count().Run(data.Session)

	if err != nil {
		t.logger.Error(err.Error())
		return false
	}

	var count int

	resp.One(&count)
	resp.Close()

	if count == 1 {
		return true
	}

	return false
}
