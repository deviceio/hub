package infra

import (
	"github.com/deviceio/hub/cluster"
	"github.com/deviceio/hub/infra/data"
	"github.com/deviceio/shared/logging"
)

// ClusterDeviceCommand ...
type ClusterDeviceCommand struct {
	logger logging.Logger
}

// NewClusterDeviceCommand ...
func NewClusterDeviceCommand(logger logging.Logger) *ClusterDeviceCommand {
	return &ClusterDeviceCommand{
		logger: logger,
	}
}

// AddOrUpdate ...
func (t *ClusterDeviceCommand) AddOrUpdate(device cluster.DeviceModel) {
	table := data.Table("Device")

	resp, err := table.Filter(data.Filter{
		"id": device.ID(),
	}).Count().Run(data.Session)

	if err != nil {
		t.logger.Error(err.Error())
		return
	}

	var count int

	resp.One(&count)
	resp.Close()

	doc := data.Document{
		"id":           device.ID(),
		"hostname":     device.Hostname(),
		"architecture": device.Architecture(),
		"platform":     device.Platform(),
		"tags":         device.Tags(),
		"is_connected": device.IsConnected(),
	}

	if count == 1 {
		table.Get(doc["id"]).Update(doc).RunWrite(data.Session)
	} else if count == 0 {
		table.Insert(doc).RunWrite(data.Session)
	} else {
		t.logger.Error("0 or 1 devices expected, recieved %v", count)
	}
}
