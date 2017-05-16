package gateway

import (
	"crypto/tls"
	"fmt"
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
)

type Service struct {
	BindAddr    string
	TLSCertPath string
	TLSKeyPath  string
	conns       map[string]*connection
	hosts       map[string][]*connection
	mutex       *sync.Mutex
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
	c, err := t.findConnectionForDevice(deviceid)

	if err != nil {
		return err
	}

	err = c.proxyRequest(rw, r, path)

	if err != nil {
		return err
	}

	return nil
}

func (t *Service) init() {
	t.conns = map[string]*connection{}
	t.hosts = map[string][]*connection{}
	t.mutex = &sync.Mutex{}
}

func (t *Service) handleConnection(conn net.Conn) {
	var gwconn *connection
	var err error
	var hostconns []*connection
	var ok bool

	if gwconn, err = newConnection(conn); err != nil {
		logrus.Error("Failed to create gateway connection:", err.Error())
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.conns[strings.ToLower(gwconn.info.ID)] = gwconn

	if hostconns, ok = t.hosts[strings.ToLower(gwconn.info.Hostname)]; !ok {
		t.hosts[strings.ToLower(gwconn.info.Hostname)] = []*connection{gwconn}
	} else {
		hostconns = append(hostconns, gwconn)
	}

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
			return nil, &ErrAmbiguousHostnameLookup{
				Message: "Ambiguous hostname lookup. Two or more devices connected with the same hostname",
			}
		}
	}

	return nil, &ErrGatewayDeviceDoesNotExist{
		DeviceID: deviceid,
		Message:  fmt.Sprintf("No such device found with id or hostname '%v'", deviceid),
	}
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
