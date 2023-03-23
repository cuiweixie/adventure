package client

import (
	"crypto/ecdsa"
	"fmt"
	gosdk "github.com/okex/exchain-go-sdk"
	"github.com/okex/exchain-go-sdk/types"
	"github.com/okex/exchain/x/common"
	"math/big"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

var (
	_ Client = (*CosmosClient)(nil)
	_ Client = (*EthClient)(nil)
)

type Client interface {
	QueryNonce(hexAddr string) (uint64, error)
	SendEthereumTx(privatekey *ecdsa.PrivateKey, nonce uint64, to ethcmn.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) (ethcmn.Hash, error)
	CreateContract(privatekey *ecdsa.PrivateKey, nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) (ethcmn.Hash, error)
}

func NewClient(ip string) Client {
	cosmosClient, err1 := NewCosmosClient(ip)
	if err1 == nil {
		return cosmosClient
	}

	ethClient, err2 := NewEthClient(ip)
	if err2 == nil {
		return ethClient
	}

	panic(fmt.Errorf(`failed to initialize client in CosmosClient or EthClient. cosmos error: %s, eth error: %s`, err1, err2))
}

func GenerateClients(ips []string) (clients []Client) {
	for _, ip := range ips {
		clients = append(clients, NewClient(ip))
	}
	return
}

func GenerateGoSDKClients(ips []string) (clients []*gosdk.Client) {
	for _, ip := range ips {
		cfg, err := types.NewClientConfig(ip, "exchain-64", types.BroadcastSync, "", 2000000, 1.5, "0.0000000001"+common.NativeToken)
		if err != nil {
			panic(fmt.Errorf("initialize client failed: %s", err))
		}
		cli := gosdk.NewClient(cfg)
		clients = append(clients, &cli)
	}
	return
}
