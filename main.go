package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/deviceio/hub/domain"
	"github.com/deviceio/hub/infra"
	"github.com/deviceio/hub/infra/data"
	"github.com/deviceio/hub/installer"
	"github.com/deviceio/shared/config"
	"github.com/deviceio/shared/logging"
)

var install = flag.Bool("install", false, "Installs the program")

func main() {
	flag.Parse()

	if *install {
		installer.Install()
		return
	}

	rand.Seed(time.Now().UnixNano()) //very important

	var configuration struct {
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

	config.SetConfigStruct(&configuration)
	config.AddConfigPath("/etc/deviceio/hub/config.json")
	config.AddConfigPath("c:/ProgramData/deviceio/hub/config.json")

	if err := config.Parse(); err != nil {
		log.Fatal(err)
	}

	data.Connect(&data.Options{
		DBName: configuration.DBName,
		DBHost: configuration.DBHost,
	})
	data.Migrate()

	hub := &domain.Hub{}

	api := domain.NewAPIService(hub, &domain.APIOptions{
		Logger:      &logging.DefaultLogger{},
		BindAddr:    configuration.APIBind,
		TLSCertPath: configuration.APITLSCertPath,
		TLSKeyPath:  configuration.APITLSKeyPath,
	})

	cluster := domain.NewClusterService(hub, &domain.ClusterOptions{
		Logger:        &logging.DefaultLogger{},
		DeviceQuery:   infra.NewClusterDeviceQuery(&logging.DefaultLogger{}),
		DeviceCommand: infra.NewClusterDeviceCommand(&logging.DefaultLogger{}),
	})

	gateway := domain.NewGatewayService(hub, &domain.GatewayOptions{
		BindAddr:    configuration.GatewayBind,
		TLSCertPath: configuration.GatewayTLSCertPath,
		TLSKeyPath:  configuration.GatewayTLSKeyPath,
		Logger:      &logging.DefaultLogger{},
	})

	hub.API = api
	hub.Cluster = cluster
	hub.Gateway = gateway

	go hub.API.Start()
	go hub.Gateway.Start()

	<-make(chan bool)
}
