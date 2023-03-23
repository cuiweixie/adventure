package transfer_cw20

import "github.com/okex/exchain/libs/cosmos-sdk/crypto/keys"

var passWd = "12345678"

type WasmTx struct {
	FromInfo     keys.Info
	PassWd       string
	AccNum       uint64
	SeqNum       uint64
	Memo         string
	ContractAddr string
	ExecMsg      string
	Amount       string
}
