package cluster

// DeviceModel ...
type DeviceModel interface {
	ID() string
	Hostname() string
	Platform() string
	Architecture() string
	Tags() []string
	IsConnected() bool
}
