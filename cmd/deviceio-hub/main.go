package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/deviceio/hub/domain"
	"github.com/deviceio/hub/infra"
	"github.com/deviceio/hub/infra/data"
	"github.com/deviceio/shared/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	startCmd *cobra.Command
	setupCmd *cobra.Command
	rootCmd  *cobra.Command
)

func main() {
	rand.Seed(time.Now().UnixNano()) //very important

	startCmd = &cobra.Command{
		Use:   "start",
		Short: "starts the hub",
		Long:  `starts the api and gateway components of the hub`,
		Run: func(cmd *cobra.Command, args []string) {
			start(cmd)
		},
	}

	startCmd.Flags().String("db-host", "127.0.0.1", "Rethinkdb host to connect to")
	startCmd.Flags().String("db-name", "DeviceioHub", "Rethinkdb database name to use")
	startCmd.Flags().String("db-user", "", "Rethinkdb user to authenticate as")
	startCmd.Flags().String("db-pass", "", "Rethinkdb password to authenticate with")
	startCmd.Flags().String("api-bind-addr", "", "ip or hostname to bind the api to. Defaults to 0.0.0.0")
	startCmd.Flags().String("api-bind-port", "4431", "port to bind the api to")
	startCmd.Flags().String("api-tls-cert-path", "", "path to the api tls certificate to use. If blank an auto-generated cert will be used")
	startCmd.Flags().String("api-tls-key-path", "", "path to the api tls key to use. If blank an auto-generated cert will be used")
	startCmd.Flags().String("gateway-bind-addr", "", "ip or hostname to bind the gateway to. Defaults to 0.0.0.0")
	startCmd.Flags().String("gateway-bind-port", "8975", "port to bind the gateway to")
	startCmd.Flags().String("gateway-tls-cert-path", "", "path to the gateway tls certificate to use. If blank an auto-generated cert will be used")
	startCmd.Flags().String("gateway-tls-key-path", "", "path to the gateway tls key to use. If blank an auto-generated cert will be used")

	setupCmd = &cobra.Command{
		Use:   "setup",
		Short: "sets up initial hub configurations",
		Long: `sets up initial hub configuration including the initial admin password, hmac keys and permitted ip addresses. 
		This only needs to be called ONCE during deployment of a new hub cluster.`,
		Run: func(cmd *cobra.Command, args []string) {
			setup(cmd)
		},
	}

	setupCmd.Flags().String("db-host", "127.0.0.1", "The rethinkdb client host to connect to")
	setupCmd.Flags().String("db-name", "DeviceioHub", "The name of the rethinkdb database name to use")
	setupCmd.Flags().String("db-user", "", "The username to use when connecting to rethinkdb")
	setupCmd.Flags().String("db-pass", "", "The password to use when connecting to rethinkdb")
	setupCmd.Flags().String("user-name", "admin", "The initial cluster admin username")
	setupCmd.Flags().String("user-ip", "127.0.0.1/32", "The initial cluster admin permitted ip addresses")
	setupCmd.Flags().String("hmac-key", "", "The initial cluster admin hmac key")
	setupCmd.Flags().String("hmac-secret", "", "The initial cluster admin hmac secret")
	setupCmd.Flags().String("hmac-hash-func", "sha-256", "The hmac hash function to use for the initial user. options: sha-256")

	rootCmd = &cobra.Command{}
	rootCmd.AddCommand(startCmd, setupCmd)
	if err := rootCmd.Execute(); err != nil {
		panic("Error executing cli: " + err.Error())
	}
}

func start(cmd *cobra.Command) {
	viper.BindPFlag("db.host", cmd.Flags().Lookup("db-host"))
	viper.BindPFlag("db.name", cmd.Flags().Lookup("db-name"))
	viper.BindPFlag("db.user", cmd.Flags().Lookup("db-user"))
	viper.BindPFlag("db.pass", cmd.Flags().Lookup("db-pass"))
	viper.BindPFlag("api.bind_addr", cmd.Flags().Lookup("api-bind-addr"))
	viper.BindPFlag("api.bind_port", cmd.Flags().Lookup("api-bind-port"))
	viper.BindPFlag("api.tls_cert_path", cmd.Flags().Lookup("api-tls-cert-path"))
	viper.BindPFlag("api.tls_key_path", cmd.Flags().Lookup("api-tls-key-path"))
	viper.BindPFlag("gateway.bind_addr", cmd.Flags().Lookup("gateway-bind-addr"))
	viper.BindPFlag("gateway.bind_port", cmd.Flags().Lookup("gateway-bind-port"))
	viper.BindPFlag("gateway.tls_cert_path", cmd.Flags().Lookup("gateway-tls-cert-path"))
	viper.BindPFlag("gateway.tls_key_path", cmd.Flags().Lookup("gateway-tls-key-path"))

	viper.SetEnvPrefix("DEVICEIO_HUB_")
	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.deviceio/hub")
	viper.AddConfigPath("/etc/deviceio/hub/")
	viper.AddConfigPath("/opt/deviceio/hub/")
	viper.AddConfigPath("c:/PROGRA~1/deviceio/hub/")
	viper.AddConfigPath("c:/ProgramData/deviceio/hub/")
	viper.AddConfigPath(".")

	viper.SetDefault("db.host", "127.0.0.1")
	viper.SetDefault("db.name", "DeviceioHub")
	viper.SetDefault("db.user", "")
	viper.SetDefault("db.pass", "")
	viper.SetDefault("api.bind_addr", "")
	viper.SetDefault("api.bind_port", "4431")
	viper.SetDefault("api.tls_cert_path", "")
	viper.SetDefault("api.tls_key_path", "")
	viper.SetDefault("gateway.bind_addr", "")
	viper.SetDefault("gateway.bind_port", "8975")
	viper.SetDefault("gateway.tls_cert_path", "")
	viper.SetDefault("gateway.tls_key_path", "")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	data.Connect(&data.Options{
		DBName: viper.GetString("db.name"),
		DBHost: viper.GetString("db.host"),
		DBUser: viper.GetString("db.user"),
		DBPass: viper.GetString("db.pass"),
	})
	data.Migrate()

	hub := &domain.Hub{}

	api := domain.NewAPIService(hub, &domain.APIOptions{
		BindAddr: fmt.Sprintf(
			"%v:%v",
			viper.GetString("api.bind_addr"),
			viper.GetString("api.bind_port"),
		),
		Logger:      &logging.DefaultLogger{},
		TLSCertPath: viper.GetString("api.tls_cert_path"),
		TLSKeyPath:  viper.GetString("api.tls_key_path"),
	})

	cluster := domain.NewClusterService(hub, &domain.ClusterOptions{
		Logger:        &logging.DefaultLogger{},
		DeviceQuery:   infra.NewClusterDeviceQuery(&logging.DefaultLogger{}),
		DeviceCommand: infra.NewClusterDeviceCommand(&logging.DefaultLogger{}),
	})

	gateway := domain.NewGatewayService(hub, &domain.GatewayOptions{
		BindAddr: fmt.Sprintf(
			"%v:%v",
			viper.GetString("gateway.bind_addr"),
			viper.GetString("gateway.bind_port"),
		),
		TLSCertPath: viper.GetString("gateway.tls_cert_path"),
		TLSKeyPath:  viper.GetString("gateway.tls_key_path"),
		Logger:      &logging.DefaultLogger{},
	})

	hub.API = api
	hub.Cluster = cluster
	hub.Gateway = gateway

	go hub.API.Start()
	go hub.Gateway.Start()

	<-make(chan bool)
}

func setup(cmd *cobra.Command) {

}
