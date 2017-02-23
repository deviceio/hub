package domain

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/deviceio/shared/logging"
	"github.com/hashicorp/yamux"

	"net/http/httputil"

	"github.com/gorilla/websocket"
)

// GatewayConnection represents the underlying connection of a connected device to
// the gateway component
type GatewayConnection struct {
	// gwservice used to interact with the parent GatewayService domain-level api
	gwservice *GatewayService

	// logger provides logging for this type
	logger logging.Logger

	// wsconn the underlying websocket connection this GatewayConnection represents
	wsconn *websocket.Conn

	session *yamux.Session

	httpclient *http.Client
}

// NewGatewayConnection instantiates a new instance of the GatewayConnection type
func NewGatewayConnection(wsconn *websocket.Conn, logger logging.Logger, gateway *GatewayService) *GatewayConnection {
	client, _ := yamux.Client(wsconn.UnderlyingConn(), nil)

	gc := &GatewayConnection{
		gwservice: gateway,
		logger:    logger,
		wsconn:    wsconn,
		session:   client,
	}

	gc.httpclient = &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return client.Open()
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	go gc.closeloop()

	return gc
}

func (t *GatewayConnection) GetConfig() (map[string]interface{}, error) {
	resp, err := t.httpclient.Get("http://localhost/config")

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var config map[string]interface{}

	err = json.Unmarshal(body, &config)

	if err != nil {
		return nil, err
	}

	return config, nil
}

func (t *GatewayConnection) ProxyRequest(w http.ResponseWriter, r *http.Request, path string) {
	proxy := &httputil.ReverseProxy{
		Transport: t.httpclient.Transport,
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Path = "/" + path
			r.URL.Host = "localhost"
		},
	}

	proxy.ServeHTTP(w, r)
}

func (t *GatewayConnection) closeloop() {
	for {
		if t.session.IsClosed() {
			log.Println("Agent Connection Closed")
			return
		}
		time.Sleep(1 * time.Second)
	}
}
