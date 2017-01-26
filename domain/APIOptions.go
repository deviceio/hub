package domain

import "quantum/shared/logging"

// APIOptions ...
type APIOptions struct {
	BindAddr    string
	Logger      logging.Logger
	TLSCertPath string
	TLSKeyPath  string
}
