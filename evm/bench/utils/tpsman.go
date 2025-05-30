package utils

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type SimpleTPSManager struct {
	ethclient.Client
	// time recorder for average TPS
	aveStartTime time.Time
	// time recorder for instant TPS
	insStartTime time.Time
	// save start BlockNum for average TPS
	startBlockNum uint64
	// save last BlockNum query for instant TPS
	lastBlockNum uint64
	maxTPS       float64
	minTPS       float64
}

func NewTPSMan(clientURL string) *SimpleTPSManager {
	//Dial EthClient
	client, err := ethclient.Dial(clientURL)
	if err != nil {
		panic(fmt.Errorf("failed to initialize tps query client: %+v", err))
	}

	return &SimpleTPSManager{
		Client:        *client,
		aveStartTime:  time.Now(),
		insStartTime:  time.Now(),
		startBlockNum: 0,
		lastBlockNum:  0,
		maxTPS:        -1,
		minTPS:        1000000,
	}
}

func (tpsman *SimpleTPSManager) GetBlockNum() uint64 {
	var blockCount uint64
	var err error

	for {
		blockCount, err = tpsman.BlockNumber(context.Background())
		if err != nil {
			time.Sleep(200 * time.Millisecond)
		} else {
			break
		}
	}
	return blockCount
}

func (tpsman *SimpleTPSManager) BlockHeder(height uint64) *types.Header {
	for {
		header, err := tpsman.HeaderByNumber(context.Background(), big.NewInt(int64(height)))
		if err != nil {
			time.Sleep(time.Millisecond * 200)
		} else {
			return header
		}
	}
}

func (tpsman *SimpleTPSManager) transactionCountAndTimestamp(height uint64) (uint64, uint64) {
	var header *types.Header
	var err error
	for {
		header, err = tpsman.HeaderByNumber(context.Background(), big.NewInt(int64(height)))
		if err != nil {
			time.Sleep(time.Millisecond * 200)
		} else {
			break
		}
	}

	var txCount uint
	for {
		txCount, err = tpsman.TransactionCount(context.Background(), header.ParentHash)
		if err != nil {
			time.Sleep(time.Millisecond * 200)
		} else {
			return uint64(txCount), header.Time
		}
	}
}

func (tpsman *SimpleTPSManager) TPSDisplay() {
	time.Sleep(time.Second * 5)
	fmt.Println("TPSDisplay")
	var initHeight uint64
	var totalTxCount uint64
	var initTime uint64
	for {
		height := tpsman.GetBlockNum()
		header, err := tpsman.HeaderByNumber(context.Background(), big.NewInt(int64(height)))
		if err != nil {
			panic(err)
		}
		txCount, err := tpsman.TransactionCount(context.Background(), header.ParentHash)
		if err != nil {
			panic(err)
		}
		// skip this block
		if txCount > 0 {
			initHeight = height
			initTime = header.Time
			break
		} else {
			fmt.Println("height", height, "hash", header.ParentHash, "txcount", txCount)
			time.Sleep(time.Millisecond * 200)
		}
	}
	fmt.Println("initHeight", initHeight)
	lastHeight := initHeight
	var avgTPS float64
	var maxTps float64
	var minTps float64 = 100000
	for {
		newblockNum := tpsman.GetBlockNum()
		// No tx is executed
		if lastHeight == newblockNum {
			time.Sleep(1 * time.Second)
			continue
		}

		for height := lastHeight + 1; height <= newblockNum; height++ {
			txCount, timestamp := tpsman.transactionCountAndTimestamp(height)
			totalTxCount += txCount
			lastHeight = height

			avgTPS = float64(totalTxCount) / float64(timestamp-initTime)
			if avgTPS > maxTps {
				maxTps = avgTPS
			}
			if avgTPS < minTps {
				minTps = avgTPS
			}
			fmt.Println("========================================================")
			fmt.Printf("[TPS log] StartBlock Num: %d, NewBlockNum: %d, totalTxCount:%d\n", initHeight+1, height, totalTxCount)
			fmt.Printf("[Summary] Average BTPS: %5.2f, Max TPS: %5.2f, Min TPS: %5.2f, Time Last: %ds\n", avgTPS, maxTps, minTps, timestamp-initTime)
			fmt.Println("========================================================")
		}

		time.Sleep(5 * time.Second)
	}
}
