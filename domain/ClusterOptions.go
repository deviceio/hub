package domain

import (
	"github.com/deviceio/hub/cluster"
	"github.com/deviceio/shared/logging"
)

// ClusterOptions ...
type ClusterOptions struct {
	Logger        logging.Logger
	DeviceQuery   cluster.DeviceQuery
	DeviceCommand cluster.DeviceCommand
}
