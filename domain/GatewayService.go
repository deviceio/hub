package domain

import (
	"net/http"
	"sync"
	"time"

	"github.com/deviceio/shared/logging"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// GatewayService is responsible for managing the upgrade process and lifecylce
// of device transport connections. GatewayService also provides a domain-level
// api to locate and retrieve active connections associated with the GatewayService
// component
type GatewayService struct {
	// conns provides a map of gateway connections indexed by AgentID
	conns map[string]*GatewayConnection

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
		conns:  make(map[string]*GatewayConnection),
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

	c, ok := t.conns[deviceid]

	if !ok {
		return nil, ErrGatewayDeviceDoesNotExist
	}

	return c, nil
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

	c := NewGatewayConnection(conn, &logging.DefaultLogger{}, t)

	info, err := c.Info()

	if err != nil {
		t.logger.Error("Device failed to provide info:", err.Error())
		c.wsconn.Close()
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.conns[info.ID] = c

	t.logger.Debug(
		"New Device Connection: LocalAddr=%v RemoteAddr=%v Info=%v",
		conn.LocalAddr(),
		conn.RemoteAddr(),
		info,
	)
}
