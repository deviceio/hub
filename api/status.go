package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

type StatusController struct {
}

func (t *StatusController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/v1/status", t.getV1Status).Methods("GET")
}

func (t *StatusController) getV1Status(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("OK"))
}
