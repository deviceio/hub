package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootcmd = &cobra.Command{
	Use:   "deviceio-hub",
	Short: "Deviceio Hub provides centralized access to all of your devices",
	Long:  `Deviceio Hub provides centralized access to all of your devices to conduct real-time automation and orchestration powering next generation IT asset management`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var _ = func() (ret bool) {
	rootcmd.PersistentFlags().StringVar(&dbhost, "db-host", "localhost", `hostname or ip address of the hub database`)
	rootcmd.PersistentFlags().IntVar(&dbport, "db-port", 5432, `port of the hub database`)
	rootcmd.PersistentFlags().StringVar(&dbuser, "db-user", "deviceio", `username of the hub database`)
	rootcmd.PersistentFlags().StringVar(&dbpass, "db-pass", "", `password of the hub database user`)

	viper.BindPFlag("db_host", rootcmd.PersistentFlags().Lookup("db-host"))
	viper.BindPFlag("db_port", rootcmd.PersistentFlags().Lookup("db-port"))
	viper.BindPFlag("db_user", rootcmd.PersistentFlags().Lookup("db-user"))
	viper.BindPFlag("db_pass", rootcmd.PersistentFlags().Lookup("db-pass"))
	viper.SetDefault("db_host", "localhost")
	viper.SetDefault("db_port", 5432)
	viper.SetDefault("db_user", "deviceio")
	viper.SetDefault("db_pass", "")

	return
}()
