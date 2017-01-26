package cluster

// DeviceQuery ...
type DeviceQuery interface {
	Exists(string) bool
	Connected(string) bool
}
