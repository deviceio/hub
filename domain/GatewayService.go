package domain

import (
	"net/http"
	"sync"
	"time"

	"github.com/deviceio/shared/logging"

	"strings"

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
	hosts map[string]*GatewayConnection

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
		hosts:  map[string]*GatewayConnection{},
		hub:    hub,
		logger: opts.Logger,
		mutex:  &sync.Mutex{},
		opts:   opts,
	}
}

// Start sets up the GatewayService and starts the http websocket listener for
// incoming agent connections. This is a blocking call.
func (t *GatewayService) Start() {
	var err error

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

	err = http.ListenAndServeTLS(
		t.opts.BindAddr,
		t.opts.TLSCertPath,
		t.opts.TLSKeyPath,
		server,
	)

	if err != nil {
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
		return h, nil
	}

	return nil, ErrGatewayDeviceDoesNotExist
}

// httpGetV1Connect is the gateway http endpoint that accepts new device connections.
// this method is responsible for accepting the connection, adjuticating it with the cluster
// and communicating with other domain components to ensure it is properly tracked and updated
// in the database.
func (t *GatewayService) httpGetV1Connect(resp http.ResponseWriter, req *http.Request) {
	conn, err := t.wsupgrader.Upgrade(resp, req, nil)

	if err != nil {
		t.logger.Error(err.Error())
		resp.WriteHeader(400)
		resp.Write([]byte(""))
		t.logger.Error("Device connection attempt failed websocket upgrade:", err.Error())
		return
	}

	c, err := NewGatewayConnection(conn, &logging.DefaultLogger{}, t)

	if err != nil {
		t.logger.Error("Failed to create gateway connection:", err.Error())
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.conns[strings.ToLower(c.info.ID)] = c
	t.hosts[strings.ToLower(c.info.Hostname)] = c

	go t.closeloop(c)

	t.logger.Debug(
		"Device Connected: LocalAddr=%v RemoteAddr=%v ID=%v Hostname=%v Platform=%v Arch=%v Tags=%v",
		conn.LocalAddr(),
		conn.RemoteAddr(),
		c.info.ID,
		c.info.Hostname,
		c.info.Platform,
		c.info.Architecture,
		c.info.Tags,
	)
}

// closeloop watches for a break in the device connection
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

			delete(t.conns, strings.ToLower(c.info.ID))
			delete(t.hosts, strings.ToLower(c.info.Hostname))

			return
		}
		time.Sleep(1 * time.Second)
	}
}
