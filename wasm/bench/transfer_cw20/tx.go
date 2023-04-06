package transfer_cw20

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	gosdk "github.com/okex/exchain-go-sdk"
	types2 "github.com/okex/exchain-go-sdk/types"
	"github.com/okex/exchain/app/crypto/ethsecp256k1"
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	"github.com/okex/exchain/libs/cosmos-sdk/x/auth"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/okex/adventure/common/client"
)

var chainId string

func RunTxs(e func() sdk.AccAddress) {
	clients := client.GenerateGoSDKClients(transferOption.TendermintUrls) // generate CosmosClient or EthClient
	mempoolSizeMap := &sync.Map{}

	chainId = transferOption.ChainId
	for _, tendermint := range transferOption.TendermintUrls {
		go func(url string) {
			for {
				size := getMempoolSize(url)
				mempoolSizeMap.Store(url, size)
				time.Sleep(500 * time.Millisecond)
			}
		}(tendermint)
	}

	ch := make(chan int, 10240)

	go func() {
		for i := range accounts {
			ch <- i
		}
	}()

	wg := &sync.WaitGroup{}
	for i := 0; i < len(clients); i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			for k := range ch {
				accounts[k].SetNonce(clients[index])
				if k%200 == 0 {
					log.Println("init account size: ", i)
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("init account completed")

	concurrency := transferOption.ConcurrentNum
	count := len(accounts) / concurrency

	for i := 0; i < concurrency; i++ {
		go func(gIndex int) {
			for j := 0; ; j++ {
				for index := gIndex * count; index < (gIndex+1)*count; index++ {
					acc := accounts[index]
					cli := clients[index%len(clients)]
					tendermintUrl := transferOption.TendermintUrls[index%len(clients)]
					mempoolSize, ok := mempoolSizeMap.Load(tendermintUrl)

					if ok && mempoolSize.(int) >= transferOption.Threshold {
						fmt.Println("达到阈值")
						continue
					}

					execute(gIndex, cli, acc, e, transferOption.CW20Address)
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

	total, _ := strconv.Atoi(result.Result.Total)
	return total
}

func execute(gIndex int, client *gosdk.Client, acc *EthAccount, generateRandAddress func() sdk.AccAddress, contractAddress string) {
	caller := acc.GetPrivateKey()

	toAddress := generateRandAddress()
	execMsg := fmt.Sprintf(`{"transfer": {"amount":"1", "recipient":"%s"}}`, toAddress.Bech32StringOptimized("ex"))

	ethAddr := acc.GetHexAddress()
	bench32Addr := sdk.AccAddress(ethAddr.Bytes())

	err := sendWasmTx(client, acc.privateKey, acc.GetAccountNumber(), acc.GetNonce(), chainId, "", contractAddress, execMsg, bench32Addr, "")
	if err != nil {
		log.Printf("[g%d] %s send tx err: %s\n", gIndex, caller, err)
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println(err.Error())
		} else if strings.Contains(err.Error(), "mempool is full") {
			time.Sleep(time.Second)
		} else if strings.Contains(err.Error(), "invalid nonce") {

		}
	} else {
		acc.AddNonce()
	}
}

func sendWasmTx(client *gosdk.Client, privateKey *ecdsa.PrivateKey, accNumber, seqNumber uint64, chainId string, memo string, contractAddr string, execMsg string, sender sdk.AccAddress, amountStr string) error {
	msg, err := parseExecuteMsg(contractAddr, execMsg, sender, amountStr)
	if err != nil {
		return err
	}

	tx, _, err := BuildStdTx(privateKey, chainId, memo, []sdk.Msg{msg}, accNumber, seqNumber)
	if err != nil {
		return err
	}

	cli := client.Auth().(types2.BaseClient)
	bytes, err := cli.GetCodec().MarshalBinaryLengthPrefixed(tx)
	t, err := cli.Broadcast(bytes, "sync")
	fmt.Println(t.TxHash)
	return err
}

func BuildStdTx(privateKey *ecdsa.PrivateKey, chainId string, memo string, msgs []sdk.Msg, accNumber, seqNumber uint64) (stdTx *auth.StdTx, signstr string, err error) {
	stdFee := auth.NewStdFee(3000000, sdk.NewCoins(sdk.NewCoin("okt", sdk.NewDecWithPrec(3, 3))))
	signMsg := auth.StdSignMsg{
		ChainID:       chainId,
		AccountNumber: accNumber,
		Sequence:      seqNumber,
		Memo:          memo,
		Msgs:          msgs,
		Fee:           stdFee,
	}

	sigBytes, err := crypto.Sign(crypto.Keccak256Hash(signMsg.Bytes()).Bytes(), privateKey)
	if err != nil {
		return
	}

	signature := auth.StdSignature{
		PubKey:    ethsecp256k1.PubKey(crypto.CompressPubkey(&privateKey.PublicKey)),
		Signature: sigBytes,
	}

	return auth.NewStdTx(signMsg.Msgs, signMsg.Fee, []auth.StdSignature{signature}, signMsg.Memo), string(signMsg.Bytes()), err
}
