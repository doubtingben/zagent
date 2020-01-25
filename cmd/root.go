package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/doubtingben/zagent/common"

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
	Run: func(cmd *cobra.Command, args []string) {
		opts := &common.Options{
			BindAddr: viper.GetString("bind-addr"),
		}

		// Start server and block, or exit
		if err := startServer(opts); err != nil {
			log.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("couldn't create server")
		}
	},
}

func startServer(opts *common.Options) error {

	fmt.Printf("started Server\n")
	return nil

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
	rootCmd.AddCommand(versionCmd)
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is current directory, lightwalletd.yaml)")
	rootCmd.Flags().String("bind-addr", "127.0.0.1:9067", "the address to listen on")
	rootCmd.Flags().Int("log-level", int(logrus.InfoLevel), "log level (logrus 1-7)")

	viper.BindPFlag("bind-addr", rootCmd.Flags().Lookup("bind-addr"))
	viper.SetDefault("bind-addr", "127.0.0.1:9067")
	viper.BindPFlag("log-level", rootCmd.Flags().Lookup("log-level"))
	viper.SetDefault("log-level", int(logrus.InfoLevel))

	logger.SetFormatter(&logrus.TextFormatter{
		//DisableColors:          true,
		FullTimestamp:          true,
		DisableLevelTruncation: true,
	})

	onexit := func() {
		fmt.Printf("zagent died with a Fatal error. Check logfile for details.\n")
	}

	log = logger.WithFields(logrus.Fields{
		"app": "zagent",
	})

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
