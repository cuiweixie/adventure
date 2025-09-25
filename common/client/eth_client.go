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

// Create optimized HTTP client for connection pooling
func createOptimizedHTTPClient() *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        300,              // Increase maximum idle connections
		MaxIdleConnsPerHost: 300,              // Increase maximum idle connections per host
		IdleConnTimeout:     30 * time.Second, // Extend idle connection timeout
		DisableKeepAlives:   false,            // ✅ Enable keep-alive (key optimization)
		MaxConnsPerHost:     300,              // Limit maximum connections per host
	}

	return &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second, // Request timeout
	}
}

func NewEthClient(ip string) (*EthClient, error) {
	// Create RPC client using optimized HTTP client
	httpClient := createOptimizedHTTPClient()
	rpcClient, err := rpc.DialHTTPWithClient(ip, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize rpc client: %+v", err)
	}

	// Create Ethereum client based on RPC client
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

// Batch send signed transactions
func (e EthClient) SendMultipleEthereumTx(signedTxs []*types.Transaction) ([]ethcmn.Hash, error) {
	if len(signedTxs) == 0 {
		return nil, fmt.Errorf("empty transaction list")
	}

	// Prepare batch RPC requests
	batch := make([]rpc.BatchElem, len(signedTxs))
	txHashes := make([]string, len(signedTxs))

	for i, signedTx := range signedTxs {
		// Encode transaction as hexadecimal string
		txData, err := signedTx.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tx %d: %v", i, err)
		}
		txHex := "0x" + fmt.Sprintf("%x", txData)

		// Prepare batch RPC call elements
		batch[i] = rpc.BatchElem{
			Method: "eth_sendRawTransaction",
			Args:   []interface{}{txHex},
			Result: &txHashes[i],
		}
	}

	// Execute batch RPC call - only one HTTP request here!
	err := e.rpcClient.BatchCall(batch)
	if err != nil {
		return nil, fmt.Errorf("batch call failed: %v", err)
	}

	// Process results
	var resultHashes []ethcmn.Hash
	var errors []string

	for i, elem := range batch {
		if elem.Error != nil {
			errors = append(errors, fmt.Sprintf("tx %d: %v", i, elem.Error))
			resultHashes = append(resultHashes, ethcmn.Hash{})
		} else {
			// Convert string to Hash
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
