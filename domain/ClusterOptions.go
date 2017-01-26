package domain

import (
	"quantum/hub/cluster"
	"quantum/shared/logging"
)

// ClusterOptions ...
type ClusterOptions struct {
	Logger      logging.Logger
	DeviceQuery cluster.DeviceQuery
	DeviceCommand cluster.DeviceCommand
}
