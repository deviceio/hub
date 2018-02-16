package user

type Entity struct {
	ID         string
	Username   string
	HmacKey    string
	HmacSecret string
	PermitAddr []string
}
