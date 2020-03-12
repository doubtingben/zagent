package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/doubtingben/zagent/pkg/common"

	"github.com/gofiber/fiber"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ybbus/jsonrpc"
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

func startServer(opts *common.Options) error {
	go startFrontend(opts)
	if err := connectZcash(opts); err != nil {
		log.Warnf("error starting zcash connections: %s", err)
	}
	log.Infof("started Server\n")
	return nil
}

func getFiberMetrics(startHeight *int, endHeight *int, rpcClient jsonrpc.RPCClient) ([]*common.BlockMetric, error) {
	if startHeight == nil {
		currentHeight, err := getCurrentHeight(rpcClient)
		if err != nil {
			return nil, err
		}
		startHeight = currentHeight
	}
	if endHeight == nil {
		*endHeight = *startHeight - 100
		if *endHeight < 0 {
			*endHeight = 0
		}
	}
	if *endHeight > *startHeight {
		return nil, fmt.Errorf("End height before Start height, bailing")
	}
	var blockMetrics []*common.BlockMetric

	for height := *endHeight; height <= *startHeight; height++ {
		var block *common.Block
		log.Debugf("Calling getblock for block %d", height)
		err := rpcClient.CallFor(&block, "getblock", strconv.Itoa(height), 2)
		if err != nil {
			return nil, err
		}

		//var blockMetric *BlockMetric
		blockMetric := &common.BlockMetric{
			Height:               height,
			NumberofTransactions: block.NumberofTransactions(),
			SaplingValuePool:     block.SaplingValuePool(),
			SproutValuePool:      block.SproutValuePool(),
		}
		blockMetric.Height = height
		blockMetric.NumberofTransactions = block.NumberofTransactions()

		blockMetrics = append(blockMetrics, blockMetric)
	}
	return blockMetrics, nil
}

func startFrontend(opts *common.Options) {
	basicAuth := base64.StdEncoding.EncodeToString([]byte(opts.RPCUser + ":" + opts.RPCPassword))
	rpcClient := jsonrpc.NewClientWithOpts("http://"+opts.RPCHost+":"+opts.RPCPort,
		&jsonrpc.RPCClientOpts{
			CustomHeaders: map[string]string{
				"Authorization": "Basic " + basicAuth,
			}})

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) {
		c.Send("zagent!")
	})

	app.Get("/metrics", func(c *fiber.Ctx) {
		currentHeight, err := getCurrentHeight(rpcClient)
		if err != nil {
			c.Next(err)
		}
		metrics, err := getFiberMetrics(currentHeight, nil, rpcClient)
		if err != nil {
			c.Next(err)
		}
		c.Set("Content-Type", "application/json")
		c.JSON(metrics)
	})

	app.Use("/metrics", func(c *fiber.Ctx) {
		c.Set("Content-Type", "application/json")
		c.Status(500).Send(c.Error())
	})

	app.Static("public", "./public")
	// app.Use("/ws", func(c *fiber.Ctx) {
	// 	// if c.Get("host") == "localhost:3000" {
	// 	// 	c.Status(403).Send("Request origin not allowed")
	// 	// } else {
	// 	c.Next()
	// 	//}
	// })
	// Upgraded websocket request
	app.WebSocket("/ws", func(c *fiber.Conn) {
		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", msg)
			err = c.WriteMessage(mt, msg)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	})

	if err := app.Listen(3001); err != nil {
		log.Fatalf("Failed to start the frontend: %s", err)
	}

}

type GetInfo struct {
	Version int `json:"version"`
}

func getCurrentHeight(rpcClient jsonrpc.RPCClient) (currentHeight *int, err error) {
	var blockChainInfo *common.GetBlockchainInfo
	if err := rpcClient.CallFor(&blockChainInfo, "getblockchaininfo"); err != nil {
		return nil, err
	}
	height := &blockChainInfo.Blocks
	return height, nil
}

func connectZcash(opts *common.Options) error {
	basicAuth := base64.StdEncoding.EncodeToString([]byte(opts.RPCUser + ":" + opts.RPCPassword))
	rpcClient := jsonrpc.NewClientWithOpts("http://"+opts.RPCHost+":"+opts.RPCPort,
		&jsonrpc.RPCClientOpts{
			CustomHeaders: map[string]string{
				"Authorization": "Basic " + basicAuth,
			}})
	var blockChainInfo *common.GetBlockchainInfo
	var currentHeight int

	for {
		if err := rpcClient.CallFor(&blockChainInfo, "getblockchaininfo"); err != nil {
			log.Warnln("Error calling getblockchaininfo", err)
			time.Sleep(time.Duration(10) * time.Second)
			continue
		}
		log.Debugf("getblockchaininfo: %#v", blockChainInfo)
		if currentHeight < blockChainInfo.Blocks {
			currentHeight = blockChainInfo.Blocks
			log.Infof("got new block! %d\n", blockChainInfo.Blocks)
			go processBlock(rpcClient, blockChainInfo.Blocks)
		}
		time.Sleep(time.Duration(10) * time.Second)
	}
}

func processBlock(client jsonrpc.RPCClient, height int) {
	log.Infof("Processing block: %d", height)
	var block *common.Block

	err := client.CallFor(&block, "getblock", strconv.Itoa(height), 2)

	if err != nil {
		log.Warnf("Error calling getblock: %s", err)
	}

	log.Debugf("Block #%d: %+v", height, block)
	log.Infof("Block #%d has %d transactions at %s", height, block.NumberofTransactions(), time.Unix(block.Time, 0))
	for i, t := range block.TX {
		vin, vout, vjoinsplit := t.TransactionTypes()
		log.Infof("Block #%d, transaction %d has %d vin, %d vout and %d vjoinsplits", height, i, vin, vout, vjoinsplit)
	}

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
	rootCmd.AddCommand(blockCmd)
	rootCmd.AddCommand(printBlockCmd)
	rootCmd.AddCommand(generateMetricsCmd)

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is current directory, zagent.yaml)")
	rootCmd.PersistentFlags().String("bind-addr", "127.0.0.1:9067", "the address to listen on")
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
	viper.SetDefault("bind-addr", "127.0.0.1:9067")
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
