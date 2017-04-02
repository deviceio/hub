package domain

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/deviceio/hub/www"
	"github.com/deviceio/shared/logging"
	"github.com/deviceio/shared/types"
	"github.com/gorilla/mux"
)

// APIService ...
type APIService struct {
	hub    *Hub
	opts   *APIOptions
	logger logging.Logger
}

// NewAPIService ...
func NewAPIService(hub *Hub, opts *APIOptions) *APIService {
	return &APIService{
		hub:    hub,
		opts:   opts,
		logger: opts.Logger,
	}
}

// Start ...
func (t *APIService) Start() {
	//server := http.NewServeMux()
	router := mux.NewRouter()

	router.HandleFunc("/status", t.status).Methods("GET")
	router.PathPrefix("/admin/").Handler(http.StripPrefix("/admin/", http.FileServer(www.EmbedHttpFS)))

	//server.Handle("/admin/", http.StripPrefix("/admin/", http.FileServer(www.EmbedFS)))
	//server.Handle("/", http.FileServer(www.EmbedFS))
	//server.Handle("/", t.auth(router))

	router.HandleFunc("/device/{deviceid}/{path:[0-9a-zA-Z\\/]+}", t.httpProxyDevice)

	t.logger.Debug("API Starting on %v", t.opts.BindAddr)

	certpath := t.opts.TLSCertPath
	keypath := t.opts.TLSKeyPath

	if t.opts.TLSCertPath == "" && t.opts.TLSKeyPath == "" {
		certgen := &types.CertGen{
			Host:      "localhost",
			ValidFrom: "Jan 1 15:04:05 2011",
			ValidFor:  867240 * time.Hour,
			IsCA:      true,
			RsaBits:   4096,
		}

		var err error
		var certBytes []byte
		var certTemp *os.File
		var keyBytes []byte
		var keyTemp *os.File

		certBytes, keyBytes = certgen.Generate()

		if certTemp, err = ioutil.TempFile("", "deviceio-hub"); err != nil {
			t.logger.Fatal(err.Error())
		}
		defer certTemp.Close()

		if keyTemp, err = ioutil.TempFile("", "deviceio-hub"); err != nil {
			t.logger.Fatal(err.Error())
		}
		defer keyTemp.Close()

		io.Copy(certTemp, bytes.NewBuffer(certBytes))
		io.Copy(keyTemp, bytes.NewBuffer(keyBytes))

		certpath = certTemp.Name()
		keypath = keyTemp.Name()

		defer os.Remove(certpath)
		defer os.Remove(keypath)

		t.logger.Debug("Api Temp cert %v", certpath)
		t.logger.Debug("Api Temp key %v", keypath)
	}

	if err := http.ListenAndServeTLS(
		t.opts.BindAddr,
		certpath,
		keypath,
		router,
	); err != nil {
		t.logger.Error(err.Error())
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

	err = c.ProxyRequest(w, r, vars["path"])

	if err != nil {
		t.opts.Logger.Error(err.Error())
		return
	}
}
