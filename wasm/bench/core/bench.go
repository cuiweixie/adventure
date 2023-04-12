package core

import (
	"fmt"
	"github.com/okex/adventure/common/client"
	"github.com/okex/adventure/wasm/bench/common/account"
	"github.com/okex/exchain/libs/cosmos-sdk/x/auth"
	"log"
	"time"
)

type BaseBench struct {
	Accounts          []*account.Account
	Concurrency       int
	BuildTxFn         func(sender int, accounts []*account.Account) (stdTx []*auth.StdTx)
	TendermintClients []*client.CosmosClient
}

func NewBaseBench() {

}

func (b *BaseBench) StartBench() {
	count := len(b.Accounts) / b.Concurrency
	for i := 0; i < b.Concurrency; i++ {
		go func(gIndex int) {
			for {
				for index := gIndex * count; index < (gIndex+1)*count; index++ {
					client := b.TendermintClients[gIndex%len(b.TendermintClients)]
					stdTxs := b.BuildTxFn(index, b.Accounts)

					execute(client, stdTxs)
				}
			}
		}(i)
	}

	select {}
}

func (b *BaseBench) StopBench() {

}

func execute(client *client.CosmosClient, txs []*auth.StdTx) {
	for i := range txs {
		for {
			hash, err := client.SendCosmosTx(txs[i])
			fmt.Println(hash)
			if err == nil {
				break
			}
			log.Println(err)
			time.Sleep(1 * time.Second)
		}
	}
}
