package user

type Repository interface {
	GetByHmacKey(key string) (user *Entity, err error)
	ExistsByHmacKey(key string) (exists bool, err error)
}
