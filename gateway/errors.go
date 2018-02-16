package gateway

type ErrAmbiguousHostnameLookup struct {
	Message string
}

func (t *ErrAmbiguousHostnameLookup) Error() string {
	return t.Message
}

type ErrGatewayDeviceDoesNotExist struct {
	DeviceID string
	Message  string
}

func (t *ErrGatewayDeviceDoesNotExist) Error() string {
	return t.Message
}
