package transfer_cw20

import (
	"fmt"
	gosdk "github.com/okex/exchain-go-sdk"
	"strings"
	"time"
)

func SendWasmTx(client *gosdk.Client, tx WasmTx) error {
	cnt := 0
	var err error
	for cnt < 10 {
		cnt++
		if _, err := client.Wasm().ExecuteContract(tx.FromInfo, tx.PassWd, tx.AccNum, tx.SeqNum, tx.Memo, tx.ContractAddr, tx.ExecMsg, tx.Amount); err != nil {
			time.Sleep(time.Second * 2)
			if strings.Contains(err.Error(), "mempool is full") || strings.Contains(err.Error(), "number of txs") {
				time.Sleep(20 * time.Second)
			} else if strings.Contains(err.Error(), "failed to replace tx for acccount") {
				err = nil
			} else if strings.Contains(err.Error(), "invalid sequence") {
				time.Sleep(2 * time.Second)
			} else {
				fmt.Println("sendTransaction failed", err, "tryCnt", cnt)
			}
		} else {
			err = nil
			break
		}

	}

	return err
}
