package cmd

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/doubtingben/zagent/pkg/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ybbus/jsonrpc"
)

// generateMetrics prints block metric data files
var generateMetricsCmd = &cobra.Command{
	Use:   "generateMetrics",
	Short: "Prints block metric data files",
	Run: func(cmd *cobra.Command, args []string) {
		opts := &common.Options{
			BindAddr:    viper.GetString("bind-addr"),
			RPCUser:     viper.GetString("rpc-user"),
			RPCPassword: viper.GetString("rpc-password"),
			RPCHost:     viper.GetString("rpc-host"),
			RPCPort:     viper.GetString("rpc-port"),
		}
		if err := writeMetrics(*opts); err != nil {
			log.Fatalf("Failed to write metrics file: %s", err)
		}

	},
}

type BlockMetric struct {
	Height               int `json:"height"`
	NumberofTransactions int `json:"number_of_transactions"`
}

func writeMetrics(opts common.Options) error {
	startHeight := viper.GetInt("start-height")
	numBlocks := viper.GetInt("num-blocks")
	outputDir := viper.GetString("output-dir")
	log.Infof("Wirting metrics for %d, %d blocks\n", startHeight, numBlocks)
	log.Infof("Writing to direcotry %s", outputDir)
	var metrics []*BlockMetric
	for height := startHeight - numBlocks; height <= startHeight; height++ {
		log.Infof("Getting block metrics for: %d", height)
		blockMetric, err := getBlockMetrics(height, opts)
		if err != nil {
			log.Warnf("Error getting block #%d metrics: %s", height, err)
		}
		metrics = append(metrics, blockMetric)
	}

	blockFile := outputDir + "/zcashmetrics.json"
	blockJSON, err := json.MarshalIndent(metrics, "", "    ")
	if err != nil {
		log.Fatalln("printBlocktoFile MarshalIndent error: %s", err)
	}
	return ioutil.WriteFile(blockFile, blockJSON, 0644)
}

func getBlockMetrics(height int, opts common.Options) (*BlockMetric, error) {
	basicAuth := base64.StdEncoding.EncodeToString([]byte(opts.RPCUser + ":" + opts.RPCPassword))
	rpcClient := jsonrpc.NewClientWithOpts("http://"+opts.RPCHost+":"+opts.RPCPort,
		&jsonrpc.RPCClientOpts{
			CustomHeaders: map[string]string{
				"Authorization": "Basic " + basicAuth,
			}})
	var block *common.Block

	err := rpcClient.CallFor(&block, "getblock", strconv.Itoa(height), 2)
	if err != nil {
		return nil, err
	}

	//var blockMetric *BlockMetric
	blockMetric := &BlockMetric{
		Height:               height,
		NumberofTransactions: block.NumberofTransactions(),
	}
	blockMetric.Height = height
	blockMetric.NumberofTransactions = block.NumberofTransactions()

	return blockMetric, nil
}
