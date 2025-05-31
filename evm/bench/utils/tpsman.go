package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/core/types"

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

	url string
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
		url:           clientURL,
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

func (tpsman *SimpleTPSManager) TPSDisplay() {
	time.Sleep(time.Second * 10)
	fmt.Println("TPSDisplay")
	var initHeight uint64
	var totalTxCount uint64
	var initTime time.Time
	for {
		height := tpsman.GetBlockNum()
		txCount, err := tpsman.transactionCountByHeight(height)
		if err != nil {
			panic(err)
		}
		// skip this block
		if txCount > 0 {
			initHeight = height
			initTime = time.Now()
			break
		} else {
			fmt.Println("height", height, "txcount", txCount)
			time.Sleep(time.Second)
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
			txCount, err := tpsman.transactionCountByHeight(height)
			if err != nil {
				panic(err)
			}
			totalTxCount += txCount
			lastHeight = height

			avgTPS = float64(totalTxCount) / float64(time.Since(initTime).Seconds())
			if avgTPS > maxTps {
				maxTps = avgTPS
			}
			if avgTPS < minTps {
				minTps = avgTPS
			}
		}
		fmt.Println("========================================================")
		fmt.Printf("[TPS log] StartBlock Num: %d, NewBlockNum: %d, totalTxCount:%d\n", initHeight+1, lastHeight, totalTxCount)
		fmt.Printf("[Summary] Average BTPS: %5.2f, Max TPS: %5.2f, Min TPS: %5.2f, Time Last: %ds\n", avgTPS, maxTps, minTps, int64(time.Since(initTime).Seconds()))
		fmt.Println("========================================================")

		time.Sleep(5 * time.Second)
	}

}

func (tpsman *SimpleTPSManager) transactionCountByHeight(height uint64) (uint64, error) {
	// 构造 JSON-RPC 请求
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{fmt.Sprintf("0x%x", height), false}, // false 表示不返回完整交易详情
		"id":      1,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return 0, fmt.Errorf("error marshaling JSON request: %v", err)
	}

	// 发送 HTTP 请求
	resp, err := http.Post(tpsman.url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %v", err)
	}

	// 解析 JSON-RPC 响应
	var rpcResponse struct {
		Result *struct {
			Transactions []string `json:"transactions"` // 当 false 时，这里是交易哈希数组
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &rpcResponse); err != nil {
		return 0, fmt.Errorf("error unmarshaling JSON response: %v", err)
	}

	// 检查是否有错误
	if rpcResponse.Error != nil {
		return 0, fmt.Errorf("JSON-RPC error: %s", rpcResponse.Error.Message)
	}

	// 检查结果是否为空
	if rpcResponse.Result == nil {
		return 0, nil
	}

	// 返回交易数量
	return uint64(len(rpcResponse.Result.Transactions)), nil
}
