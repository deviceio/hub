package domain

type User struct {
	ID         string
	Username   string
	HmacKey    string
	HmacSecret string
	PermitAddr []string
}
