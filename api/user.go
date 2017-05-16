package api

import (
	"github.com/deviceio/hub/user"
	"github.com/gorilla/mux"
)

type UserController struct {
	UserService *user.Service
}

func (t *UserController) RegisterRoutes(router *mux.Router) {
}
