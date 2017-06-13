package gateway

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/url"

	"github.com/deviceio/shared/types"
	"github.com/google/uuid"
	"github.com/hashicorp/yamux"
	"github.com/palantir/stacktrace"
)

// connectionInfo contains the device identity and envrionment information supplied by
// the device during initial connection to the gateway server.
type connectionInfo struct {
	// ID as a V4 UUID in string format (hyphenated, uppercase)
	ID string

	// Hostname of the device.
	Hostname string

	// Architecture indicated by the device. amd64 i386 etc. usually the golang
	// values of GOARCH but no guarrenteed
	Architecture string

	//Platform indicated by the device. windows, linux, macosx etc. usually the
	// golang values of GOOS
	Platform string

	// Tags
	Tags []string
}

// connection represents the underlying connection of a device to the gateway server.
type connection struct {
	// info supplied to this connection. It is not the responsibility of the
	// connection to ascertain the validity of this data beyond its inherit structure.
	info *connectionInfo

	// conn represents the underlying net.Conn of this gateway connection
	conn net.Conn

	// session represents our multiplexed connection to the device. The actual
	// connection is a multiplex of streams, which we can treat as individual tcp
	// net.Listener's or net.Conn's.
	session *yamux.Session

	// httpclient is used to issue requests to the device tunnel http server.
	// this http.Client contains a custom transport to address the device's http
	// server over the session multiplexer
	httpclient *http.Client

	// httpproxy is used to proxy requests to the device tunnel http server.
	// this httputil.ReverseProxy contains a custom transport to address the device's http
	// server over the session multiplexer
	httpproxy *types.HttpStreamProxy
}

// newConnection instantiates a new instance of the connection type
func newConnection(conn net.Conn) (*connection, error) {
	client, err := yamux.Client(conn, nil)

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create mux client")
	}

	gc := &connection{
		conn:    conn,
		session: client,
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

	target, err := url.Parse("http://localhost/")

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to parse proxy url")
	}

	gc.httpproxy = types.NewSingleHostReverseProxy(target)
	gc.httpproxy.Transport = &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return client.Open()
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	gc.httpproxy.BufferPool = &bufpool{
		size: 250000,
	}

	resp, err := gc.httpclient.Get("http://localhost/info")

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed retrieving device info")
	}

	err = json.NewDecoder(resp.Body).Decode(&gc.info)

	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to decode device info")
	}

	if _, err := uuid.Parse(gc.info.ID); err != nil {
		return nil, stacktrace.Propagate(err, "agent id is not a valid UUID")
	}

	return gc, nil
}

// proxyRequest takes a http request originating elsewhere and proxies the request
// to the device's local http server over a new multiplexed stream. This function
// is responsible to mutate the request before sending adding or removing information
// as necessary making ready for device consumption.
func (t *connection) proxyRequest(w http.ResponseWriter, r *http.Request, path string) error {
	r.URL.Path = "/" + path
	t.httpproxy.ServeHTTP(w, r)

	return nil
}
