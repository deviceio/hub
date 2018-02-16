package cluster

import (
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"sync"

	"encoding/base64"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/ed25519"
)

type ServiceTestSuite struct {
	suite.Suite
	service *service
}

func (t *ServiceTestSuite) SetupTest() {
	t.service = &service{}
}

func (t *ServiceTestSuite) Test_AuthenticateAPIRequest_failure_on_missing_auth_header_value() {
	req := &http.Request{}

	err := t.service.AuthenticateAPIRequest(req)

	assert.NotEqual(t.T(), err, nil)
	assert.Equal(t.T(), "authentication header empty", err.Error())

	authfailed, ok := err.(*AuthenticationFailed)

	assert.True(t.T(), ok)
	assert.Equal(t.T(), "authentication header empty", authfailed.Reason)
}

func (t *ServiceTestSuite) Test_AuthenticateAPIRequest_failure_on_no_auth_header_value() {
	req, _ := http.NewRequest("whatever", "https://somethign.com/", nil)

	req.Header.Set("Authorization", "type-but-no-value     ")

	err := t.service.AuthenticateAPIRequest(req)

	assert.NotEqual(t.T(), err, nil)
	assert.Equal(t.T(), "authorization header does not contain valid type and value", err.Error())

	authfailed, ok := err.(*AuthenticationFailed)

	assert.True(t.T(), ok)
	assert.Equal(t.T(), "authorization header does not contain valid type and value", authfailed.Reason)
}

func (t *ServiceTestSuite) Test_AuthenticateAPIRequest_failure_on_invalid_auth_header_type() {
	req, _ := http.NewRequest("whatever", "https://somethign.com/", nil)

	req.Header.Set("Authorization", "invalid-type value")

	err := t.service.AuthenticateAPIRequest(req)

	assert.NotEqual(t.T(), err, nil)
	assert.Equal(t.T(), "authorization header <type> must be 'DEVICEIO-HUB-AUTH'", err.Error())

	authfailed, ok := err.(*AuthenticationFailed)

	assert.True(t.T(), ok)
	assert.Equal(t.T(), "authorization header <type> must be 'DEVICEIO-HUB-AUTH'", authfailed.Reason)
}

func (t *ServiceTestSuite) Test_AuthenticateAPIRequest_failure_on_invalid_auth_header_formatting() {
	req, _ := http.NewRequest("whatever", "https://somethign.com/", nil)

	req.Header.Set("Authorization", "DEVICEIO-HUB-AUTH invalid-formatting")

	err := t.service.AuthenticateAPIRequest(req)

	assert.NotEqual(t.T(), err, nil)
	assert.Equal(t.T(), "authorization value does not have required format <user_id>:<ed25519_signature_base64>", err.Error())

	authfailed, ok := err.(*AuthenticationFailed)

	assert.True(t.T(), ok)
	assert.Equal(t.T(), "authorization value does not have required format <user_id>:<ed25519_signature_base64>", authfailed.Reason)
}

func (t *ServiceTestSuite) Test_AuthenticateAPIRequest_signature_mismatch_when_totp_passcode_expires() {
	totpkey, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "dfasdfasdf",
		AccountName: "sdfsdfsdf",
	})

	if err != nil {
		t.T().Fatalf(err.Error())
	}

	pubkey, privkey, err := ed25519.GenerateKey(rand.Reader)

	if err != nil {
		t.T().Fatalf(err.Error())
	}

	t.service.userCacheMu = &sync.Mutex{}
	t.service.userCache = map[string]*User{
		"whatever": &User{
			ID:               "whatever",
			Login:            "admin",
			Email:            "admin@localhost",
			TOTPSecret:       []byte(totpkey.Secret()),
			ED25519PublicKey: pubkey,
		},
	}

	r, err := http.NewRequest("GET", "https://something.com/?one=foo&two=bar", nil)

	if err != nil {
		t.T().Fatalf(err.Error())
	}

	passcode, err := totp.GenerateCode(totpkey.Secret(), time.Now().Add(-30*time.Second))

	if err != nil {
		t.T().Fatalf(err.Error())
	}

	hash := sha512.New()
	hash.Write([]byte(strings.Join([]string{
		"whatever",
		passcode,
		r.Method,
		r.Host,
		r.URL.Path,
		r.URL.RawQuery,
		r.Header.Get("Content-Type"),
	}, "\r\n")))

	signed := base64.StdEncoding.EncodeToString(ed25519.Sign(privkey, hash.Sum(nil)))

	r.Header.Set("Authorization", fmt.Sprintf("%v %v:%v", "DEVICEIO-HUB-AUTH", "whatever", signed))

	err = t.service.AuthenticateAPIRequest(r)

	assert.NotNil(t.T(), err)
	assert.Equal(t.T(), "signature mismatch", err.Error())
}

func (t *ServiceTestSuite) Test_AuthenticateAPIRequest_valid_authentication() {
	totpkey, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "dfasdfasdf",
		AccountName: "sdfsdfsdf",
	})

	if err != nil {
		t.T().Fatalf(err.Error())
	}

	pubkey, privkey, err := ed25519.GenerateKey(rand.Reader)

	if err != nil {
		t.T().Fatalf(err.Error())
	}

	t.service.userCacheMu = &sync.Mutex{}
	t.service.userCache = map[string]*User{
		"whatever": &User{
			ID:               "whatever",
			Login:            "admin",
			Email:            "admin@localhost",
			TOTPSecret:       []byte(totpkey.Secret()),
			ED25519PublicKey: pubkey,
		},
	}

	r, err := http.NewRequest("GET", "https://something.com/?one=foo&two=bar", nil)

	if err != nil {
		t.T().Fatalf(err.Error())
	}

	passcode, err := totp.GenerateCode(totpkey.Secret(), time.Now())

	if err != nil {
		t.T().Fatalf(err.Error())
	}

	hash := sha512.New()
	hash.Write([]byte(strings.Join([]string{
		"whatever",
		passcode,
		r.Method,
		r.Host,
		r.URL.Path,
		r.URL.RawQuery,
		r.Header.Get("Content-Type"),
	}, "\r\n")))

	signed := base64.StdEncoding.EncodeToString(ed25519.Sign(privkey, hash.Sum(nil)))

	r.Header.Set("Authorization", fmt.Sprintf("%v %v:%v", "DEVICEIO-HUB-AUTH", "whatever", signed))

	err = t.service.AuthenticateAPIRequest(r)

	assert.Nil(t.T(), err)
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
