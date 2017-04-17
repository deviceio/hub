package domain

// UserRepository provides CRUD methods against the domain User Entity
type UserRepository interface {
	// GetByHmacKey returns a user by the supplied Hmac Key
	GetByHmacKey(key string) (user *User, err error)

	// ExistsByHmacKey indicates if a user exists by an Hmac Key
	ExistsByHmacKey(key string) (exists bool, err error)
}
