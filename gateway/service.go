package gateway

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/deviceio/shared/types"

	"strings"

	"bytes"

	"github.com/palantir/stacktrace"
)

type serviceConnections struct {
	items map[string]*connection
	*sync.RWMutex
}

type Service struct {
	BindAddr    string
	TLSCertPath string
	TLSKeyPath  string
	conns       *serviceConnections
}

func (t *Service) Start() {
	t.init()

	logrus.WithFields(logrus.Fields{
		"bindAddr": t.BindAddr,
	}).Info("gateway starting")

	certpath := t.TLSCertPath
	keypath := t.TLSKeyPath

	if t.TLSCertPath == "" && t.TLSKeyPath == "" {
		certpath, keypath = t.makeTempCertificates()

		logrus.WithField("cert", certpath).Info("gateway temporary certificate")
		logrus.WithField("key", keypath).Info("gateway temporary key")
	}

	cer, err := tls.LoadX509KeyPair(certpath, keypath)

	if err != nil {
		logrus.Fatal("error loading gateway certificates", err.Error())
		return
	}

	ln, err := tls.Listen("tcp", t.BindAddr, &tls.Config{
		Certificates: []tls.Certificate{cer},
	})

	if err != nil {
		logrus.Fatal("error starting gateway tls listener", err.Error())
		return
	}

	defer ln.Close()
	defer os.Remove(certpath)
	defer os.Remove(keypath)

	for {
		conn, err := ln.Accept()
		if err != nil {
			logrus.Info("error accepting gateway connection", err.Error())
			continue
		}

		go t.handleConnection(conn)
	}
}

func (t *Service) ProxyHTTPRequest(deviceid string, path string, rw http.ResponseWriter, r *http.Request) error {
	if deviceid == "" {
		return stacktrace.NewError("deviceid is empty")
	}

	if rw == nil {
		return stacktrace.NewError("http.ResponseWriter is nil")
	}

	if r == nil {
		return stacktrace.NewError("http.Request is nil")
	}

	c, err := t.findConnectionForDevice(deviceid)

	if err != nil {
		return stacktrace.Propagate(err, "gateway failed to locate device")
	}

	err = c.proxyRequest(rw, r, path)

	if err != nil {
		return stacktrace.Propagate(err, "gateway failed to proxy on connection")
	}

	return nil
}

func (t *Service) init() {
	t.conns = &serviceConnections{
		items:   map[string]*connection{},
		RWMutex: &sync.RWMutex{},
	}
}

func (t *Service) handleConnection(conn net.Conn) {
	var gwconn *connection
	var err error

	if gwconn, err = newConnection(conn); err != nil {
		logrus.Error("Failed to create gateway connection:", err.Error())
		return
	}

	t.conns.Lock()
	defer t.conns.Unlock()

	id := strings.ToLower(gwconn.info.ID)
	hostname := strings.ToLower(gwconn.info.Hostname)

	if c, cok := t.conns.items[id]; cok {
		logrus.WithFields(logrus.Fields{
			"id": id,
			"connectedDeviceAddr":  c.conn.RemoteAddr().String(),
			"connectingDeviceAddr": gwconn.conn.RemoteAddr().String(),
		}).Error("device connections closed due to duplicate id")
		c.conn.Close()
		gwconn.conn.Close()
		return
	}

	if c, hok := t.conns.items[hostname]; hok {
		logrus.WithFields(logrus.Fields{
			"hostname":             hostname,
			"connectedDeviceAddr":  c.conn.RemoteAddr().String(),
			"connectingDeviceAddr": gwconn.conn.RemoteAddr().String(),
		}).Error("device connections closed due to duplicate hostname")
		c.conn.Close()
		gwconn.conn.Close()
		return
	}

	t.conns.items[id] = gwconn
	t.conns.items[hostname] = gwconn

	logrus.WithFields(logrus.Fields{
		"localAddr":    conn.LocalAddr(),
		"remoteAddr":   conn.RemoteAddr(),
		"id":           gwconn.info.ID,
		"hostname":     gwconn.info.Hostname,
		"platform":     gwconn.info.Platform,
		"architecture": gwconn.info.Architecture,
		"tags":         gwconn.info.Tags,
	}).Info("device connected")

	go t.closeloop(gwconn)
}

func (t *Service) makeTempCertificates() (string, string) {
	certgen := &types.CertGen{
		Host:      "localhost",
		ValidFrom: "Jan 1 15:04:05 2011",
		ValidFor:  8760 * time.Hour,
		IsCA:      false,
		RsaBits:   4096,
	}

	var err error
	var certBytes []byte
	var certTemp *os.File
	var keyBytes []byte
	var keyTemp *os.File

	certBytes, keyBytes = certgen.Generate()

	if certTemp, err = ioutil.TempFile("", "deviceio-hub"); err != nil {
		logrus.Fatal(err.Error())
	}
	defer certTemp.Close()

	if keyTemp, err = ioutil.TempFile("", "deviceio-hub"); err != nil {
		logrus.Fatal(err.Error())
	}
	defer keyTemp.Close()

	io.Copy(certTemp, bytes.NewBuffer(certBytes))
	io.Copy(keyTemp, bytes.NewBuffer(keyBytes))

	return certTemp.Name(), keyTemp.Name()
}

func (t *Service) findConnectionForDevice(deviceid string) (*connection, error) {
	if deviceid == "" {
		return nil, stacktrace.NewError("deviceid is empty")
	}

	t.conns.RLock()
	defer t.conns.RUnlock()

	deviceid = strings.ToLower(deviceid)

	c, cok := t.conns.items[deviceid]

	if cok {
		return c, nil
	}

	return nil, stacktrace.NewError("No such device found with id or hostname '%v'", deviceid)
}

func (t *Service) closeloop(c *connection) {
	for {
		if c.session.IsClosed() {
			logrus.WithFields(logrus.Fields{
				"localAddr":    c.conn.LocalAddr(),
				"remoteAddr":   c.conn.RemoteAddr(),
				"id":           c.info.ID,
				"hostname":     c.info.Hostname,
				"platform":     c.info.Platform,
				"architecture": c.info.Architecture,
				"tags":         c.info.Tags,
			}).Info("device disconnected")

			t.conns.Lock()
			defer t.conns.Unlock()

			id := strings.ToLower(c.info.ID)
			hostname := strings.ToLower(c.info.Hostname)

			if _, cok := t.conns.items[id]; cok {
				delete(t.conns.items, id)
			}

			if _, hok := t.conns.items[hostname]; hok {
				delete(t.conns.items, hostname)
			}

			return
		}
		time.Sleep(1 * time.Second)
	}
}
