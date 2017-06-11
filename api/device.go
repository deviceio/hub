package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/deviceio/hub/cluster"
	"github.com/gorilla/mux"
)

type DeviceController struct {
	ClusterService cluster.Service
}

func (t *DeviceController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/device", t.httpGetDevices)
	router.HandleFunc("/device/", t.httpGetDevices)
	router.HandleFunc("/device/{deviceid}", t.httpProxyDevice)
	router.HandleFunc("/device/{deviceid}/", t.httpProxyDevice)
	router.HandleFunc("/device/{deviceid}/{path:.*}", t.httpProxyDevice)
}

func (t *DeviceController) httpGetDevices(rw http.ResponseWriter, r *http.Request) {

}

func (t *DeviceController) httpProxyDevice(rw http.ResponseWriter, r *http.Request) {
	var err error

	if err = t.ClusterService.AuthenticateAPIRequest(r); err != nil {
		rw.WriteHeader(http.StatusForbidden)
		rw.Write([]byte(""))

		logrus.WithFields(logrus.Fields{
			"remoteAddr":    r.RemoteAddr,
			"authorization": r.Header.Get("Authorization"),
		}).Error(err.Error())

		return
	}

	vars := mux.Vars(r)

	r.Header.Add(
		"X-Deviceio-Parent-Path",
		fmt.Sprintf("/device/%v", vars["deviceid"]),
	)

	logrus.WithFields(logrus.Fields{
		"remoteAddr": r.RemoteAddr,
		"user": strings.Split(
			strings.Split(r.Header.Get("Authorization"), " ")[1],
			":",
		)[0],
		"deviceId":       vars["deviceid"],
		"deviceEndpoint": vars["path"],
	}).Info("device access")

	t.ClusterService.ProxyDeviceRequest(
		vars["deviceid"],
		vars["path"],
		rw,
		r,
	)
}
