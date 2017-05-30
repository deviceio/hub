package cluster

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	linq "github.com/ahmetb/go-linq"
	"github.com/deviceio/hub/db"
	"github.com/deviceio/shared/types"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/ed25519"
)

type Service interface {
	AuthenticateAPIRequest(r *http.Request) (failure error)
	ProxyDeviceRequest(deviceid string, path string, rw http.ResponseWriter, r *http.Request)
	Start()
}

func NewService(config *Config) Service {
	return &service{
		config: config,
	}
}

type service struct {
	config        *Config
	userCache     map[string]*User
	userCacheMu   *sync.Mutex
	memberCache   map[string]*Member
	memberCacheMu *sync.Mutex
	deviceCache   map[string]*Device
	deviceCacheMu *sync.Mutex
}

func (t *service) AuthenticateAPIRequest(r *http.Request) error {
	authheader := r.Header.Get("Authorization")

	if authheader == "" {
		return &AuthenticationFailed{
			Reason: "authentication header empty",
		}
	}

	authHeaderTypeAndValue := strings.Split(strings.TrimSpace(authheader), " ")

	if len(authHeaderTypeAndValue) != 2 {
		return &AuthenticationFailed{
			Reason: "authorization header does not contain valid type and value",
		}
	}

	if authHeaderTypeAndValue[0] != "DEVICEIO-HUB-AUTH" {
		return &AuthenticationFailed{
			Reason: "authorization header <type> must be 'DEVICEIO-HUB-AUTH'",
		}
	}

	authHeaderValues := strings.Split(authHeaderTypeAndValue[1], ":")

	if len(authHeaderValues) != 2 {
		return &AuthenticationFailed{
			Reason: "authorization value does not have required format <user_id>:<ed25519_signature_base64>",
		}
	}

	suppliedID := authHeaderValues[0]
	suppliedSignatrue, err := base64.StdEncoding.DecodeString(authHeaderValues[1])

	if err != nil {
		return &AuthenticationFailed{
			Reason: err.Error(),
		}
	}

	t.userCacheMu.Lock()
	var user *User
	for _, a := range t.userCache {
		if a.ID == suppliedID || a.Login == suppliedID || a.Email == suppliedID {
			user = a
			break
		}
	}
	t.userCacheMu.Unlock()

	if user == nil {
		return &AuthenticationFailed{
			Reason: "no such user",
		}
	}

	passcode, err := totp.GenerateCode(string(user.TOTPSecret), time.Now())

	if err != nil {
		return &AuthenticationFailed{
			Reason: err.Error(),
		}
	}

	message := strings.Join(
		[]string{
			suppliedID,
			passcode,
			r.Method,
			r.Host,
			r.URL.Path,
			r.URL.RawQuery,
			r.Header.Get("Content-Type"),
		},
		"\r\n",
	)

	hash := sha512.New()
	hash.Write([]byte(message))

	sigok := ed25519.Verify(
		ed25519.PublicKey(user.ED25519PublicKey),
		hash.Sum(nil),
		suppliedSignatrue,
	)

	if !sigok {
		return &AuthenticationFailed{
			Reason: "signature mismatch",
		}
	}

	return nil
}

func (t *service) ProxyDeviceRequest(deviceid string, path string, rw http.ResponseWriter, r *http.Request) {
	// TODO: proxy to other cluster members
	t.config.LocalDeviceProxyFunc(deviceid, path, rw, r)
}

func (t *service) Start() {
	logrus.WithFields(logrus.Fields{
		"bindAddr":    t.config.BindAddr,
		"tlsCertPath": t.config.TLSCertPath,
		"tlsKeyPath":  t.config.TLSKeyPath,
	}).Info("cluster starting")

	go t.hydrateUserCache()
	go t.hydrateMemberCache()
	go t.hydrateDeviceCache()

	t.makeDefaultAdmin()

	server := http.NewServeMux()
	router := mux.NewRouter()

	server.Handle("/", router)

	certpath := t.config.TLSCertPath
	keypath := t.config.TLSKeyPath

	if t.config.TLSCertPath == "" && t.config.TLSKeyPath == "" {
		certpath, keypath = t.makeTempCertificates()

		logrus.WithField("cert", certpath).Info("cluster temporary certificate")
		logrus.WithField("key", keypath).Info("cluster temporary key")

		defer os.Remove(certpath)
		defer os.Remove(keypath)
	}

	if err := http.ListenAndServeTLS(
		t.config.BindAddr,
		certpath,
		keypath,
		server,
	); err != nil {
		logrus.Fatal(err.Error())
	}
}

func (t *service) makeDefaultAdmin() {
	var count int

	cursor, err := db.Table(db.UserTable).Filter(db.Filter{
		"admin": true,
		"login": "admin",
	}).Count().Run(db.Session)

	if err != nil {
		logrus.Fatal(err.Error())
	}

	if err = cursor.One(&count); err != nil {
		logrus.Fatal(err.Error())
	}

	if count > 0 {
		return
	}

	adminTOTPKey, _ := totp.Generate(totp.GenerateOpts{
		Algorithm:   otp.AlgorithmSHA512,
		Issuer:      "deviceio-hub",
		AccountName: "admin@localhost",
	})

	adminPasswordPlain, _ := uuid.NewRandom()
	adminPasswordSalt, _ := uuid.NewRandom()

	hash := sha512.New()
	hash.Write([]byte(adminPasswordSalt.String() + adminPasswordPlain.String()))

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)

	if err != nil {
		logrus.WithField("error", err.Error()).Fatal("Error generating ED255519 keypair")
	}

	user := &User{
		Login:            "admin",
		Admin:            true,
		Email:            "admin@localhost",
		TOTPSecret:       []byte(adminTOTPKey.Secret()),
		PasswordHash:     hash.Sum(nil),
		PasswordSalt:     adminPasswordSalt.String(),
		ED25519PublicKey: pubKey,
	}

	resp, err := db.Table(db.UserTable).Insert(user).RunWrite(db.Session)

	if err != nil {
		logrus.Fatal(err.Error())
	}

	logrus.WithFields(logrus.Fields{
		"id":          resp.GeneratedKeys[0],
		"login":       user.Login,
		"password":    adminPasswordPlain,
		"totp_secret": string(user.TOTPSecret),
		"private_key": base64.StdEncoding.EncodeToString(privKey),
	}).Info("initial admin credentials")
}

func (t *service) makeTempCertificates() (string, string) {
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

func (t *service) hydrateUserCache() {
	t.userCache = map[string]*User{}
	t.userCacheMu = &sync.Mutex{}

	var users []*User

	cursor, err := db.Table(db.UserTable).Run(db.Session)

	if err != nil {
		logrus.Fatal(err)
	}

	cursor.All(&users)
	cursor.Close()

	linq.From(users).ForEach(func(a interface{}) {
		user := a.(*User)
		t.userCache[user.ID] = user
	})

	var changed struct {
		Old *User `gorethink:"old_val"`
		New *User `gorethink:"new_val"`
	}

	changes, err := db.Table(db.UserTable).Changes().Run(db.Session)

	for changes.Next(&changed) {
		t.userCacheMu.Lock()

		if changed.New == nil {
			_, ok := t.userCache[changed.Old.ID]

			if ok {
				delete(t.userCache, changed.Old.ID)
			}
		} else {
			t.userCache[changed.New.ID] = changed.New
		}

		t.userCacheMu.Unlock()
	}
}

func (t *service) hydrateMemberCache() {
	t.memberCache = map[string]*Member{}
	t.memberCacheMu = &sync.Mutex{}

	var members []*Member

	cursor, err := db.Table(db.MemberTable).Run(db.Session)

	if err != nil {
		logrus.Fatal(err)
	}

	cursor.All(&members)
	cursor.Close()

	t.memberCacheMu.Lock()
	for _, member := range members {
		t.memberCache[member.ID] = member
	}
	t.memberCacheMu.Unlock()

	var changed struct {
		Old *Member `gorethink:"old_val"`
		New *Member `gorethink:"new_val"`
	}

	changes, err := db.Table(db.MemberTable).Changes().Run(db.Session)

	for changes.Next(&changed) {
		t.memberCacheMu.Lock()

		if changed.New == nil {
			_, ok := t.memberCache[changed.Old.ID]

			if ok {
				delete(t.memberCache, changed.Old.ID)
			}
		} else {
			t.memberCache[changed.New.ID] = changed.New
		}

		t.memberCacheMu.Unlock()
	}
}

func (t *service) hydrateDeviceCache() {
	t.deviceCache = map[string]*Device{}
	t.deviceCacheMu = &sync.Mutex{}

	var devices []*Device

	cursor, err := db.Table(db.DeviceTable).Run(db.Session)

	if err != nil {
		logrus.Fatal(err)
	}

	cursor.All(&devices)
	cursor.Close()

	t.deviceCacheMu.Lock()
	for _, device := range devices {
		t.deviceCache[device.ID] = device
	}
	t.deviceCacheMu.Unlock()

	var changed struct {
		Old *Device `gorethink:"old_val"`
		New *Device `gorethink:"new_val"`
	}

	changes, err := db.Table(db.DeviceTable).Changes().Run(db.Session)

	for changes.Next(&changed) {
		t.deviceCacheMu.Lock()

		if changed.New == nil {
			_, ok := t.deviceCache[changed.Old.ID]

			if ok {
				delete(t.deviceCache, changed.Old.ID)
			}
		} else {
			t.deviceCache[changed.New.ID] = changed.New
		}

		t.deviceCacheMu.Unlock()
	}
}
