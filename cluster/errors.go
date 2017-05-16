package cluster

type AuthenticationFailed struct {
	Reason string
}

func (t *AuthenticationFailed) Error() string {
	return t.Reason
}
