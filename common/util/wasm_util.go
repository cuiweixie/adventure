package util

import (
	"fmt"
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	"github.com/okex/exchain/x/wasm/ioutils"
	"github.com/okex/exchain/x/wasm/types"
	"io/ioutil"
)

func ParseStoreCodeMsg(wasmFilePath string, sender sdk.AccAddress, onlyAddrStr string, everybody, nobody bool) (types.MsgStoreCode, error) {
	wasm, err := ioutil.ReadFile(wasmFilePath)
	if err != nil {
		return types.MsgStoreCode{}, err
	}

	// gzip the wasm file
	if ioutils.IsWasm(wasm) {
		wasm, err = ioutils.GzipIt(wasm)

		if err != nil {
			return types.MsgStoreCode{}, err
		}
	} else if !ioutils.IsGzip(wasm) {
		return types.MsgStoreCode{}, fmt.Errorf("invalid input file. Use wasm binary or gzip")
	}

	var perm *types.AccessConfig

	if onlyAddrStr != "" {
		allowedAddr, err := sdk.AccAddressFromBech32(onlyAddrStr)
		if err != nil {
			return types.MsgStoreCode{}, err
		}
		x := types.AccessTypeOnlyAddress.With(allowedAddr)
		perm = &x
	} else if everybody {
		perm = &types.AllowEverybody
	} else if nobody {
		perm = &types.AllowNobody
	}

	msg := types.MsgStoreCode{
		Sender:                sender.String(),
		WASMByteCode:          wasm,
		InstantiatePermission: perm,
	}
	return msg, nil
}

func ParseInstantiateMsg(codeID uint64, initMsg string, sender sdk.AccAddress, amountStr, label, adminStr string, noAdmin bool) (types.MsgInstantiateContract, error) {
	amount, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return types.MsgInstantiateContract{}, fmt.Errorf("amount: %s", err)
	}

	if label == "" {
		return types.MsgInstantiateContract{}, fmt.Errorf("label is required on all contracts")
	}

	// ensure sensible admin is set (or explicitly immutable)
	if adminStr == "" && !noAdmin {
		return types.MsgInstantiateContract{}, fmt.Errorf("you must set an admin or explicitly pass no-admin to make it immutible (wasmd issue #719)")
	}
	if adminStr != "" && noAdmin {
		return types.MsgInstantiateContract{}, fmt.Errorf("you set an admin and passed no-admin, those cannot both be true")
	}

	// build and sign the transaction, then broadcast to Tendermint
	msg := types.MsgInstantiateContract{
		Sender: sender.String(),
		CodeID: codeID,
		Label:  label,
		Funds:  sdk.CoinsToCoinAdapters(amount),
		Msg:    []byte(initMsg),
		Admin:  adminStr,
	}
	return msg, nil
}
