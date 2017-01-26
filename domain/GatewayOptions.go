package domain

import "quantum/shared/logging"

// GatewayOptions ...
type GatewayOptions struct {
	BindAddr    string
	TLSCertPath string
	TLSKeyPath  string
	Logger      logging.Logger
}
