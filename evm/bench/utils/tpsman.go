package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type SimpleTPSManager struct {
	client ethclient.Client
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
	//Init BlockNumber
	var initBlockNum uint64
	for {
		initBlockNum, err = client.BlockNumber(context.Background())
		if err != nil {
			time.Sleep(1000 * time.Microsecond)
		} else {
			break
		}
	}
	return &SimpleTPSManager{
		client:        *client,
		aveStartTime:  time.Now(),
		insStartTime:  time.Now(),
		startBlockNum: initBlockNum,
		lastBlockNum:  initBlockNum,
		maxTPS:        -1,
		minTPS:        1000000,
	}
}

func (tpsman *SimpleTPSManager) GetBlockNum() uint64 {
	var blockCount uint64
	var err error

	for {
		blockCount, err = tpsman.client.BlockNumber(context.Background())
		if err != nil {
			time.Sleep(1000 * time.Microsecond)
		} else {
			break
		}
	}
	return blockCount
}

func (tpsman *SimpleTPSManager) TPSDisplay() {
	fmt.Println("TPSDisplay")
	for {
		newblockNum := tpsman.GetBlockNum()
		// No tx is executed
		if tpsman.lastBlockNum == newblockNum {
			time.Sleep(1 * time.Second)
			continue
		}
		aveTimeInterval := time.Now().Sub(tpsman.aveStartTime)
		insTimeInterval := time.Now().Sub(tpsman.insStartTime)
		tpsman.insStartTime = time.Now()

		aveBlockExec := newblockNum - tpsman.startBlockNum + 1
		aveTPS := float64(aveBlockExec) / aveTimeInterval.Seconds()

		insBlockExec := newblockNum - tpsman.lastBlockNum + 1
		tpsman.lastBlockNum = newblockNum
		insTPS := float64(insBlockExec) / insTimeInterval.Seconds()
		if tpsman.minTPS > insTPS {
			tpsman.minTPS = insTPS
		}
		if tpsman.maxTPS < insTPS {
			tpsman.maxTPS = insTPS
		}
		fmt.Println("========================================================")
		fmt.Printf("[TPS log] StartBlock Num: %d, LastBlockNum: %d, NewBlockNum: %d\n", tpsman.startBlockNum, tpsman.lastBlockNum, newblockNum)
		fmt.Printf("[TPS log] Average BTPS: %5.2f, Time Last: %dms, Total BlockExec: %d\n", aveTPS, aveTimeInterval.Milliseconds(), aveBlockExec)
		fmt.Printf("[TPS log] Instant TPS %5.2f, Time Interval: %dms, BlockExec: %d\n", insTPS, insTimeInterval.Milliseconds(), insBlockExec)
		fmt.Printf("[Summary] Average BTPS: %5.2f, Max TPS: %5.2f, Min TPS: %5.2f, Time Last: %dms\n", aveTPS, tpsman.maxTPS, tpsman.minTPS, aveTimeInterval.Milliseconds())
		fmt.Println("========================================================")

		time.Sleep(2 * time.Second)
	}
}
