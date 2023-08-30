package multiwmt

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func getPrivateKey(key string) *ecdsa.PrivateKey {
	privateKey, err := crypto.HexToECDSA(key)
	panicerr(err)
	return privateKey
}

func transferOkt(key string, to common.Address, nonce uint64, value *big.Int) *types.Transaction {
	privateKey := getPrivateKey(key)

	tx, err := types.SignTx(types.NewTransaction(nonce, to, value, gasLimit, gasPrice, nil), signer, privateKey)
	panicerr(err)
	return tx
}

func SignTxWithNonce(privateKey *ecdsa.PrivateKey, to common.Address, payLoad []byte, nonce uint64) *types.Transaction {
	tx, err := types.SignTx(types.NewTransaction(nonce, to, new(big.Int), gasLimit, gasPrice, payLoad), signer, privateKey)
	panicerr(err)
	return tx
}

func panicerr(err error) {
	if err != nil {
		panic(err)
	}
}

type SwapContract struct {
	Token0          common.Address
	Token1          common.Address
	Token2          common.Address
	Token3          common.Address
	FakewethAddress common.Address
	Factory         common.Address
	Router          common.Address
	Lp1             common.Address
	Lp2             common.Address
	StakingRewards1 common.Address
	StakingRewards2 common.Address
}

func LoadContractList(file string) []SwapContract {
	data, err := ioutil.ReadFile(file)
	panicerr(err)
	cList := make([]SwapContract, 0)
	err = json.Unmarshal(data, &cList)
	panicerr(err)
	return cList
}

func keyToAcc(key string) *acc {
	privateKey := getPrivateKey(key)
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("keyToAcc")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	return &acc{
		privateKey: key,
		ecdsaPriv:  privateKey,
		ethAddress: fromAddress,
	}
}

func SendTxs(client *okcClient, txs []*types.Transaction) error {
	for index, v := range txs {
		//time.Sleep(200 * time.Microsecond)
		cnt := 0
		var err error
		for cnt < 10 {
			cnt++
			if e := client.SendTransaction(context.Background(), v); e != nil {
				err = e
				time.Sleep(time.Second * 2)
				if strings.Contains(e.Error(), "mempool is full") || strings.Contains(e.Error(), "number of txs") {
					time.Sleep(20 * time.Second)
				} else if strings.Contains(e.Error(), "failed to replace tx for acccount") {
					err = nil
				} else if strings.Contains(e.Error(), "invalid sequence") {
					time.Sleep(2 * time.Second)
				} else {
					fmt.Println("sendTransaction failed", index, e, "tryCnt", cnt)
				}
			} else {
				err = nil
				break
			}

		}
		if err != nil {
			return err
		}

		if index != 0 && index%200 == 0 {
			fmt.Println("send tx index", index)
		}
	}
	//time.Sleep(2 * time.Second)
	return nil
}

type wmtConfig struct {
	RPC             []string
	ContractPath    string
	SuperAcc        string
	WorkerPath      string
	ParaNum         int
	SendOKTToWorker bool
	Threshold       int
}

func loadWMTConfig(file string) *wmtConfig {
	data, err := ioutil.ReadFile(file)
	panicerr(err)
	c := new(wmtConfig)
	err = json.Unmarshal(data, c)
	panicerr(err)
	return c
}
