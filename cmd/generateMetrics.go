package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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
		basicAuth := base64.StdEncoding.EncodeToString([]byte(opts.RPCUser + ":" + opts.RPCPassword))
		rpcClient := jsonrpc.NewClientWithOpts("http://"+opts.RPCHost+":"+opts.RPCPort,
			&jsonrpc.RPCClientOpts{
				CustomHeaders: map[string]string{
					"Authorization": "Basic " + basicAuth,
				}})

		if err := generateMetrics(rpcClient); err != nil {
			log.Fatalf("Failed to write metrics file: %s", err)
		}

	},
}

func generateMetrics(rpcClient jsonrpc.RPCClient) error {
	outputDir := viper.GetString("output-dir")
	numBlocks := viper.GetInt("num-blocks")
	if numBlocks == 0 {
		numBlocks = 10
	}
	currentHeight, err := getCurrentHeight(rpcClient)
	if err != nil {
		return err
	}
	var endHeight *int = new(int)
	*endHeight = 419200
	fmt.Printf("Getting metrics startng at %d through %d\n", *currentHeight, *endHeight)
	metrics, err := getFiberMetrics(currentHeight, endHeight, rpcClient)
	if err != nil {
		return err
	}

	blockFile := outputDir + "/zcashmetrics.json"
	blockJSON, err := json.MarshalIndent(metrics, "", "    ")
	if err != nil {
		log.Fatalln("generateMetrics MarshalIndent error: %s", err)
	}
	return ioutil.WriteFile(blockFile, blockJSON, 0644)

}

func writeMetrics(opts common.Options) error {
	startHeight := viper.GetInt("start-height")
	numBlocks := viper.GetInt("num-blocks")
	outputDir := viper.GetString("output-dir")
	log.Infof("Writing metrics for %d, %d blocks\n", startHeight, numBlocks)
	log.Infof("Writing to direcotry %s", outputDir)
	var metrics []*common.BlockMetric
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

func getBlockMetrics(height int, opts common.Options) (*common.BlockMetric, error) {
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
	blockMetric := &common.BlockMetric{
		Height:               height,
		NumberofTransactions: block.NumberofTransactions(),
		SaplingValuePool:     block.SaplingValuePool(),
		SproutValuePool:      block.SaplingValuePool(),
	}
	blockMetric.Height = height
	blockMetric.NumberofTransactions = block.NumberofTransactions()

	return blockMetric, nil
}
