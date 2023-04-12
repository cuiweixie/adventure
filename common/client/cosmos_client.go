package client

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/okex/adventure/common/util"
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	"github.com/okex/exchain/libs/cosmos-sdk/x/auth"
	"math/big"

	ethcmn "github.com/ethereum/go-ethereum/common"
	gosdk "github.com/okex/exchain-go-sdk"
	"github.com/okex/exchain-go-sdk/types"
	"github.com/okex/exchain-go-sdk/utils"
	"github.com/okex/exchain/x/common"
)

type CosmosClient struct {
	*gosdk.Client
}

func NewCosmosClient(ip string) (*CosmosClient, error) {
	chainId, err := queryChainIdFromCosmos(ip)
	if err != nil {
		return nil, err
	}

	cfg, err := types.NewClientConfig(ip, chainId, types.BroadcastSync, "", 2000000, 1.5, "0.0000000001"+common.NativeToken)
	if err != nil {
		panic(fmt.Errorf("initialize client failed: %s", err))
	}
	cli := gosdk.NewClient(cfg)

	return &CosmosClient{
		&cli,
	}, nil
}

func queryChainIdFromCosmos(ip string) (string, error) {
	tmpcCfg, err := types.NewClientConfig(ip, "unknown-1", types.BroadcastSync, "", 2000000, 1.5, "0.0000000001"+common.NativeToken)
	if err != nil {
		panic(fmt.Errorf("initialize client failed: %s", err))
	}
	tmpCli := gosdk.NewClient(tmpcCfg)

	status, err := tmpCli.Tendermint().QueryStatus()
	if err != nil {
		return "", err
	}
	chainID := status.NodeInfo.Network

	// if chainid = 65, should be resolved into exchain-65, not okexchain-65
	if chainID == "okexchain-65" {
		chainID = "exchain-65"
	}
	return chainID, nil
}

func (c *CosmosClient) QueryNonce(hexAddr string) (uint64, error) {
	cosmosAddr, err := utils.ToCosmosAddress(hexAddr)
	if err != nil {
		return 0, err
	}

	account, err := c.Auth().QueryAccount(cosmosAddr.String())
	if err != nil {
		return 0, err
	}
	return account.GetSequence(), nil
}

func (c *CosmosClient) QueryChainID() (string, error) {
	status, err := c.Client.Tendermint().QueryStatus()
	if err != nil {
		return "", err
	}
	chainID := status.NodeInfo.Network
	return chainID, nil
}
func (c *CosmosClient) SendEthereumTx(privatekey *ecdsa.PrivateKey, nonce uint64, to ethcmn.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) (ethcmn.Hash, error) {
	res, err := c.Evm().SendTxEthereum(privatekey, nonce, to, amount, gasLimit, gasPrice, data)
	if err != nil {
		return ethcmn.Hash{}, err
	}
	return ethcmn.HexToHash(res.TxHash), nil
}

func (c *CosmosClient) CreateContract(privatekey *ecdsa.PrivateKey, nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) (ethcmn.Hash, error) {
	res, err := c.Evm().CreateContractEthereum(privatekey, nonce, amount, gasLimit, gasPrice, data)
	if err != nil {
		return ethcmn.Hash{}, err
	}
	return ethcmn.HexToHash(res.TxHash), nil
}

func (c *CosmosClient) SendCosmosTx(signedTx *auth.StdTx) (string, error) {
	cli := c.Client.Auth().(types.BaseClient)
	bytes, err := cli.GetCodec().MarshalBinaryLengthPrefixed(signedTx)
	tx, err := cli.Broadcast(bytes, c.GetConfig().BroadcastMode)
	return tx.TxHash, err
}

func (c *CosmosClient) SendWasmTx(privateKey *ecdsa.PrivateKey, accNumber, seqNumber uint64, chainId string, memo string, contractAddr string, execMsg string, sender sdk.AccAddress, amountStr string) (string, error) {
	msg, err := util.ParseExecuteMsg(contractAddr, execMsg, sender, amountStr)
	if err != nil {
		return "", err
	}

	signedTx, _, err := util.BuildStdTx(privateKey, chainId, memo, []sdk.Msg{msg}, accNumber, seqNumber)
	if err != nil {
		return "", err
	}

	cli := c.Client.Auth().(types.BaseClient)
	bytes, err := cli.GetCodec().MarshalBinaryLengthPrefixed(signedTx)
	tx, err := cli.Broadcast(bytes, c.GetConfig().BroadcastMode)
	return tx.TxHash, err
}
