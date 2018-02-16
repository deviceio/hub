package api

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/deviceio/hub/www"
	"github.com/deviceio/shared/types"
	"github.com/gorilla/mux"
)

type Service struct {
	BindAddr    string
	TLSCertPath string
	TLSKeyPath  string
	Controllers []Controller
}

func (t *Service) Start() {
	router := mux.NewRouter()

	for _, controller := range t.Controllers {
		controller.RegisterRoutes(router)
	}

	router.PathPrefix("/admin/").Handler(http.StripPrefix("/admin/", http.FileServer(www.EmbedHttpFS)))

	//server.Handle("/admin/", http.StripPrefix("/admin/", http.FileServer(www.EmbedFS)))
	//server.Handle("/", http.FileServer(www.EmbedFS))
	//server.Handle("/", t.auth(router))

	logrus.WithFields(logrus.Fields{
		"bindAddr":    t.BindAddr,
		"tlsCertPath": t.TLSCertPath,
		"tlsKeyPath":  t.TLSKeyPath,
	}).Info("api starting")

	certpath := t.TLSCertPath
	keypath := t.TLSKeyPath

	if t.TLSCertPath == "" && t.TLSKeyPath == "" {
		certpath, keypath = t.makeTempCertificates()

		logrus.WithField("cert", certpath).Info("api temporary certificate")
		logrus.WithField("key", keypath).Info("api temporary key")

		defer os.Remove(certpath)
		defer os.Remove(keypath)
	}

	if err := http.ListenAndServeTLS(
		t.BindAddr,
		certpath,
		keypath,
		router,
	); err != nil {
		logrus.Fatal(err.Error())
	}
}

func (t *Service) makeTempCertificates() (string, string) {
	certgen := &types.CertGen{
		Host:      "localhost",
		ValidFrom: "Jan 1 15:04:05 2011",
		ValidFor:  8760 * time.Hour,
		IsCA:      false,
		RsaBits:   4096,
	}

	var err error
	var certBytes []byte
	var certTemp *os.File
	var keyBytes []byte
	var keyTemp *os.File

	certBytes, keyBytes = certgen.Generate()

	if certTemp, err = ioutil.TempFile("", "deviceio-hub"); err != nil {
		logrus.Fatal(err.Error())
	}
	defer certTemp.Close()

	if keyTemp, err = ioutil.TempFile("", "deviceio-hub"); err != nil {
		logrus.Fatal(err.Error())
	}
	defer keyTemp.Close()

	io.Copy(certTemp, bytes.NewBuffer(certBytes))
	io.Copy(keyTemp, bytes.NewBuffer(keyBytes))

	return certTemp.Name(), keyTemp.Name()
}
