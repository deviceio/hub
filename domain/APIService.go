package domain

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/deviceio/hub/www"
	"github.com/gorilla/mux"
)

// APIService ...
type APIService struct {
	hub  *Hub
	opts *APIOptions
}

// NewAPIService ...
func NewAPIService(hub *Hub, opts *APIOptions) *APIService {
	return &APIService{
		hub:  hub,
		opts: opts,
	}
}

// Start ...
func (t *APIService) Start() {
	var err error

	//server := http.NewServeMux()
	router := mux.NewRouter()

	router.HandleFunc("/status", t.status).Methods("GET")
	router.PathPrefix("/admin/").Handler(http.StripPrefix("/admin/", http.FileServer(www.EmbedHttpFS)))

	//server.Handle("/admin/", http.StripPrefix("/admin/", http.FileServer(www.EmbedFS)))
	//server.Handle("/", http.FileServer(www.EmbedFS))
	//server.Handle("/", t.auth(router))

	router.HandleFunc("/device/{deviceid}/{path:[0-9a-zA-Z\\/]+}", t.httpProxyDevice)

	err = http.ListenAndServeTLS(
		t.opts.BindAddr,
		t.opts.TLSCertPath,
		t.opts.TLSKeyPath,
		router,
	)

	if err != nil {
		log.Fatal(err, string(debug.Stack()))
	}
}

// auth ...
func (t *APIService) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// httpGetStatus GET /v1/status
func (t *APIService) status(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("OK"))
}

// httpProxyDevice
func (t *APIService) httpProxyDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	c, err := t.hub.Gateway.FindConnectionForDevice(vars["deviceid"])

	if err != nil && err == ErrGatewayDeviceDoesNotExist {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	c.ProxyRequest(w, r, vars["path"])
}
