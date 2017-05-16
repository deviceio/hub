package cluster

type User struct {
	ID              string `gorethink:"id,omitempty"`
	Admin           bool   `gorethink:"admin,omitempty"`
	Login           string `gorethink:"login,omitempty"`
	Email           string `gorethink:"email,omitempty"`
	EmailFailedTOTP bool   `gorethink:"email_failed_totp,omitempty"`
	PasswordHash    []byte `gorethink:"password_hash,omitempty"`
	PasswordSalt    string `gorethink:"password_salt,omitempty"`
	TOTPSecret      string `gorethink:"totp_secret,omitempty"`
}
