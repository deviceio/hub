package domain

import "github.com/deviceio/shared/logging"

// GatewayOptions ...
type GatewayOptions struct {
	BindAddr    string
	TLSCertPath string
	TLSKeyPath  string
	Logger      logging.Logger
}
