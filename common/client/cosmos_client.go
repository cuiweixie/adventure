package client

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	"github.com/okex/exchain/libs/cosmos-sdk/types/errors"

	"github.com/okex/adventure/common/util"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	gosdk "github.com/okex/exchain-go-sdk"
	"github.com/okex/exchain-go-sdk/types"
	"github.com/okex/exchain-go-sdk/utils"
	"github.com/okex/exchain/x/common"

	cmwraptx "github.com/okex/adventure/common/types"
)

type CosmosClient struct {
	*gosdk.Client
	Mempoolsize int
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

	c := &CosmosClient{
		&cli,
		0,
	}

	go func() {
		for {
			c.Mempoolsize = c.getMempoolSize()
			time.Sleep(time.Second)
		}
	}()

	return c, nil
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

func (c *CosmosClient) getMempoolSize() int {
	var result rpcResult
	response, err := http.Get(fmt.Sprintf("%s/num_unconfirmed_txs", c.Client.GetConfig().NodeURI))
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

	fmt.Println("mempool size :", result.Result.Total)
	total, _ := strconv.Atoi(result.Result.Total)
	return total
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

// 批量发送已签名的交易 - CosmosClient暂不支持，返回空实现
func (c *CosmosClient) SendMultipleEthereumTx(signedTxs []*ethtypes.Transaction) ([]ethcmn.Hash, error) {
	return nil, fmt.Errorf("multiple transaction not supported for CosmosClient")
}

func (c *CosmosClient) SendCosmosTx(signedTx *cmwraptx.WrapCMTx) (string, error) {
	cli := c.Client.Auth().(types.BaseClient)
	txBytes, err := cli.GetCodec().MarshalJSON(signedTx)
	if err != nil {
		return "", errors.Wrap(err, "MarshalJSON fail")
	}

	tx, err := cli.Broadcast(txBytes, "sync")
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

func (c *CosmosClient) DeployContract(privateKey *ecdsa.PrivateKey, accNumber, seqNumber uint64, wasmFile string, initMsg string, sender sdk.AccAddress) (string, error) {

	// store code
	codeId, err := c.StoreCode(privateKey, accNumber, seqNumber, wasmFile, sender)
	if err != nil {
		return "", err
	}

	seqNumber++
	// instantiate code
	return c.InstantiateContract(privateKey, accNumber, seqNumber, codeId, initMsg, sender, "", "deploy", sender.String())
}

func (c *CosmosClient) StoreCode(privateKey *ecdsa.PrivateKey, accNumber, seqNumber uint64, wasmFile string, sender sdk.AccAddress) (uint64, error) {
	msg, err := util.ParseStoreCodeMsg(wasmFile, sender, "", true, false)
	if err != nil {
		return 0, errors.Wrapf(err, "parse StoreCodeMsg failed")
	}

	if err = msg.ValidateBasic(); err != nil {
		return 0, err
	}

	chainID, err := c.QueryChainID()
	if err != nil {
		return 0, err
	}

	signedTx, _, err := util.BuildStdTx(privateKey, chainID, "store", []sdk.Msg{msg}, accNumber, seqNumber)
	if err != nil {
		return 0, err
	}

	cli := c.Client.Auth().(types.BaseClient)
	txBytes, err := cli.GetCodec().MarshalBinaryLengthPrefixed(signedTx)
	if err != nil {
		return 0, errors.Wrap(err, "MarshalJSON fail")
	}

	wrapTx := cmwraptx.WrapCMTx{
		txBytes,
		seqNumber,
	}

	wtxBytes, err := cli.GetCodec().MarshalJSON(wrapTx)
	if err != nil {
		return 0, errors.Wrap(err, "MarshalJSON fail")
	}

	tx, err := cli.Broadcast(wtxBytes, "block")
	if err != nil {
		return 0, err
	}

	return parseCodeID(tx.RawLog), nil
}

func (c CosmosClient) InstantiateContract(privateKey *ecdsa.PrivateKey, accNumber, seqNumber uint64, codeID uint64, initMsg string, sender sdk.AccAddress, amount, label string, admin string) (string, error) {
	msg, err := util.ParseInstantiateMsg(codeID, initMsg, sender, amount, label, admin, false)
	if err != nil {
		return "", errors.Wrapf(err, "parse InstantiateContractMsg failed")
	}

	if err = msg.ValidateBasic(); err != nil {
		return "", err
	}

	chainID, err := c.QueryChainID()
	if err != nil {
		return "", err
	}

	signedTx, _, err := util.BuildStdTx(privateKey, chainID, "store", []sdk.Msg{msg}, accNumber, seqNumber)
	if err != nil {
		return "", err
	}

	cli := c.Client.Auth().(types.BaseClient)
	txBytes, err := cli.GetCodec().MarshalBinaryLengthPrefixed(signedTx)
	if err != nil {
		return "", errors.Wrap(err, "MarshalJSON fail")
	}

	wrapTx := cmwraptx.WrapCMTx{
		txBytes,
		seqNumber,
	}

	wtxBytes, err := cli.GetCodec().MarshalJSON(wrapTx)
	if err != nil {
		return "", errors.Wrap(err, "MarshalJSON fail")
	}

	txRes, err := cli.Broadcast(wtxBytes, "block")
	if err != nil {
		return "", err
	}

	return parseContractAddress(txRes.RawLog), nil
}

func (c *CosmosClient) CreatePair(privateKey *ecdsa.PrivateKey, accNumber, seqNumber uint64, chainId string, memo string, contractAddr string, execMsg string, sender sdk.AccAddress, amountStr string) (string, error) {
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
	txRes, err := cli.Broadcast(bytes, "block")
	return parsePairAddress(txRes.RawLog), err
}

func parseCodeID(str string) uint64 {
	index := strings.LastIndex(str, ":")
	codeIDStr := str[index:]
	codeIDStr = codeIDStr[2 : strings.Index(codeIDStr, "}")-1]
	codeID, _ := strconv.Atoi(codeIDStr)

	return uint64(codeID)
}

func parseContractAddress(str string) string {
	index := strings.Index(str, "address")
	contractAddr := str[index+18 : index+18+42]
	return contractAddr
}

func parsePairAddress(str string) string {
	index := strings.Index(str, "pair_contract_addr")
	pairAddr := str[index+29 : index+29+42]
	return pairAddr
}

type rpcResult struct {
	Result MempoolResult `json:"result"`
}
type MempoolResult struct {
	Txs        string `json:"n_txs"`
	Total      string `json:"total"`
	TotalBytes string `json:"total_bytes"`
}
