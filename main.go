package main

import (
	"log"

	"github.com/deviceio/hub/domain"
	"github.com/deviceio/hub/infra"
	"github.com/deviceio/hub/infra/data"
	"github.com/deviceio/shared/config"
	"github.com/deviceio/shared/logging"
)

func main() {
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
	config.AddFileName("config.json")
	config.AddFilePath("/etc/github.com/deviceio/hub")
	config.AddFilePath("c:/ProgramData/github.com/deviceio/hub")

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
