package cluster

// DeviceCommand ...
type DeviceCommand interface {
	AddOrUpdate(DeviceModel)
}
