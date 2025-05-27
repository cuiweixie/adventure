package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net/http"
	"time"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type EthClient struct {
	*ethclient.Client
	rpcClient *rpc.Client
	signer    types.Signer
}

// 创建优化的HTTP客户端，用于连接池
func createOptimizedHTTPClient() *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        300,              // 增加最大空闲连接数
		MaxIdleConnsPerHost: 300,              // 增加每个主机的最大空闲连接数
		IdleConnTimeout:     30 * time.Second, // 延长空闲连接超时
		DisableKeepAlives:   false,            // ✅ 启用keep-alive（关键优化）
		MaxConnsPerHost:     300,              // 限制每个主机的最大连接数
	}

	return &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second, // 请求超时
	}
}

func NewEthClient(ip string) (*EthClient, error) {
	// 使用优化的HTTP客户端创建RPC客户端
	httpClient := createOptimizedHTTPClient()
	rpcClient, err := rpc.DialHTTPWithClient(ip, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize rpc client: %+v", err)
	}

	// 基于RPC客户端创建以太坊客户端
	cli := ethclient.NewClient(rpcClient)

	chainId, err := cli.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	return &EthClient{
		cli,
		rpcClient,
		types.NewLondonSigner(chainId),
	}, nil
}

func (e EthClient) QueryNonce(hexAddr string) (uint64, error) {
	nonce, err := e.PendingNonceAt(context.Background(), ethcmn.HexToAddress(hexAddr))
	if err != nil {
		return 0, err
	}
	return nonce, nil
}

func (e EthClient) SendEthereumTx(privatekey *ecdsa.PrivateKey, nonce uint64, to ethcmn.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) (ethcmn.Hash, error) {
	// 1. make tx
	unsignedTx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)

	// 2. sign unsignedTx -> rawTx
	signedTx, err := types.SignTx(unsignedTx, e.signer, privatekey)
	if err != nil {
		return ethcmn.Hash{}, err
	}

	// 3. send rawTx
	err = e.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return ethcmn.Hash{}, err
	}

	return signedTx.Hash(), err
}

func (e EthClient) CreateContract(privatekey *ecdsa.PrivateKey, nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) (ethcmn.Hash, error) {
	// 1. make tx
	unsignedTx := types.NewContractCreation(nonce, amount, gasLimit, gasPrice, data)

	// 2. sign unsignedTx -> rawTx
	signedTx, err := types.SignTx(unsignedTx, e.signer, privatekey)
	if err != nil {
		return ethcmn.Hash{}, err
	}

	// 3. send rawTx
	err = e.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return ethcmn.Hash{}, err
	}

	return signedTx.Hash(), err
}

// 批量发送已签名的交易
func (e EthClient) SendMultipleEthereumTx(signedTxs []*types.Transaction) ([]ethcmn.Hash, error) {
	if len(signedTxs) == 0 {
		return nil, fmt.Errorf("empty transaction list")
	}

	// 准备批量RPC请求
	batch := make([]rpc.BatchElem, len(signedTxs))
	txHashes := make([]string, len(signedTxs))

	for i, signedTx := range signedTxs {
		// 将交易编码为十六进制字符串
		txData, err := signedTx.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tx %d: %v", i, err)
		}
		txHex := "0x" + fmt.Sprintf("%x", txData)

		// 准备批量RPC调用元素
		batch[i] = rpc.BatchElem{
			Method: "eth_sendRawTransaction",
			Args:   []interface{}{txHex},
			Result: &txHashes[i],
		}
	}

	// 执行批量RPC调用 - 这里只有一次HTTP请求！
	err := e.rpcClient.BatchCall(batch)
	if err != nil {
		return nil, fmt.Errorf("batch call failed: %v", err)
	}

	// 处理结果
	var resultHashes []ethcmn.Hash
	var errors []string

	for i, elem := range batch {
		if elem.Error != nil {
			errors = append(errors, fmt.Sprintf("tx %d: %v", i, elem.Error))
			resultHashes = append(resultHashes, ethcmn.Hash{})
		} else {
			// 将字符串转换为Hash
			if txHashes[i] != "" {
				resultHashes = append(resultHashes, ethcmn.HexToHash(txHashes[i]))
			} else {
				resultHashes = append(resultHashes, ethcmn.Hash{})
			}
		}
	}

	if len(errors) > 0 {
		return resultHashes, fmt.Errorf("batch errors: %v", errors)
	}

	return resultHashes, nil
}
