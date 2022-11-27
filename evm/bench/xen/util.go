package xen

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"sync"
	"time"
)

var (
	xenConfigFile = "./config/wmt.json"
)

func panicerr(err error) {
	if err != nil {
		panic(err)
	}
}

type nonceManager struct {
	mu sync.Mutex
	mp map[common.Address]uint64
}

func (n *nonceManager) addrSize() int {
	n.mu.Lock()

	defer n.mu.Unlock()
	return len(n.mp)
}

func (n *nonceManager) setNonce(addr common.Address, nonce uint64) {
	n.mu.Lock()
	n.mp[addr] = nonce
	n.mu.Unlock()
}
func (n *nonceManager) getNonce(addr common.Address) uint64 {
	n.mu.Lock()
	defer n.mu.Unlock()
	nonce := n.mp[addr]
	return nonce
}

type acc struct {
	index      int
	privateKey string
	ecdsaPriv  *ecdsa.PrivateKey
	ethAddress common.Address
}

type xenInfo struct {
	mintTime  int64
	claimTime int64
}

func (x *xenInfo) string() string {
	return fmt.Sprintf("mintIndex:%d clainIndex:%d", x.mintTime, x.claimTime)
}

type xenManager struct {
	clientList      []*ethclient.Client
	coinToolAddress common.Address
	xenAddress      common.Address
	nonceM          *nonceManager

	worker  []*acc
	paraNum int

	xenInfos []*xenInfo
}

func getPrivateKey(key string) *ecdsa.PrivateKey {
	privateKey, err := crypto.HexToECDSA(key)
	panicerr(err)
	return privateKey
}

func keyToAcc(key string) *acc {
	privateKey := getPrivateKey(key)
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("keyToAcc")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	return &acc{
		privateKey: key,
		ecdsaPriv:  privateKey,
		ethAddress: fromAddress,
	}
}

func GetNonce(client *ethclient.Client, privateKey *ecdsa.PrivateKey) uint64 {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("GetNonce Failed")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	cnt := 0
	for cnt < 50 {
		nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
		if err != nil {
			time.Sleep(1000 * time.Microsecond)
		} else {
			return nonce
		}
		cnt++
	}
	panic("GetNonce Failed")

}

type xenConfig struct {
	RPC             []string
	XenAddress      string
	CoinToolAddress string
	WorkerPath      string
	ParaNum         int
}

func loadXenConfig(file string) *xenConfig {
	data, err := ioutil.ReadFile(file)
	panicerr(err)
	c := new(xenConfig)
	err = json.Unmarshal(data, c)
	panicerr(err)
	return c
}

func initClient(c *xenConfig) {
	client, err := ethclient.Dial(c.RPC[0])
	panicerr(err)
	chainID, err := client.ChainID(context.Background())
	panicerr(err)
	signer = types.NewEIP155Signer(chainID)
}

func getM() *xenManager {
	c := loadXenConfig(xenConfigFile)

	initBuilder()
	initClient(c)

	clients := make([]*ethclient.Client, 0)
	for _, v := range c.RPC {
		client, err := ethclient.Dial(v)
		panicerr(err)
		clients = append(clients, client)
	}

	m := newXenManager(clients, common.Address{}, c.ParaNum, c.WorkerPath)
	m.xenAddress = common.HexToAddress(c.XenAddress)
	m.coinToolAddress = common.HexToAddress(c.CoinToolAddress)
	return m
}
