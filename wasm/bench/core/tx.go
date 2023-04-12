package core

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/okex/exchain/app/crypto/ethsecp256k1"
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	"github.com/okex/exchain/libs/cosmos-sdk/x/auth"
	"github.com/okex/exchain/x/wasm/types"
)

func BuildWasmTx(privateKey *ecdsa.PrivateKey, accNumber, seqNumber uint64, chainId string, memo string, contractAddr string, execMsg string, sender sdk.AccAddress, amountStr string) (stdTx *auth.StdTx, err error) {
	msg, err := parseExecuteMsg(contractAddr, execMsg, sender, amountStr)
	if err != nil {
		return nil, err
	}

	tx, _, err := buildStdTx(privateKey, chainId, memo, []sdk.Msg{msg}, accNumber, seqNumber)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func buildStdTx(privateKey *ecdsa.PrivateKey, chainId string, memo string, msgs []sdk.Msg, accNumber, seqNumber uint64) (stdTx *auth.StdTx, signstr string, err error) {
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

func parseExecuteMsg(contractAddr string, execMsg string, sender sdk.AccAddress, amountStr string) (types.MsgExecuteContract, error) {
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
