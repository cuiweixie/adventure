package core

import (
	"log"
	"time"

	"github.com/okex/adventure/common/client"
	cmwraptx "github.com/okex/adventure/common/types"
	"github.com/okex/adventure/wasm/bench/common/account"
)

type BaseBench struct {
	Accounts          []*account.Account
	Concurrency       int
	BuildTxFn         func(sender int, accounts []*account.Account, client *client.CosmosClient) (stdTx []*cmwraptx.WrapCMTx)
	TendermintClients []*client.CosmosClient
	MempoolThreshold  int
}

func (b *BaseBench) StartBench() {
	count := len(b.Accounts) / b.Concurrency
	for i := 0; i < b.Concurrency; i++ {
		go func(gIndex int) {
			for {
				for index := gIndex * count; index < (gIndex+1)*count; index++ {
					client := b.TendermintClients[gIndex%len(b.TendermintClients)]
					if b.MempoolThreshold > 0 && client.Mempoolsize >= b.MempoolThreshold {
						log.Println("mempool 达到阈值")
						time.Sleep(200 * time.Millisecond)
						continue
					}
					stdTxs := b.BuildTxFn(index, b.Accounts, client)
					execute(client, stdTxs)
				}
			}
		}(i)
	}

	select {}
}

func (b *BaseBench) StopBench() {
	//todo
}

func execute(client *client.CosmosClient, txs []*cmwraptx.WrapCMTx) {
	for i := range txs {
		for {
			_, err := client.SendCosmosTx(txs[i])
			if err == nil {
				break
			}
			log.Println(err)
			time.Sleep(1 * time.Second)
		}
	}
}
