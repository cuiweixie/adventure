package utils

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	ethcmm "github.com/ethereum/go-ethereum/common"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/okex/adventure/common"
	"github.com/okex/adventure/common/client"
	"github.com/okex/adventure/evm/config"
)

type TxParam struct {
	to       ethcmm.Address
	amount   *big.Int
	gasLimit uint64
	gasPrice *big.Int
	data     []byte
}

var (
	lstTxHash    = make([]string, 0)
	duration     int64
	ratio        float32
	tps          int64
	lstRlpEncode = make([]string, 0)
	chainId      = new(big.Int).SetUint64(65)
	signer       = types.NewLondonSigner(chainId)

	// chainId缓存 - 所有节点共享同一个chainId
	cachedChainId *big.Int
	chainIdOnce   sync.Once
)

// 获取缓存的chainId - 只查询一次
func getCachedChainId(ethClient *client.EthClient) (*big.Int, error) {
	var err error
	chainIdOnce.Do(func() {
		cachedChainId, err = ethClient.ChainID(context.Background())
	})
	return cachedChainId, err
}

// Default GasPrice for X1 is set to 10GWei/gas
func ParseGasPriceToBigInt(gasPriceFloat float64, prec int) *big.Int {
	mul, err := strconv.ParseFloat(fmt.Sprintf(`1%0`+strconv.Itoa(prec)+`s`, ""), 64)
	if err != nil {
		return new(big.Int).SetUint64(10000000000)
	}
	gasPriceWeiFloat := gasPriceFloat * mul
	if hasDecimal(gasPriceWeiFloat) {
		return new(big.Int).SetUint64(10000000000)
	}
	return new(big.Int).SetUint64(uint64(gasPriceWeiFloat))
}

func hasDecimal(num float64) bool {
	intPart := math.Floor(num)
	return intPart != num
}

/*
*
作用：用来计算并发携程一次发送完毕后的的成功率
*/
func GetTxTpsAndSuccessRatio(lstTxHash []string, cocurrent int64) (ratio float32, tps int64) {
	num := len(lstTxHash)
	ratio = float32(num) / float32(cocurrent)
	tps = int64(num*1000) / duration
	return
}

func getTxHashList(gIndex int, cli client.Client, acc *EthAccount, e func(ethcmm.Address) []TxParam) []string {
	acc.Lock()
	defer acc.Unlock()

	caller := common.GetEthAddressFromPK(acc.GetPrivateKey())
	if err := acc.SetNonce(cli); err != nil {
		log.Println(fmt.Errorf("[g%d] failed to query %s nonce, error: %s", gIndex, caller, err))
		return lstTxHash
	}

	eParams := e(caller)
	for _, eParam := range eParams {
		txhash, err := cli.SendEthereumTx(acc.GetPrivateKey(), acc.GetNonce(), eParam.to, eParam.amount, eParam.gasLimit, eParam.gasPrice, eParam.data)
		if err != nil {
			log.Printf("[g%d] %s send tx err: %s\n", gIndex, caller, err)
			if strings.Contains(err.Error(), "already exists") {
				acc.AddNonce()
			} else if strings.Contains(err.Error(), "mempool is full") {
				time.Sleep(time.Second)
			} else if strings.Contains(err.Error(), "invalid nonce") {
				acc.AddNonce()
			}
		} else {
			log.Printf("[g%d] %s txhash: %s\n", gIndex, caller, txhash)
			lstTxHash = append(lstTxHash, txhash.String())
			acc.AddNonce()
		}
	}
	return lstTxHash
}

/*
*
功能：获取返回所有账户的rlpencode
*/
func getTxRlpEncodeList(cli client.Client, acc *EthAccount, e func(ethcmm.Address) []TxParam) {
	caller := common.GetEthAddressFromPK(acc.GetPrivateKey())
	if err := acc.SetNonce(cli); err != nil {
		log.Println(err)
	}

	eParams := e(caller)
	for _, eParam := range eParams {
		rlpencode, err := GetEthTxRlpEncode(acc.GetPrivateKey(), acc.GetNonce(), eParam.to, eParam.amount, eParam.gasLimit, eParam.gasPrice, eParam.data)
		if err != nil {
			log.Println(err)
		} else {
			lstRlpEncode = append(lstRlpEncode, rlpencode)
		}
	}
	//return lstRlpEncode
}

/*
*
功能：获取到单个交易的rlpencode
*/
func GetEthTxRlpEncode(pk *ecdsa.PrivateKey, nonce uint64, to ethcmm.Address, amount *big.Int, gaslimit uint64, gasprice *big.Int, data []byte) (string, error) {
	//make tx
	unsignedTx := types.NewTransaction(nonce, to, amount, gaslimit, gasprice, data)

	//sign tx
	signedTx, err := types.SignTx(unsignedTx, signer, pk)
	if err != nil {
		log.Println(err)
	}
	//当需要调用 eth_sendRawTransaction 函数中的 params的时候，通过下面这个rlp来构造
	b, err := rlp.EncodeToBytes(signedTx)
	params := "0x" + hex.EncodeToString(b)
	log.Printf("%s\n", params)
	return params, nil
}

func RunTxGetRlpEncodeList(p BasepParam, e func(ethcmm.Address) []TxParam) {
	clients := client.GenerateClients(p.ips)    // generate CosmosClient or EthClient
	accounts := generateAccounts(p.privateKeys) // generate accounts

	for j := 0; j < len(accounts); j++ {
		acc := accounts[j]
		cli := clients[0]
		getTxRlpEncodeList(cli, acc, e)
	}

	//for i :=0; i<len(lstRlpEncode); i++{
	//	log.Printf("%s\n", lstRlpEncode[i])
	//}
}

func NewTxParam(to ethcmm.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) TxParam {
	return TxParam{
		to,
		amount,
		gasLimit,
		gasPrice,
		data,
	}
}

/**
功能：获取同时并发的交易，收到tx时候花费的总时间，并统计成功率和tps
*/

func RunTxRpc(p BasepParam, e func(ethcmm.Address) []TxParam) {
	clients := client.GenerateClients(p.ips)    // generate CosmosClient or EthClient
	accounts := generateAccounts(p.privateKeys) // generate accounts

	startTime := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < p.concurrency; i++ {
		wg.Add(1)
		go func(gIndex int) {
			//j<1是为了获取一次交易
			for j := 0; j < 1; j++ {
				aIndex := (gIndex + j*p.concurrency) % len(accounts) // make sure accounts will be picked in order by round-robin
				acc := accounts[aIndex]
				cli := clients[aIndex%len(clients)]

				getTxHashList(gIndex, cli, acc, e)
				//time.Sleep(time.Millisecond * time.Duration(p.sleep))
			}
			defer wg.Done()
		}(i)
	}
	wg.Wait()
	duration = time.Since(startTime).Milliseconds()
	elapsed := strconv.FormatInt(time.Since(startTime).Milliseconds(), 10) + "ms"
	ratio, tps = GetTxTpsAndSuccessRatio(lstTxHash, int64(p.concurrency))
	log.Printf(" %d tx sent and received txhash and total time cost: %s\n", p.concurrency, elapsed)
	log.Printf(" %d tx send success, and sucess ratio is : %d, and tx tps is : %d\n", len(lstTxHash), ratio, tps)
}

func RunTxs(p BasepParam, e func(ethcmm.Address) []TxParam) {
	clients := client.GenerateClients(config.TransferCfg.Rpc)    // generate CosmosClient or EthClient
	accounts := generateAccounts(config.TransferCfg.PrivateKeys) // generate accounts
	mempoolSizeMap := &sync.Map{}
	mempoolSizeMap.Store(0, 0)

	go func() {
		for {
			size := getMempoolSizeV2(config.TransferCfg.Rpc[0])
			mempoolSizeMap.Store(0, size)
			time.Sleep(time.Second)
		}
	}()

	tpsman := NewTPSMan(config.TransferCfg.Rpc[0])

	concurrency := config.TransferCfg.Concurrency
	count := len(accounts) / concurrency
	for i := 0; i < concurrency; i++ {
		go func(gIndex int) {
			for {
				mempoolSize, ok := mempoolSizeMap.Load(0)
				//fmt.Printf("Mempool size: %d\n", mempoolSize)
				if ok && mempoolSize.(int) >= config.TransferCfg.Threshold {
					fmt.Println("达到阈值:", mempoolSize.(int))
					time.Sleep(time.Second)
					continue
				}

				start := gIndex * count
				end := start + count
				if end > len(accounts) {
					end = len(accounts)
				}
				batchAccounts := accounts[start:end]

				// 使用批量执行
				cli := clients[gIndex%len(clients)]
				executeBatch(gIndex, cli, batchAccounts, e)
				time.Sleep(time.Millisecond * 10) // 批量发送后稍微休息长一点

			}
		}(i)
	}

	go tpsman.TPSDisplay()
	select {}
}

type rpcResult struct {
	Result MempoolResult `json:"result"`
}
type MempoolResult struct {
	Txs        string `json:"n_txs"`
	Total      string `json:"total"`
	TotalBytes string `json:"total_bytes"`
}

func getMempoolSize(client *ethclient.Client) int {
	var txcount uint
	var err error

	for {
		txcount, err = client.PendingTransactionCount(context.Background())
		if err != nil {
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return int(txcount)
}

func getGasPrice(client *ethclient.Client) *big.Int {
	return big.NewInt(1)
	var gp *big.Int
	var err error
	var incAmount = new(big.Int).SetUint64(10000000000)

	for {
		gp, err = client.SuggestGasPrice(context.Background())
		if err != nil {
			time.Sleep(10 * time.Microsecond)
		} else {
			break
		}
	}
	gp.Add(gp, incAmount)
	return gp
}

var defaultGasPrice = big.NewInt(1)

// 全局HTTP客户端，用于复用连接
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:          300,              // 增加全局最大空闲连接数
		MaxIdleConnsPerHost:   300,              // 增加每个主机的最大空闲连接数
		MaxConnsPerHost:       300,              // 限制每个主机的最大连接数
		IdleConnTimeout:       90 * time.Second, // 空闲连接超时时间
		TLSHandshakeTimeout:   10 * time.Second, // TLS握手超时
		ExpectContinueTimeout: 1 * time.Second,  // Expect: 100-continue超时
		DisableKeepAlives:     false,            // 启用Keep-Alive（默认就是false，明确设置）
		DisableCompression:    false,            // 启用压缩
		ForceAttemptHTTP2:     false,            // 对于RPC调用，HTTP/1.1就够了
	},
}

type TxPoolStatus struct {
	BaseFee string `json:"baseFee"`
	Pending string `json:"pending"`
	Queued  string `json:"queued"`
}

type TxPoolResponse struct {
	JsonRPC string       `json:"jsonrpc"`
	ID      int          `json:"id"`
	Result  TxPoolStatus `json:"result"`
}

// 新的getMempoolSize方法，使用txpool_status接口
func getMempoolSizeV2(rpcURL string) int {
	// 构造JSON-RPC请求
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "txpool_status",
		"params":  []interface{}{},
		"id":      1,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Failed to marshal request: %v\n", err)
		return 0
	}

	// 使用复用的HTTP客户端发送请求
	resp, err := httpClient.Post(rpcURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send request: %v\n", err)
		return 0
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response: %v\n", err)
		return 0
	}

	// 解析响应
	var response TxPoolResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Failed to unmarshal response: %v\n", err)
		return 0
	}

	// 解析十六进制字符串并计算总数
	baseFee := hexToInt(response.Result.BaseFee)
	pending := hexToInt(response.Result.Pending)
	queued := hexToInt(response.Result.Queued)

	return baseFee + pending + queued
}

func hexToInt(hexStr string) int {
	if hexStr == "" || hexStr == "0x" {
		return 0
	}

	// 去掉0x前缀
	if strings.HasPrefix(hexStr, "0x") {
		hexStr = hexStr[2:]
	}

	// 解析十六进制
	val, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		log.Printf("Failed to parse hex string %s: %v\n", hexStr, err)
		return 0
	}

	return int(val)
}

// 批量执行多个账户的交易
func executeBatch(gIndex int, cli client.Client, accounts []*EthAccount, e func(ethcmm.Address) []TxParam) {
	const maxBatchSize = 100

	// 检查是否支持批量发送
	ethClient, ok := cli.(*client.EthClient)
	if !ok {
		panic("eth client is not a eth client")
	}

	// 获取交易参数模板（只用第一个账户的第一个交易作为模板）
	if len(accounts) == 0 {
		return
	}

	eParams := e(accounts[0].caller)
	txTemplate := eParams[0] // 永远只有1个，所有交易都用相同的参数

	// 计算总交易数
	totalTxs := len(accounts)

	// 分批发送，每批最多100笔
	for i := 0; i < totalTxs; i += maxBatchSize {
		end := i + maxBatchSize
		if end > totalTxs {
			end = totalTxs
		}

		startAccountIndex := i
		endAccountIndex := end
		if endAccountIndex > len(accounts) {
			endAccountIndex = len(accounts)
		}

		sendSimpleBatch(gIndex, ethClient, txTemplate, accounts[startAccountIndex:endAccountIndex])

		time.Sleep(time.Millisecond * 100)
	}
}

func sendSimpleBatch(gIndex int, ethClient *client.EthClient, txTemplate TxParam, accounts []*EthAccount) {
	if len(accounts) == 0 {
		return
	}

	// 构造并签名所有交易
	var signedTxs []*types.Transaction
	chainId, err := getCachedChainId(ethClient)
	if err != nil {
		log.Printf("[g%d] failed to get chainId: %v\n", gIndex, err)
		return
	}
	signer := types.NewLondonSigner(chainId)

	for _, acc := range accounts {
		acc.Lock()
		if err := acc.SetNonce(ethClient); err != nil {
			log.Printf("[g%d] failed to query nonce: %s\n", gIndex, err)
			acc.Unlock()
			continue
		}

		// 创建交易
		unsignedTx := types.NewTransaction(
			acc.GetNonce(),
			txTemplate.to,
			txTemplate.amount,
			txTemplate.gasLimit,
			defaultGasPrice,
			txTemplate.data,
		)

		// 签名交易
		signedTx, err := types.SignTx(unsignedTx, signer, acc.GetPrivateKey())
		if err != nil {
			log.Printf("[g%d] failed to sign tx: %v\n", gIndex, err)
			acc.Unlock()
			continue
		}

		signedTxs = append(signedTxs, signedTx)
		acc.Unlock()
	}

	if len(signedTxs) == 0 {
		return
	}

	// 使用批量发送接口
	txHashes, err := ethClient.SendMultipleEthereumTx(signedTxs)
	if err != nil {
		log.Printf("[g%d] batch send failed: %v\n", gIndex, err)
		return
	}

	// 统计成功的交易并更新nonce
	successCount := 0
	for i, txHash := range txHashes {
		if txHash != (ethcmn.Hash{}) {
			// 发送成功，更新对应账户的nonce
			accounts[i].AddNonce()
			successCount++
		}
	}

	if successCount != len(signedTxs) {
		log.Printf("[g%d] batch sent %d/%d transactions successfully\n", gIndex, successCount, len(signedTxs))
	}
}

func execute(gIndex int, cli client.Client, acc *EthAccount, e func(ethcmm.Address) []TxParam) {

	acc.Lock()
	defer acc.Unlock()

	caller := acc.caller
	if err := acc.SetNonce(cli); err != nil {
		log.Println(fmt.Errorf("[g%d] failed to query %s nonce, error: %s", gIndex, caller, err))
		return
	}

	eParams := e(caller)
	if len(eParams) == 0 {
		return
	}

	// 单个发送
	for _, eParam := range eParams {
		_, err := cli.SendEthereumTx(acc.GetPrivateKey(), acc.GetNonce(), eParam.to, eParam.amount, eParam.gasLimit, defaultGasPrice, eParam.data)
		if err == nil {
			acc.AddNonce()
		} else {
			log.Printf("[g%d] %s send tx err: %s\n", gIndex, caller, err)
		}
	}
}

func RunTxsForPoly(e func(ethcmm.Address) []TxParam) {
	clients := client.GenerateClients(config.Bridgecfg.L1RPC)  // generate CosmosClient or EthClient
	accounts := generateAccounts(config.Bridgecfg.PrivateKeys) // generate accounts
	mempoolSizeMap := &sync.Map{}

	// ethClient for mempool query
	l1cli, err := ethclient.Dial(config.Bridgecfg.L1RPC[0])
	l2cli, err := ethclient.Dial(config.Bridgecfg.L2RPC[0])

	if err != nil {
		panic(fmt.Errorf("failed to initialize client: %+v", err))
	}

	go func(client1, client2 *ethclient.Client) {
		for {
			l1size := getMempoolSize(client1)
			l2size := getMempoolSize(client2)
			fmt.Printf("L1 Mempool size: %d, L2 Mempool size: %d\n", l1size, l2size)
			mempoolSizeMap.Store(0, l1size)
			time.Sleep(500 * time.Millisecond)
		}
	}(l1cli, l2cli)

	tpsman := NewTPSMan(config.Bridgecfg.L2RPC[0])

	concurrency := config.Bridgecfg.Concurrency
	count := len(accounts) / concurrency
	for i := 0; i < concurrency; i++ {
		go func(gIndex int) {
			for j := 0; ; j++ {
				for index := gIndex * count; index < (gIndex+1)*count; index++ {
					acc := accounts[index]
					cli := clients[index%len(clients)]

					mempoolSize, ok := mempoolSizeMap.Load(0)
					if ok && mempoolSize.(int) >= config.Bridgecfg.Threshold {
						fmt.Println("达到阈值")
						continue
					}
					execute(gIndex, cli, acc, e)
				}

			}
		}(i)
	}

	go tpsman.TPSDisplay()
	select {}
}
