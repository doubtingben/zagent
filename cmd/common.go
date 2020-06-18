package cmd

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/doubtingben/zagent/pkg/common"
	"github.com/gofiber/fiber"
	"github.com/ybbus/jsonrpc"
)

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
			Height:           height,
			SaplingValuePool: block.SaplingValuePool(),
			SproutValuePool:  block.SproutValuePool(),
			Size:             block.Size,
			Time:             block.Time,
		}

		for _, tx := range block.TX {
			blockMetric.NumberofTransactions = blockMetric.NumberofTransactions + 1
			if tx.IsTransparent() {
				blockMetric.NumberofTransparent = blockMetric.NumberofTransparent + 1
			} else if tx.IsMixed() {
				blockMetric.NumberofMixed = blockMetric.NumberofMixed + 1
			} else if tx.IsShielded() {
				blockMetric.NumberofShielded = blockMetric.NumberofShielded + 1
			}
		}

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

	if err := app.Listen(opts.BindAddr); err != nil {
		log.Fatalf("Failed to start the frontend: %s", err)
	}

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

}
