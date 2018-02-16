package cluster

type User struct {
	ID               string `gorethink:"id,omitempty"`
	Admin            bool   `gorethink:"admin,omitempty"`
	Login            string `gorethink:"login,omitempty"`
	Email            string `gorethink:"email,omitempty"`
	PasswordHash     []byte `gorethink:"password_hash,omitempty"`
	PasswordSalt     string `gorethink:"password_salt,omitempty"`
	TOTPSecret       []byte `gorethink:"totp_secret,omitempty"`
	ED25519PublicKey []byte `gorethink:"ed22519_public_key,omitempty"`
}
