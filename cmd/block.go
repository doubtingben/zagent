package cmd

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/doubtingben/zagent/pkg/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ybbus/jsonrpc"
)

// blockCmd represents the block command
var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Get a block",
	Run: func(cmd *cobra.Command, args []string) {
		numToProcess := 10
		if len(args) == 0 {
			log.Fatalln("A block height is required")
		}
		if len(args) == 2 {
			var err error
			if numToProcess, err = strconv.Atoi(args[1]); err != nil {
				log.Fatalf("Number of blocks must be an integer")
			}
		}
		height, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalln("Blocks must be an integer: %v", err)
		}
		opts := &common.Options{
			BindAddr:    viper.GetString("bind-addr"),
			RPCUser:     viper.GetString("rpc-user"),
			RPCPassword: viper.GetString("rpc-password"),
			RPCHost:     viper.GetString("rpc-host"),
			RPCPort:     viper.GetString("rpc-port"),
		}

		err = printBlocks(height, numToProcess, *opts)

	},
}

// printBlockCmd represents the block command
var printBlockCmd = &cobra.Command{
	Use:   "printBlockCmd",
	Short: "Print a block",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatalln("A block height is required")
		}
		height, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalln("Blocks must be an integer: %v", err)
		}
		opts := &common.Options{
			BindAddr:    viper.GetString("bind-addr"),
			RPCUser:     viper.GetString("rpc-user"),
			RPCPassword: viper.GetString("rpc-password"),
			RPCHost:     viper.GetString("rpc-host"),
			RPCPort:     viper.GetString("rpc-port"),
		}

		err = printBlocktoFile(height, ".", *opts)

	},
}

func printBlocktoFile(height int, filePath string, opts common.Options) error {
	basicAuth := base64.StdEncoding.EncodeToString([]byte(opts.RPCUser + ":" + opts.RPCPassword))
	rpcClient := jsonrpc.NewClientWithOpts("http://"+opts.RPCHost+":"+opts.RPCPort,
		&jsonrpc.RPCClientOpts{
			CustomHeaders: map[string]string{
				"Authorization": "Basic " + basicAuth,
			}})

	var blockFile string
	if filePath == "." {
		blockFile = filePath + strconv.Itoa(height) + ".json"

	} else {
		blockFile = strconv.Itoa(height) + ".json"
	}
	log.Infof("Printing block #%d to file %s", height, blockFile)

	var block *common.Block

	err := rpcClient.CallFor(&block, "getblock", strconv.Itoa(height), 2)

	if err != nil {
		log.Fatalln("printBlocktoFile getblock error: %s", err)
	}

	blockJSON, err := json.MarshalIndent(block, "", "    ")
	if err != nil {
		log.Fatalln("printBlocktoFile MarshalIndent error: %s", err)
	}

	return ioutil.WriteFile(blockFile, blockJSON, 0644)
}

func printBlocks(height int, numToProcess int, opts common.Options) error {
	basicAuth := base64.StdEncoding.EncodeToString([]byte(opts.RPCUser + ":" + opts.RPCPassword))
	rpcClient := jsonrpc.NewClientWithOpts("http://"+opts.RPCHost+":"+opts.RPCPort,
		&jsonrpc.RPCClientOpts{
			CustomHeaders: map[string]string{
				"Authorization": "Basic " + basicAuth,
			}})

	log.Infof("Starting at block: %d", height)

	for currentHeight := height; (height - currentHeight) < numToProcess; currentHeight-- {
		var block *common.Block

		err := rpcClient.CallFor(&block, "getblock", strconv.Itoa(currentHeight), 2)

		if err != nil {
			log.Fatalln("Error at block %d calling getblock: %s", currentHeight, err)
		}

		log.Debugf("Block #%d: %+v", currentHeight, block)
		log.Infof("Block #%d has %d transactions at %s", currentHeight, block.NumberofTransactions(), time.Unix(block.Time, 0))
	}

	return nil
}
