package domain

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/deviceio/shared/logging"
	"github.com/hashicorp/yamux"

	"github.com/gorilla/websocket"
)

// GatewayConnection represents the underlying connection of a connected device to
// the gateway component
type GatewayConnection struct {
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

// Info returns key information about the device that the hub requires to track
// and address the device throughout the cluster and the API. This information
// is assumed to fit the prescribed model.
func (t *GatewayConnection) Info() (*DeviceInfoModel, error) {
	resp, err := t.httpclient.Get("http://localhost/info")

	if err != nil {
		t.logger.Error("Error retrieving device info:", err.Error())
		return nil, err
	}

	var config *DeviceInfoModel

	err = json.NewDecoder(resp.Body).Decode(&config)

	if err != nil {
		t.logger.Error("Error decoding device info:", err.Error())
		return nil, err
	}

	return config, nil
}

// ProxyRequest takes a http request originating elsewhere within the domain and
// proxies the request to the device's local http server over a new multiplexed
// stream. This function is responsible to mutate the request before sending adding
// or removing information as necessary making ready for device consumption.
func (t *GatewayConnection) ProxyRequest(w http.ResponseWriter, r *http.Request, path string) {
	flusher, _ := w.(http.Flusher)

	r.RequestURI = ""
	r.URL.Scheme = "http"
	r.URL.Path = "/" + path
	r.URL.Host = "localhost"

	resp, err := t.httpclient.Do(r)

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	if resp.Header.Get("Transport-Encoding") != "chunked" {
		io.Copy(w, resp.Body)
		return
	}

	if resp.Header.Get("Transport-Encoding") == "chunked" {
		buf := make([]byte, 250000)
		for {
			i, err := resp.Body.Read(buf)

			if err == io.EOF {
				break
			}

			w.Write(buf[:i])
			flusher.Flush()
		}
		return
	}
}

// closeloop watches for a break in the device connection
func (t *GatewayConnection) closeloop() {
	for {
		if t.session.IsClosed() {
			t.logger.Error("Agent Connection Closed")
			return
		}
		time.Sleep(1 * time.Second)
	}
}
