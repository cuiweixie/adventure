package utils

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/okex/adventure/evm/config"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	ethcmm "github.com/ethereum/go-ethereum/common"
	"github.com/okex/adventure/common"
	"github.com/okex/adventure/common/client"
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

/**
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

/**
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

/**
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

	concurrency := config.TransferCfg.Concurrency
	count := len(accounts) / concurrency
	for i := 0; i < concurrency; i++ {
		go func(gIndex int) {
			for j := 0; ; j++ {
				//aIndex := (gIndex + j*concurrency) % len(accounts) // make sure accounts will be picked in order by round-robin
				for index :=gIndex * count; index < (gIndex+1)*count; index++ {
					acc := accounts[index]
					cli := clients[index%len(clients)]
					tendermintUrl := config.TransferCfg.TenderMint[index%len(clients)]
					if config.TransferCfg.Threshold > 0 && j%10 == 0 && getMempoolSize(tendermintUrl) >= config.TransferCfg.Threshold {
						fmt.Println("达到阈值")
						continue
					}

					execute(gIndex, cli, acc, e)
				}

			}
		}(i)
	}

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

func getMempoolSize(tendermintUrl string) int {
	var result rpcResult
	response, err := http.Get(fmt.Sprintf("%s/num_unconfirmed_txs", tendermintUrl))
	if err != nil {
		fmt.Println(err)
		return 0
	}

	bts, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	err = json.Unmarshal(bts, &result)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	//fmt.Println("mempool size :", result.Result.Total)
	total, _ := strconv.Atoi(result.Result.Total)
	return total
}

func execute(gIndex int, cli client.Client, acc *EthAccount, e func(ethcmm.Address) []TxParam) {

	acc.Lock()
	defer acc.Unlock()

	caller := common.GetEthAddressFromPK(acc.GetPrivateKey())
	if err := acc.SetNonce(cli); err != nil {
		log.Println(fmt.Errorf("[g%d] failed to query %s nonce, error: %s", gIndex, caller, err))
		return
	}

	eParams := e(caller)
	for _, eParam := range eParams {
		_, err := cli.SendEthereumTx(acc.GetPrivateKey(), acc.GetNonce(), eParam.to, eParam.amount, eParam.gasLimit, eParam.gasPrice, eParam.data)
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
			//log.Printf("[g%d] %s txhash: %s\n", gIndex, caller, txhash)
			acc.AddNonce()
		}
	}
}
