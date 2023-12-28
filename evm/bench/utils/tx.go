package utils

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	ethcmm "github.com/ethereum/go-ethereum/common"
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
)

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

	// ethClient for mempool query
	cli, err := ethclient.Dial(config.TransferCfg.Rpc[0])

	if err != nil {
		panic(fmt.Errorf("failed to initialize client: %+v", err))
	}

	go func(client *ethclient.Client) {
		for {
			size := getMempoolSize(client)
			mempoolSizeMap.Store(0, size)
			time.Sleep(500 * time.Millisecond)
		}
	}(cli)

	tpsman := NewTPSMan(config.TransferCfg.Rpc[0])

	concurrency := config.TransferCfg.Concurrency
	count := len(accounts) / concurrency
	for i := 0; i < concurrency; i++ {
		go func(gIndex int) {
			for j := 0; ; j++ {
				for index := gIndex * count; index < (gIndex+1)*count; index++ {
					acc := accounts[index]
					cli := clients[index%len(clients)]

					mempoolSize, ok := mempoolSizeMap.Load(0)
					//fmt.Printf("Mempool size: %d\n", mempoolSize)
					if ok && mempoolSize.(int) >= config.TransferCfg.Threshold {
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
			time.Sleep(1000 * time.Microsecond)
		} else {
			break
		}
	}
	return int(txcount)
}

func getGasPrice(client *ethclient.Client) *big.Int {
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

func execute(gIndex int, cli client.Client, acc *EthAccount, e func(ethcmm.Address) []TxParam) {

	acc.Lock()
	defer acc.Unlock()

	caller := common.GetEthAddressFromPK(acc.GetPrivateKey())
	if err := acc.SetNonce(cli); err != nil {
		log.Println(fmt.Errorf("[g%d] failed to query %s nonce, error: %s", gIndex, caller, err))
		return
	}

	// Query GasPrice
	var gasPrice *big.Int
	ethClient, ok := cli.(*client.EthClient)
	if ok {
		gasPrice = getGasPrice(ethClient.Client)
	}

	eParams := e(caller)

	var txhash ethcmm.Hash
	var err error

	for _, eParam := range eParams {
		if eParam.gasPrice.Cmp(gasPrice) < 0 {
			txhash, err = cli.SendEthereumTx(acc.GetPrivateKey(), acc.GetNonce(), eParam.to, eParam.amount, eParam.gasLimit, gasPrice, eParam.data)
		} else {
			txhash, err = cli.SendEthereumTx(acc.GetPrivateKey(), acc.GetNonce(), eParam.to, eParam.amount, eParam.gasLimit, eParam.gasPrice, eParam.data)
		}
		if err != nil {
			log.Printf("[g%d] %s send tx err: %s\n", gIndex, caller, err)
			if strings.Contains(err.Error(), "already exists") {
				acc.AddNonce()
			} else if strings.Contains(err.Error(), "mempool is full") {
				//time.Sleep(time.Second)
			} else if strings.Contains(err.Error(), "invalid nonce") {
				acc.AddNonce()
			}
		} else {
			log.Printf("[g%d] %s txhash: %s\n", gIndex, caller, txhash)
			acc.AddNonce()
			time.Sleep(50 * time.Millisecond)
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
