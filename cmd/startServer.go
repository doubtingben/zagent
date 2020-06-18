package cmd

import (
	"github.com/doubtingben/zagent/pkg/common"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var startServerCmd = &cobra.Command{
	Use:   "startServer",
	Short: "startServer runs an http interface for a zcashd instance",
	Run: func(cmd *cobra.Command, args []string) {
		opts := &common.Options{
			BindAddr:    viper.GetString("bind-addr"),
			RPCUser:     viper.GetString("rpc-user"),
			RPCPassword: viper.GetString("rpc-password"),
			RPCHost:     viper.GetString("rpc-host"),
			RPCPort:     viper.GetString("rpc-port"),
		}
		log.Warnf("Options: %#v", opts)

		// Start server and block, or exit
		if err := startServer(opts); err != nil {
			log.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("couldn't create server")
		}
	},
}
