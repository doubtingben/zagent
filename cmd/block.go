package cmd

import (
	"encoding/base64"
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

		err = printBlock(height, *opts)

	},
}

func printBlock(height int, opts common.Options) error {
	basicAuth := base64.StdEncoding.EncodeToString([]byte(opts.RPCUser + ":" + opts.RPCPassword))
	rpcClient := jsonrpc.NewClientWithOpts("http://"+opts.RPCHost+":"+opts.RPCPort,
		&jsonrpc.RPCClientOpts{
			CustomHeaders: map[string]string{
				"Authorization": "Basic " + basicAuth,
			}})

	log.Infof("Starting at block: %d", height)

	for currentHeight := height; (height - currentHeight) < 20; currentHeight-- {
		var block *common.Block

		err := rpcClient.CallFor(&block, "getblock", strconv.Itoa(currentHeight), 2)

		if err != nil {
			log.Fatalln("Error at block %d calling getblock: %s", currentHeight, err)
		}

		log.Debugf("Block #%d: %+v", currentHeight, block)
		log.Infof("Block #%d has %d transactions at %s", currentHeight, block.NumberofTransactions(), time.Unix(block.Time, 0))
		for i, t := range block.TX {
			vin, vout, vjoinsplit := t.TransactionTypes()
			if len(t.VShieldedOutput) > 0 {
				log.Infof("Block #%d, transaction %d has %d vin, %d vout and %d vjoinsplits", currentHeight, i, vin, vout, vjoinsplit)
				log.Infof("Block #%d, transaction %d %#v", currentHeight, i, t)
			}
			//log.Infof("Block #%d, transaction %d %#v", currentHeight, i, t)

		}
	}
	return nil
}
