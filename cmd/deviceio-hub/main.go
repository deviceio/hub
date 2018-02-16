package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/deviceio/hub/api"
	"github.com/deviceio/hub/cluster"
	"github.com/deviceio/hub/db"
	"github.com/deviceio/hub/gateway"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/palantir/stacktrace"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	startCmd *cobra.Command
	initCmd  *cobra.Command
	rootCmd  *cobra.Command
)

func main() {
	rand.Seed(time.Now().UnixNano()) //very important

	startCmd = &cobra.Command{
		Use:   "start",
		Short: "starts the hub",
		Long:  `starts the api, cluster and gateway components of the hub`,
		Run: func(cmd *cobra.Command, args []string) {
			start(cmd, false)
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
	startCmd.Flags().String("cluster-bind-addr", "", "ip or hostname to bind to the cluster instance")
	startCmd.Flags().String("cluster-bind-port", "5531", "port to bind to the cluster instance")
	startCmd.Flags().String("cluster-tls-cert-path", "", "path to the cluster tls certificate to use. If blank an auto-generated cert will be used")
	startCmd.Flags().String("cluster-tls-key-path", "", "path to the cluster tls key to use. If blank an auto-generated cert will be used")
	startCmd.Flags().String("gateway-bind-addr", "", "ip or hostname to bind the gateway to. Defaults to 0.0.0.0")
	startCmd.Flags().String("gateway-bind-port", "8975", "port to bind the gateway to")
	startCmd.Flags().String("gateway-tls-cert-path", "", "path to the gateway tls certificate to use. If blank an auto-generated cert will be used")
	startCmd.Flags().String("gateway-tls-key-path", "", "path to the gateway tls key to use. If blank an auto-generated cert will be used")

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "init hub cluster",
		Long:  `initializes the hub cluster by generating the initial admin user`,
		Run: func(cmd *cobra.Command, args []string) {
			start(cmd, true)
		},
	}

	initCmd.Flags().String("db-host", "127.0.0.1", "Rethinkdb host to connect to")
	initCmd.Flags().String("db-name", "DeviceioHub", "Rethinkdb database name to use")
	initCmd.Flags().String("db-user", "", "Rethinkdb user to authenticate as")
	initCmd.Flags().String("db-pass", "", "Rethinkdb password to authenticate with")

	rootCmd = &cobra.Command{}
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(stacktrace.Propagate(err, "Error executing cli"))
	}
}

func start(cmd *cobra.Command, init bool) {
	homedir, err := homedir.Dir()

	if err != nil {
		log.Fatal(stacktrace.Propagate(err, "failed to locate home directory"))
	}

	viper.BindPFlag("db.host", cmd.Flags().Lookup("db-host"))
	viper.BindPFlag("db.name", cmd.Flags().Lookup("db-name"))
	viper.BindPFlag("db.user", cmd.Flags().Lookup("db-user"))
	viper.BindPFlag("db.pass", cmd.Flags().Lookup("db-pass"))
	viper.BindPFlag("api.bind_addr", cmd.Flags().Lookup("api-bind-addr"))
	viper.BindPFlag("api.bind_port", cmd.Flags().Lookup("api-bind-port"))
	viper.BindPFlag("api.tls_cert_path", cmd.Flags().Lookup("api-tls-cert-path"))
	viper.BindPFlag("api.tls_key_path", cmd.Flags().Lookup("api-tls-key-path"))
	viper.BindPFlag("cluster.bind_addr", cmd.Flags().Lookup("cluster-bind-addr"))
	viper.BindPFlag("cluster.bind_port", cmd.Flags().Lookup("cluster-bind-port"))
	viper.BindPFlag("cluster.tls_cert_path", cmd.Flags().Lookup("cluster-tls-cert-path"))
	viper.BindPFlag("cluster.tls_key_path", cmd.Flags().Lookup("cluster-tls-key-path"))
	viper.BindPFlag("gateway.bind_addr", cmd.Flags().Lookup("gateway-bind-addr"))
	viper.BindPFlag("gateway.bind_port", cmd.Flags().Lookup("gateway-bind-port"))
	viper.BindPFlag("gateway.tls_cert_path", cmd.Flags().Lookup("gateway-tls-cert-path"))
	viper.BindPFlag("gateway.tls_key_path", cmd.Flags().Lookup("gateway-tls-key-path"))

	viper.SetEnvPrefix("DEVICEIO_HUB_")
	viper.SetConfigName("config")
	viper.AddConfigPath(fmt.Sprintf("%v/.deviceio/hub/", homedir))
	viper.AddConfigPath("$HOME/.deviceio/hub/")
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
	viper.SetDefault("cluster.bind_addr", "")
	viper.SetDefault("cluster.bind_port", "5531")
	viper.SetDefault("cluster.tls_cert_path", "")
	viper.SetDefault("cluster.tls_key_path", "")
	viper.SetDefault("gateway.bind_addr", "")
	viper.SetDefault("gateway.bind_port", "8975")
	viper.SetDefault("gateway.tls_cert_path", "")
	viper.SetDefault("gateway.tls_key_path", "")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	logrus.WithField("config", viper.ConfigFileUsed()).Println("configuration loaded")

	db.Connect(&db.Options{
		DBName: viper.GetString("db.name"),
		DBHost: viper.GetString("db.host"),
		DBUser: viper.GetString("db.user"),
		DBPass: viper.GetString("db.pass"),
	})
	db.Migrate()

	if init {
		cluster.NewService(&cluster.Config{}).Initialize()
		return
	}

	gatewayService := &gateway.Service{
		BindAddr: fmt.Sprintf(
			"%v:%v",
			viper.GetString("gateway.bind_addr"),
			viper.GetString("gateway.bind_port"),
		),
		TLSCertPath: viper.GetString("gateway.tls_cert_path"),
		TLSKeyPath:  viper.GetString("gateway.tls_key_path"),
	}

	clusterService := cluster.NewService(&cluster.Config{
		BindAddr: fmt.Sprintf(
			"%v:%v",
			viper.GetString("cluster.bind_addr"),
			viper.GetString("cluster.bind_port"),
		),
		TLSCertPath: viper.GetString("cluster.tls_cert_path"),
		TLSKeyPath:  viper.GetString("cluster.tls_key_path"),
		LocalDeviceProxyFunc: func(deviceid string, path string, rw http.ResponseWriter, r *http.Request) error {
			if deviceid == "" {
				return stacktrace.NewError("deviceid is empty")
			}

			if rw == nil {
				return stacktrace.NewError("http.ResponseWriter is nil")
			}

			if r == nil {
				return stacktrace.NewError("http.Request is nil")
			}

			err := gatewayService.ProxyHTTPRequest(deviceid, path, rw, r)

			if err != nil {
				return stacktrace.Propagate(err, "device proxy func failed")
			}

			return nil
		},
	})

	apiService := &api.Service{
		BindAddr: fmt.Sprintf(
			"%v:%v",
			viper.GetString("api.bind_addr"),
			viper.GetString("api.bind_port"),
		),
		TLSCertPath: viper.GetString("api.tls_cert_path"),
		TLSKeyPath:  viper.GetString("api.tls_key_path"),
		Controllers: []api.Controller{
			&api.UserController{},
			&api.StatusController{},
			&api.DeviceController{
				ClusterService: clusterService,
			},
		},
	}

	go apiService.Start()
	go clusterService.Start()
	go gatewayService.Start()

	<-make(chan bool)
}
