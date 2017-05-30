package types

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"
)

// listenAndServeTLSKeyPair start a server using in-memory TLS KeyPair
func ListenAndServeTLSKeyPair(addr string, cert tls.Certificate,
	handler http.Handler) error {

	if addr == "" {
		return errors.New("Invalid address string")
	}

	server := &http.Server{Addr: addr, Handler: handler}

	config := &tls.Config{}
	config.NextProtos = []string{"http/1.1"}
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0] = cert

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)},
		config)

	return server.Serve(tlsListener)
}

// From https://golang.org/src/net/http/server.go
// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
