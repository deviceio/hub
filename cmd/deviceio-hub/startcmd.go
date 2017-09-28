package main

import "github.com/spf13/cobra"

var startcmd = &cobra.Command{
	Use:   "start",
	Short: "starts an instance of the Deviceio Hub",
	Run: func(cmd *cobra.Command, args []string) {
		start()
	},
}

var _ = func() (ret bool) {
	startcmd.PersistentFlags().StringVar(&apibind, "api-bind", ":4431", `hostname or ip address plus port to bind the api server to. formatted in standard host:port format. empty host (ex: :port) binds to all interfaces`)
	startcmd.PersistentFlags().StringVar(&apicert, "api-tls-cert", "", `path to tls cert to bind for the api server`)
	startcmd.PersistentFlags().StringVar(&apikey, "api-tls-key", "", `path to tls key to bind for the api server`)

	return
}()
