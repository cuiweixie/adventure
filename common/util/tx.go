package util

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	gosdk "github.com/okex/exchain-go-sdk"
	types2 "github.com/okex/exchain-go-sdk/types"
	"github.com/okex/exchain/app/crypto/ethsecp256k1"
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	"github.com/okex/exchain/libs/cosmos-sdk/x/auth"
	"github.com/okex/exchain/x/wasm/types"
)

func SendWasmTx(client *gosdk.Client, privateKey *ecdsa.PrivateKey, accNumber, seqNumber uint64, chainId string, memo string, contractAddr string, execMsg string, sender sdk.AccAddress, amountStr string) error {
	msg, err := ParseExecuteMsg(contractAddr, execMsg, sender, amountStr)
	if err != nil {
		return err
	}

	tx, _, err := BuildStdTx(privateKey, chainId, memo, []sdk.Msg{msg}, accNumber, seqNumber)
	if err != nil {
		return err
	}

	cli := client.Auth().(types2.BaseClient)
	bytes, err := cli.GetCodec().MarshalBinaryLengthPrefixed(tx)
	_, err = cli.Broadcast(bytes, "sync")
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

func ParseExecuteMsg(contractAddr string, execMsg string, sender sdk.AccAddress, amountStr string) (types.MsgExecuteContract, error) {
	amount, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return types.MsgExecuteContract{}, err
	}

	return types.MsgExecuteContract{
		Sender:   sender.String(),
		Contract: contractAddr,
		Funds:    sdk.CoinsToCoinAdapters(amount),
		Msg:      []byte(execMsg),
	}, nil
}
