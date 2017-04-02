package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/deviceio/hub/domain"
	"github.com/deviceio/hub/infra"
	"github.com/deviceio/hub/infra/data"
	"github.com/deviceio/shared/logging"
)

var (
	cli = kingpin.New("cli", "Deviceio Hub Command Line Interface")

	startCommand            = cli.Command("start", "Starts the Hub")
	startDBHost             = startCommand.Flag("db-host", "The rethinkdb client host to connect to").Default("127.0.0.1").String()
	startDBName             = startCommand.Flag("db-name", "The name of the rethinkdb database name to use").Default("DeviceioHub").String()
	startDBUser             = startCommand.Flag("db-user", "The username to use when connecting to rethinkdb").Default("").String()
	startDBPass             = startCommand.Flag("db-pass", "The password to use when connecting to rethinkdb").Default("").String()
	startAPIBind            = startCommand.Flag("api-bind", "The hostname or ip address to bind the http api to").Default("127.0.0.1").String()
	startAPIPort            = startCommand.Flag("api-port", "The tcp port to use when binding the http api").Default("443").Int()
	startAPITLSCertPath     = startCommand.Flag("api-tls-cert-path", "Path on the local filesystem of the custom TLS certificate to use for the http api").Default("").String()
	startAPITLSKeyPath      = startCommand.Flag("api-tls-key-path", "Path on the local filesystem of the custom TLS key to use for the http api").Default("").String()
	startGatewayBind        = startCommand.Flag("gateway-bind", "The hostname or ip address to bind the gateway to").Default("127.0.0.1").String()
	startGatewayPort        = startCommand.Flag("gateway-port", "The tcp port to use when binding the gateway").Default("8975").String()
	startGatewayTLSCertPath = startCommand.Flag("gateway-tls-cert-path", "Path on the local filesystem of the custom TLS certificate to use for the gateway").Default("").String()
	startGatewayTLSKeyPath  = startCommand.Flag("gateway-tls-key-path", "Path on the local filesystem of the custom TLS key to use for the gateway").Default("").String()
)

func main() {
	flag.Parse()

	rand.Seed(time.Now().UnixNano()) //very important

	switch kingpin.MustParse(cli.Parse(os.Args[1:])) {
	case startCommand.FullCommand():
		break
	default:
		return
	}

	data.Connect(&data.Options{
		DBName: *startDBName,
		DBHost: *startDBHost,
		DBUser: *startDBUser,
		DBPass: *startDBPass,
	})
	data.Migrate()

	hub := &domain.Hub{}

	api := domain.NewAPIService(hub, &domain.APIOptions{
		BindAddr:    fmt.Sprintf("%v:%v", *startAPIBind, *startAPIPort),
		Logger:      &logging.DefaultLogger{},
		TLSCertPath: *startAPITLSCertPath,
		TLSKeyPath:  *startAPITLSKeyPath,
	})

	cluster := domain.NewClusterService(hub, &domain.ClusterOptions{
		Logger:        &logging.DefaultLogger{},
		DeviceQuery:   infra.NewClusterDeviceQuery(&logging.DefaultLogger{}),
		DeviceCommand: infra.NewClusterDeviceCommand(&logging.DefaultLogger{}),
	})

	gateway := domain.NewGatewayService(hub, &domain.GatewayOptions{
		BindAddr:    fmt.Sprintf("%v:%v", *startGatewayBind, *startGatewayPort),
		TLSCertPath: *startGatewayTLSCertPath,
		TLSKeyPath:  *startGatewayTLSKeyPath,
		Logger:      &logging.DefaultLogger{},
	})

	hub.API = api
	hub.Cluster = cluster
	hub.Gateway = gateway

	go hub.API.Start()
	go hub.Gateway.Start()

	<-make(chan bool)
}
