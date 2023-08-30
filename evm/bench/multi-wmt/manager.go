package multiwmt

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type acc struct {
	privateKey string
	ecdsaPriv  *ecdsa.PrivateKey
	ethAddress common.Address
}

type nonceManager struct {
	mu     sync.Mutex
	mp     map[common.Address]uint64
	nonces []uint64
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

type okcClient struct {
	*ethclient.Client
}

type rpcResult struct {
	Result MempoolResult `json:"result"`
}
type MempoolResult struct {
	Txs        string `json:"n_txs"`
	Total      string `json:"total"`
	TotalBytes string `json:"total_bytes"`
}

func (okc *okcClient) GetMempoolSize() int {

	//if okc.rpc == "" {
	//	return math.MaxInt
	//}
	//
	//var result rpcResult
	//response, err := http.Get(fmt.Sprintf("%s/num_unconfirmed_txs", okc.rpc))
	//if err != nil {
	//	fmt.Println(err)
	//	return 0
	//}
	//
	//bts, err := ioutil.ReadAll(response.Body)
	//if err != nil {
	//	fmt.Println(err)
	//	return 0
	//}
	//
	//err = json.Unmarshal(bts, &result)
	//if err != nil {
	//	fmt.Println(err)
	//	return 0
	//}
	//
	//fmt.Println("mempool size :", result.Result.Total)
	//total, _ := strconv.Atoi(result.Result.Total)
	//return total

	var txcount uint
	var err error

	for {
		txcount, err = okc.PendingTransactionCount(context.Background())
		if err != nil {
			time.Sleep(1000 * time.Microsecond)
		} else {
			break
		}
	}

	var curblocknum uint64
	for {
		curblocknum, err = okc.BlockNumber(context.Background())
		if err != nil {
			time.Sleep(1000 * time.Microsecond)
		} else {
			break
		}
	}

	fmt.Printf("Get PendingTransactionCount = %d, Current Block Number is = %d\n", txcount, curblocknum)
	return int(txcount)
}

type wmtManager struct {
	clientList  []*okcClient
	contracList []SwapContract
	superAcc    *acc
	nonceM      *nonceManager

	worker          []*acc
	paraNum         int
	sendOKTToWorker bool
	threshold       int
}

func newManager(cList []SwapContract, superAcc *acc, workPath string, paraNum int, clients []*okcClient, sendOKTToWorker bool, threshold int) *wmtManager {
	m := &wmtManager{
		clientList:      clients,
		contracList:     cList,
		superAcc:        superAcc,
		paraNum:         paraNum,
		sendOKTToWorker: sendOKTToWorker,
		threshold:       threshold,
		nonceM: &nonceManager{
			mu: sync.Mutex{},
			mp: make(map[common.Address]uint64),
		},
	}
	m.prePareWorker(workPath)
	m.displayDetail()
	m.initNonce()
	return m
}

func (m *wmtManager) initNonce() {
	fmt.Println("init nonce start , please wait")
	var wg sync.WaitGroup
	for _, v := range m.worker {
		v := v
		wg.Add(1)
		go func() {
			nonce := GetNonce(m.clientList[0], v.ecdsaPriv)
			m.nonceM.setNonce(v.ethAddress, nonce)
			wg.Done()
			if m.nonceM.addrSize()%200 == 0 {
				fmt.Println("workerSize", len(m.worker), "have init", m.nonceM.addrSize())
			}
		}()
	}
	wg.Wait()
	fmt.Println("init nonce end ", "workerSize", len(m.worker), "nonceManagerAddr", m.nonceM.addrSize())
}

func (m *wmtManager) displayDetail() {
	fmt.Println("contract size", len(m.contracList), "worker size:", len(m.worker), "paraNum", m.paraNum, "clientNum", len(m.clientList))
}

func (m *wmtManager) prePareWorker(path string) {
	f, err := os.Open(path)
	panicerr(err)
	defer f.Close()

	accList := make([]*acc, 0)
	rd := bufio.NewReader(f)
	for true {
		privKey, err := rd.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}

		acc := keyToAcc(strings.TrimSpace(privKey))
		accList = append(accList, acc)
	}
	m.worker = accList
}

func GetNonce(client *okcClient, privateKey *ecdsa.PrivateKey) uint64 {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("GetNonce Failed")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	for {
		nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
		if err != nil {
			time.Sleep(1000 * time.Microsecond)
		} else {
			return nonce
		}
	}
}

func (m *wmtManager) Loop() {
	tasks := make([][]int, m.paraNum)

	for index, _ := range m.worker {
		paraIndex := index % m.paraNum
		tasks[paraIndex] = append(tasks[paraIndex], index)
	}

	for index := 0; index < m.paraNum; index++ {
		fmt.Println("goRoutine", index, "worker", tasks[index])
	}

	fmt.Println("====== begin send wmt =======")

	var wg sync.WaitGroup
	for index := 0; index < m.paraNum; index++ {
		index := index
		wg.Add(1)
		go func() {
			m.run(tasks[index])
			wg.Done()
		}()
	}
	wg.Wait()
}

func (m *wmtManager) needTransferToWorker() bool {
	return false
	// TODO : later to fix
	//for _, acc := range m.worker {
	//	bal, err := client.BalanceAt(context.Background(), acc.ethAddress, nil)
	//	if err == nil && bal.Int64() == 0 {
	//		return true
	//	}
	//}
	//return false
}

var (
	ether = new(big.Int).Mul(new(big.Int).SetInt64(1000000000), new(big.Int).SetInt64(1000000000))
)

func display(client *okcClient, acc *acc, to common.Address, payload []byte) {
	data, err := client.CallContract(context.Background(), ethereum.CallMsg{
		From:     acc.ethAddress,
		To:       &to,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     payload,
	}, nil)
	if err == nil {
		fmt.Println("addr", acc.ethAddress.String(), "token balance", new(big.Int).SetBytes(data).String())
	} else {
		fmt.Println("err", err)
	}
}

func (m *wmtManager) DisPlayToken() {
	for _, acc := range m.worker {
		for _, c := range m.contracList {
			payload, err := erc20Builder.Build("balanceOf", acc.ethAddress)
			panicerr(err)
			display(m.clientList[0], acc, c.Token0, payload)
			display(m.clientList[0], acc, c.Token2, payload)
		}
	}
}

func (m *wmtManager) TransferGas(amount int64) {
	gasAmount := new(big.Int).Mul(new(big.Int).SetInt64(1000000000), new(big.Int).SetInt64(1000000000))
	gasAmount = new(big.Int).Mul(gasAmount, new(big.Int).SetInt64(amount))
	nonce := GetNonce(m.clientList[0], m.superAcc.ecdsaPriv)
	txs := make([]*types.Transaction, 0)
	for _, acc := range m.worker {
		if m.sendOKTToWorker {
			tx := transferOkt(m.superAcc.privateKey, acc.ethAddress, nonce, gasAmount)
			nonce++
			txs = append(txs, tx)
		}
	}
	fmt.Println("sendTx", len(txs), "use one node,may slow")
	if err := SendTxs(m.clientList[0], txs); err != nil {
		panic(err)
	}
	fmt.Println("end gas transfer")
}

func (m *wmtManager) TransferToken0ToAccount() {

	nonce := GetNonce(m.clientList[0], m.superAcc.ecdsaPriv)
	fmt.Println("Begin TransferToken0ToAccount", "transferOkT:", m.sendOKTToWorker)
	txs := make([]*types.Transaction, 0)
	for _, acc := range m.worker {
		if m.sendOKTToWorker {
			tx := transferOkt(m.superAcc.privateKey, acc.ethAddress, nonce, ether)
			nonce++
			txs = append(txs, tx)
		}
		for _, c := range m.contracList {

			payload, err := erc20Builder.Build("transfer", acc.ethAddress, new(big.Int).SetInt64(10000000000))
			panicerr(err)
			tx := SignTxWithNonce(m.superAcc.ecdsaPriv, c.Token0, payload, nonce)
			nonce++
			txs = append(txs, tx)

			tx = SignTxWithNonce(m.superAcc.ecdsaPriv, c.Token2, payload, nonce)
			nonce++
			txs = append(txs, tx)
		}
	}
	fmt.Println("sendTx", len(txs), "use one node,may slow")
	if err := SendTxs(m.clientList[0], txs); err != nil {
		panic(err)
	}
	fmt.Println("end transferToken0ToAccount")
}

func (m *wmtManager) randomContract() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(5)
}

func (m *wmtManager) runPool(poolIndex int, workIndex int, getReward bool) error {
	contractIndex := workIndex % len(m.contracList)
	a := m.worker[workIndex]
	c := m.contracList[contractIndex]

	if m.clientList[workIndex%len(m.clientList)].GetMempoolSize() > m.threshold {
		fmt.Println("达到阈值")
		return nil
	}

	fmt.Println("run---", "workerIndex", workIndex, "contractIndex", contractIndex)

	token0 := c.Token0
	token1 := c.Token1
	lp := c.Lp1
	stakeRewards := c.StakingRewards1

	if poolIndex == 1 {
		token0 = c.Token2
		token1 = c.Token3
		lp = c.Lp2
		stakeRewards = c.StakingRewards2
	}

	txList := make([]*types.Transaction, 0)

	nonce := m.nonceM.getNonce(a.ethAddress)
	// approve token0
	payload, err := erc20Builder.Build("approve", c.Router, new(big.Int).SetInt64(1000))
	panicerr(err)
	txList = append(txList, SignTxWithNonce(a.ecdsaPriv, token0, payload, nonce))
	nonce++

	// swap token0->token1
	payload, err = routerBuilder.Build("swapExactTokensForTokens",
		new(big.Int).SetInt64(500), new(big.Int),
		[]common.Address{token0, token1},
		a.ethAddress, big.NewInt(1956981781),
	)
	panicerr(err)
	txList = append(txList, SignTxWithNonce(a.ecdsaPriv, c.Router, payload, nonce))
	nonce++

	// approve token1 (for addLiquidity)
	payload, err = erc20Builder.Build("approve", c.Router, new(big.Int).SetInt64(30))
	panicerr(err)
	txList = append(txList, SignTxWithNonce(a.ecdsaPriv, token1, payload, nonce))
	nonce++

	// addLiquidity
	payload, err = routerBuilder.Build("addLiquidity", token0, token1, new(big.Int).SetInt64(30), new(big.Int).SetInt64(30), new(big.Int), new(big.Int), a.ethAddress, big.NewInt(1956981781))
	panicerr(err)
	txList = append(txList, SignTxWithNonce(a.ecdsaPriv, c.Router, payload, nonce))
	nonce++

	// approve lp for stakingRewards
	payload, err = erc20Builder.Build("approve", stakeRewards, new(big.Int).SetInt64(10))
	panicerr(err)
	txList = append(txList, SignTxWithNonce(a.ecdsaPriv, lp, payload, nonce))
	nonce++

	// stake lp for stakingRewards
	payload, err = StakingRewardsBuilder.Build("stake", big.NewInt(10))
	panicerr(err)
	txList = append(txList, SignTxWithNonce(a.ecdsaPriv, stakeRewards, payload, nonce))
	nonce++

	if getReward {
		// getReward for stakingRewards
		payload, err = StakingRewardsBuilder.Build("getReward")
		panicerr(err)
		txList = append(txList, SignTxWithNonce(a.ecdsaPriv, stakeRewards, payload, nonce))
		nonce++
	}

	payload, err = StakingRewardsBuilder.Build("withdraw", big.NewInt(3))
	panicerr(err)
	txList = append(txList, SignTxWithNonce(a.ecdsaPriv, stakeRewards, payload, nonce))
	nonce++

	if getReward {
		payload, err = StakingRewardsBuilder.Build("exit")
		panicerr(err)
		txList = append(txList, SignTxWithNonce(a.ecdsaPriv, stakeRewards, payload, nonce))
		nonce++
	}

	if err := SendTxs(m.clientList[workIndex%len(m.clientList)], txList); err != nil {
		fmt.Println("SendTxs failed", err)
		time.Sleep(60 * time.Second)
		m.nonceM.setNonce(a.ethAddress, GetNonce(m.clientList[0], a.ecdsaPriv))
		return err
	}

	time.Sleep(2 * time.Second)
	m.nonceM.setNonce(a.ethAddress, nonce)
	return nil
}

func (m *wmtManager) run(tasks []int) {

	rand.Seed(time.Now().UnixNano())
	sleepTime := rand.Intn(10)
	time.Sleep(time.Duration(sleepTime) * time.Second)
	turns := 0
	for true {
		for _, workIndex := range tasks {
			getReward := turns%2 == 1
			if err := m.runPool(0, workIndex, getReward); err != nil {
				fmt.Println("runErr-0", workIndex, err)
				continue
			}

			if err := m.runPool(1, workIndex, getReward); err != nil {
				fmt.Println("runErr-1", workIndex, err)
				continue
			}

		}
		turns++
	}
}
