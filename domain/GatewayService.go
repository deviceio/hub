package domain

import (
	"net/http"
	"quantum/shared/logging"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// GatewayService is responsible for managing the upgrade process and lifecylce
// of agent transport connections. GatewayService also provides a domain-level
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

	// wsupgrader is used to upgrade incoming websocket upgrade requests
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
	router.HandleFunc("/v1/status", t.httpGetV1Status).Methods("GET")

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
// GatewayService by device ID (AgentID) and returns the connection for use.
func (t *GatewayService) FindConnectionForDevice(deviceid string) *GatewayConnection {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.conns[deviceid]
}

// addConnection adds a GatewayConnection to this GatewayService's list of active
// connections. This is a locking call.
func (t *GatewayService) addConnection(c *GatewayConnection) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.conns[c.handshake.AgentID] = c
}

// delConnection removes a GatewayConnection from this GatewayService's list of active
// connections. This is a locking call.
func (t *GatewayService) delConnection(c *GatewayConnection) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	delete(t.conns, c.handshake.AgentID)
}

// httpGetV1Status -> GET /v1/status
func (t *GatewayService) httpGetV1Status(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("OK"))
}

// httpGetV1Connect -> GET /v1/connect
func (t *GatewayService) httpGetV1Connect(resp http.ResponseWriter, req *http.Request) {
	conn, err := t.wsupgrader.Upgrade(resp, req, nil)

	if err != nil {
		t.logger.Error(err.Error())
		resp.WriteHeader(400)
		resp.Write([]byte(""))
		return
	}

	t.logger.Debug(
		"New Agent Connection: LocalAddr=%v RemoteAddr=%v",
		conn.LocalAddr(),
		conn.RemoteAddr(),
	)

	c := NewGatewayConnection(conn, &logging.DefaultLogger{}, t)
	c.start()
}
