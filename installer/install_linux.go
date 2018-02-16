package installer

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"time"

	"github.com/deviceio/shared/types"
)

func Install() {
	log.Println("Starting linux installer")

	GenerateFolders()
	GenerateCerts()
	CopyBinary()
	CopyConfig()
}

func GenerateFolders() {
	folders := []string{
		"/etc/deviceio/hub/ssl",
		"/opt/deviceio/hub/bin",
	}

	for _, folder := range folders {
		err := os.MkdirAll(folder, 0700)

		if err != nil {
			panic(err)
		}
	}
}

func GenerateCerts() {
	certgen := &types.CertGen{
		Host:      "localhost",
		IsCA:      false,
		ValidFrom: "Jan 1 15:04:05 2011",
		ValidFor:  867240 * time.Hour,
		RsaBits:   4096,
	}

	apicert, apikey := certgen.Generate()
	gwcert, gwkey := certgen.Generate()

	if !Exists("/etc/deviceio/hub/ssl/apicert") {
		ioutil.WriteFile("/etc/deviceio/hub/ssl/apicert", apicert, 0700)
	}

	if !Exists("/etc/deviceio/hub/ssl/apikey") {
		ioutil.WriteFile("/etc/deviceio/hub/ssl/apikey", apikey, 0700)
	}

	if !Exists("/etc/deviceio/hub/ssl/gatewaycert") {
		ioutil.WriteFile("/etc/deviceio/hub/ssl/gatewaycert", gwcert, 0700)
	}

	if !Exists("/etc/deviceio/hub/ssl/gatewaykey") {
		ioutil.WriteFile("/etc/deviceio/hub/ssl/gatewaykey", gwkey, 0700)
	}
}

func CopyBinary() {
	exe, err := os.Executable()

	if err != nil {
		log.Fatal(err)
	}

	err = Copy("/opt/deviceio/hub/bin/hub", exe)

	if err != nil {
		log.Fatal(err)
	}

	err = os.Chmod("/opt/deviceio/hub/bin/hub", 0700)

	if err != nil {
		log.Fatal(err)
	}
}

func CopyConfig() {
	type template struct {
		APIBind            string
		APITLSCertPath     string
		APITLSKeyPath      string
		DBName             string
		DBHost             string
		DBPassword         string
		DBUsername         string
		GatewayBind        string
		GatewayTLSCertPath string
		GatewayTLSKeyPath  string
	}

	jsonb, err := json.MarshalIndent(&template{
		APIBind:            "0.0.0.0:4431",
		APITLSCertPath:     "/etc/deviceio/hub/ssl/apicert",
		APITLSKeyPath:      "/etc/deviceio/hub/ssl/apikey",
		DBName:             "DeviceioHub",
		DBHost:             "127.0.0.1:28015",
		GatewayBind:        "0.0.0.0:8975",
		GatewayTLSCertPath: "/etc/deviceio/hub/ssl/gatewaycert",
		GatewayTLSKeyPath:  "/etc/deviceio/hub/ssl/gatewaykey",
	}, "", "    ")

	if err != nil {
		panic(err)
	}

	if !Exists("/etc/deviceio/hub/config.json") {
		err = ioutil.WriteFile("/etc/deviceio/hub/config.json", jsonb, 0700)

		if err != nil {
			panic(err)
		}
	}
}
