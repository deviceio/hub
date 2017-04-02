package domain

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/deviceio/shared/logging"
	"github.com/deviceio/shared/types"

	"strings"

	"bytes"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// GatewayService is responsible for managing the upgrade process and lifecylce
// of device transport connections. GatewayService also provides a domain-level
// api to locate and retrieve active connections associated with the GatewayService
// component
type GatewayService struct {
	// conns provides a map of gateway connections indexed by Device ID
	conns map[string]*GatewayConnection

	// hosts provides a map of gateway connections indexed by Device Hostname
	hosts map[string][]*GatewayConnection

	// hub provides access the the top-level hub root aggregate of the domain
	hub *Hub

	// logger provides logging for this type
	logger logging.Logger

	// mutex provides syncronization for this type
	mutex *sync.Mutex

	// opts various GatewayService options
	opts *GatewayOptions

	// wsupgrader is used to upgrade incoming websocket connections
	wsupgrader *websocket.Upgrader
}

// NewGatewayService creates a new instance of the GatewayService type
func NewGatewayService(hub *Hub, opts *GatewayOptions) *GatewayService {
	return &GatewayService{
		conns:  map[string]*GatewayConnection{},
		hosts:  map[string][]*GatewayConnection{},
		hub:    hub,
		logger: opts.Logger,
		mutex:  &sync.Mutex{},
		opts:   opts,
	}
}

// Start sets up the GatewayService and starts the http websocket listener for
// incoming agent connections. This is a blocking call.
func (t *GatewayService) Start() {
	var router = mux.NewRouter()
	var server = http.NewServeMux()

	t.wsupgrader = &websocket.Upgrader{
		HandshakeTimeout:  time.Duration(time.Second * 30),
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	router.HandleFunc("/v1/connect", t.httpGetV1Connect).Methods("GET")

	server.Handle("/", router)

	t.logger.Debug("Gateway Starting on %v", t.opts.BindAddr)

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

		t.logger.Debug("Gateway Temp cert %v", certpath)
		t.logger.Debug("Gateway Temp key %v", keypath)
	}

	if err := http.ListenAndServeTLS(
		t.opts.BindAddr,
		certpath,
		keypath,
		server,
	); err != nil {
		t.logger.Error(err.Error())
	}
}

// FindConnectionForDevice locates a GatewayConnection currently connected to this
// GatewayService by device ID and returns the connection for use.
func (t *GatewayService) FindConnectionForDevice(deviceid string) (*GatewayConnection, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	deviceid = strings.ToLower(deviceid)

	c, cok := t.conns[deviceid]
	h, hok := t.hosts[deviceid]

	if cok {
		return c, nil
	}

	if hok {
		if len(h) == 1 {
			return h[0], nil
		} else if len(h) > 1 {
			return nil, fmt.Errorf("Ambiguous hostname lookup. Two or more devices connected with the same hostname")
		} else {
			return nil, fmt.Errorf("No device connections indexed by this hostname")
		}
	}

	return nil, ErrGatewayDeviceDoesNotExist
}

// httpGetV1Connect is the gateway http endpoint that accepts new device connections.
// this method is responsible for accepting the connection, adjuticating it with the cluster
// and communicating with other domain components to ensure it is properly tracked and updated
// in the database.
func (t *GatewayService) httpGetV1Connect(resp http.ResponseWriter, req *http.Request) {
	var err error
	var ok bool
	var wsconn *websocket.Conn
	var gwconn *GatewayConnection
	var hostconns []*GatewayConnection

	if wsconn, err = t.wsupgrader.Upgrade(resp, req, nil); err != nil {
		t.logger.Error(err.Error())
		resp.WriteHeader(400)
		resp.Write([]byte(""))
		t.logger.Error("Device connection attempt failed websocket upgrade:", err.Error())
		return
	}

	if gwconn, err = NewGatewayConnection(wsconn, &logging.DefaultLogger{}, t); err != nil {
		t.logger.Error("Failed to create gateway connection:", err.Error())
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.conns[strings.ToLower(gwconn.info.ID)] = gwconn

	if hostconns, ok = t.hosts[strings.ToLower(gwconn.info.Hostname)]; !ok {
		t.hosts[strings.ToLower(gwconn.info.Hostname)] = []*GatewayConnection{gwconn}
	} else {
		hostconns = append(hostconns, gwconn)
	}

	go t.closeloop(gwconn)

	t.logger.Debug(
		"Device Connected: LocalAddr=%v RemoteAddr=%v ID=%v Hostname=%v Platform=%v Arch=%v Tags=%v",
		wsconn.LocalAddr(),
		wsconn.RemoteAddr(),
		gwconn.info.ID,
		gwconn.info.Hostname,
		gwconn.info.Platform,
		gwconn.info.Architecture,
		gwconn.info.Tags,
	)
}

// closeloop watches for a break in the device connection and cleans up the connection
// from the gateway service
func (t *GatewayService) closeloop(c *GatewayConnection) {
	for {
		if c.session.IsClosed() {
			t.logger.Debug(
				"Device Disconnected: LocalAddr=%v RemoteAddr=%v ID=%v Hostname=%v Platform=%v Arch=%v Tags=%v",
				c.wsconn.LocalAddr(),
				c.wsconn.RemoteAddr(),
				c.info.ID,
				c.info.Hostname,
				c.info.Platform,
				c.info.Architecture,
				c.info.Tags,
			)

			t.mutex.Lock()
			defer t.mutex.Unlock()

			id := strings.ToLower(c.info.ID)
			hostname := strings.ToLower(c.info.Hostname)

			delete(t.conns, id)

			for a, host := range t.hosts[hostname] {
				if host == c {
					t.hosts[hostname] = append(t.hosts[hostname][:a], t.hosts[hostname][a+1:]...)
					break
				}
			}

			return
		}
		time.Sleep(1 * time.Second)
	}
}
