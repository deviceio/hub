package domain

import "github.com/deviceio/shared/logging"

// APIOptions ...
type APIOptions struct {
	BindAddr    string
	Logger      logging.Logger
	TLSCertPath string
	TLSKeyPath  string
}
