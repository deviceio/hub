package domain

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"

	"github.com/deviceio/shared/logging"
	"github.com/hashicorp/yamux"

	"net/http/httputil"

	"github.com/gorilla/websocket"
)

// GatewayConnection represents the underlying connection of a connected device to
// the gateway component
type GatewayConnection struct {
	// info contains information about the connecting device
	info struct {
		ID           string
		Hostname     string
		Architecture string
		Platform     string
		IsConnected  bool
		Tags         []string
	}

	// gwservice used to interact with the parent GatewayService domain-level api
	gwservice *GatewayService

	// logger provides logging for this type
	logger logging.Logger

	// wsconn the underlying websocket connection this GatewayConnection represents.
	// this connection is only used durring the connection phase, as we take over
	// the underlying tls tcp connection for use with multiplexing of multiple tcp
	// streams.
	wsconn *websocket.Conn

	// session represents our multiplexed connection to the device. The actual
	// connection is a multiplex of streams, which we can treat as individual tcp
	// net.Listener's or net.Conn's.
	session *yamux.Session

	// httpclient is used to issue requests to the device's local http server.
	// this http.Client contains a custom transport to address the device's http
	// server over the session multiplexer
	httpclient *http.Client

	// httpproxy is used to proxy requests to the device's local http server.
	// this httputil.ReverseProxy contains a custom transport to address the device's http
	// server over the session multiplexer
	httpproxy *httputil.ReverseProxy
}

// NewGatewayConnection instantiates a new instance of the GatewayConnection type
func NewGatewayConnection(wsconn *websocket.Conn, logger logging.Logger, gateway *GatewayService) (*GatewayConnection, error) {
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

	gc.httpproxy = &httputil.ReverseProxy{
		Director: func(r *http.Request) {
		},
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return client.Open()
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := gc.httpclient.Get("http://localhost/info")

	if err != nil {
		gc.logger.Error("Error retrieving device info:", err.Error())
		return nil, err
	}

	err = json.NewDecoder(resp.Body).Decode(&gc.info)

	if err != nil {
		gc.logger.Error("Error decoding device info:", err.Error())
		return nil, err
	}

	return gc, nil
}

// ProxyRequest takes a http request originating elsewhere within the domain and
// proxies the request to the device's local http server over a new multiplexed
// stream. This function is responsible to mutate the request before sending adding
// or removing information as necessary making ready for device consumption.
func (t *GatewayConnection) ProxyRequest(w http.ResponseWriter, r *http.Request, path string) error {
	r.RequestURI = ""
	r.URL.Scheme = "http"
	r.URL.Path = "/" + path
	r.URL.Host = "localhost"

	/*resp, err := t.httpclient.Do(r)

	if err != nil {
		return err
	}

	t.copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	buf := make([]byte, 250000)

	if _, err = io.CopyBuffer(w, resp.Body, buf); err != nil {
		if err != io.EOF {
			return err
		}
	}*/

	t.httpproxy.ServeHTTP(w, r)

	return nil
}
