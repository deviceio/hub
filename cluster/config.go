package cluster

import "net/http"

type Config struct {
	BindAddr             string
	TLSCertPath          string
	TLSKeyPath           string
	LocalDeviceProxyFunc func(deviceid string, path string, rw http.ResponseWriter, r *http.Request) error
}
