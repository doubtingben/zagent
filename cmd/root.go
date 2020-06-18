package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/doubtingben/zagent/cmd/run"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var log *logrus.Entry
var logger = logrus.New()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zagent",
	Short: "zagent is an agent service to the Zcash blockchain",
	Long:  `zagent is an agent service to the Zcash blockchain`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(startServerCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(blockCmd)
	rootCmd.AddCommand(printBlockCmd)
	rootCmd.AddCommand(generateMetricsCmd)
	rootCmd.AddCommand(run.RunCmd)

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is current directory, zagent.yaml)")
	rootCmd.PersistentFlags().String("bind-addr", "127.0.0.1:3000", "the address to listen on")
	rootCmd.PersistentFlags().Uint32("log-level", uint32(logrus.InfoLevel), "log level (logrus 1-7)")
	rootCmd.PersistentFlags().String("rpc-user", "zcashrpc", "rpc user account")
	rootCmd.PersistentFlags().String("rpc-password", "notsecret", "rpc password")
	rootCmd.PersistentFlags().String("rpc-host", "127.0.0.1", "rpc host")
	rootCmd.PersistentFlags().String("rpc-port", "38232", "rpc port")

	generateMetricsCmd.PersistentFlags().Int("start-height", 0, "Starting block height (working backwards)")
	generateMetricsCmd.PersistentFlags().Int("end-height", 0, "Ending block height (working backwards)")
	generateMetricsCmd.PersistentFlags().Int("num-blocks", 10, "Number of blocks")
	generateMetricsCmd.PersistentFlags().String("output-dir", "./blocks", "Output directory")

	viper.BindPFlag("bind-addr", rootCmd.PersistentFlags().Lookup("bind-addr"))
	viper.SetDefault("bind-addr", "127.0.0.1:3000")
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.SetDefault("log-level", int(logrus.InfoLevel))

	viper.BindPFlag("rpc-user", rootCmd.PersistentFlags().Lookup("rpc-user"))
	viper.SetDefault("rpc-user", "zcashrpc")
	viper.BindPFlag("rpc-password", rootCmd.PersistentFlags().Lookup("rpc-password"))
	viper.SetDefault("rpc-password", "notsecret")
	viper.BindPFlag("rpc-host", rootCmd.PersistentFlags().Lookup("rpc-host"))
	viper.SetDefault("rpc-host", "127.0.0.1")
	viper.BindPFlag("rpc-port", rootCmd.PersistentFlags().Lookup("rpc-port"))
	viper.SetDefault("rpc-port", "38232")

	viper.BindPFlag("start-height", generateMetricsCmd.PersistentFlags().Lookup("start-height"))
	viper.BindPFlag("end-height", generateMetricsCmd.PersistentFlags().Lookup("end-height"))
	viper.BindPFlag("num-blocks", generateMetricsCmd.PersistentFlags().Lookup("num-blocks"))
	viper.BindPFlag("output-dir", generateMetricsCmd.PersistentFlags().Lookup("output-dir"))

	logger.SetFormatter(&logrus.TextFormatter{
		//DisableColors:          true,
		//FullTimestamp:          true,
		//DisableLevelTruncation: true,
	})

	onexit := func() {
		fmt.Printf("zagent died with a Fatal error. Check logfile for details.\n")
	}

	log = logger.WithFields(logrus.Fields{
		"app": "zagent",
	})
	log.Logger.SetLevel(7)
	logrus.RegisterExitHandler(onexit)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Look in the current directory for a configuration file
		viper.AddConfigPath(".")
		// Viper auto appends extention to this config name
		// For example, lightwalletd.yml
		viper.SetConfigName("zagent")
	}

	// Replace `-` in config options with `_` for ENV keys
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	var err error
	if err = viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

}
