package celt

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/okex/adventure/evm/bench/celt/abi_bin"
	"github.com/petermattis/goid"
	"io"
	"log"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"
)

// CeltConfig define CeltConfig
type CeltConfig struct {
	RPC             []string
	ContractPath    string
	SuperAcc        string
	WorkerPath      string
	ParaNum         int
	SendOKTToWorker bool
}

// CeltContract define a Celt Project that contains all contract address
type CeltContract struct {
	Supreme          common.Address
	Common           common.Address
	InvitationCenter common.Address
	NftPool          common.Address
	LinearUnlock     common.Address
	CeltManager      common.Address
}

//
type acc struct {
	privateKey string
	ecdsaPriv  *ecdsa.PrivateKey
	ethAddress common.Address
}

type CeltManager struct {
	clientList  []*ethclient.Client
	contracList []CeltContract
	superAcc    *acc

	worker          []*acc
	paraNum         int
	sendOKTToWorker bool
}

func newManager(cList []CeltContract, superAcc *acc, workPath string, paraNum int, clients []*ethclient.Client, sendOKTToWorker bool) *CeltManager {
	m := &CeltManager{
		clientList:      clients,
		contracList:     cList,
		superAcc:        superAcc,
		paraNum:         paraNum,
		sendOKTToWorker: sendOKTToWorker,
	}
	m.prePareWorker(workPath)
	m.displayDetail()
	return m
}

func (m *CeltManager) displayDetail() {
	fmt.Println("contract size", len(m.contracList), "worker size:", len(m.worker), "paraNum", m.paraNum, "clientNum", len(m.clientList))
}

func (m *CeltManager) prePareWorker(path string) {
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
			time.Sleep(2 * time.Second)
		} else {
			return nonce
		}
		cnt++
	}
	panic("GetNonce Failed")

}

func (m *CeltManager) Loop() {
	fmt.Printf("begin send celt")

	var wg sync.WaitGroup
	workerIndex := 0

	for index := 0; index < m.paraNum; index++ {
		groupSize := len(m.worker) / m.paraNum
		workIndexList := make([]int, 0, groupSize)
		for i := 0; i < groupSize; i++ {
			workIndexList = append(workIndexList, workerIndex)
			workerIndex++
		}

		contractIndex := index % len(m.contracList)
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.run(workIndexList, contractIndex)
		}()
	}
	wg.Wait()
}

var (
	ether = new(big.Int).Mul(new(big.Int).SetInt64(1000000000), new(big.Int).SetInt64(1000000000))
)

func (m *CeltManager) Init() {
	m.TransferOKTToAccount()
	fmt.Println("init mint")
	if err := m.InitMint(); err != nil {
		panic(err)
	}

	fmt.Println("init register")
	if err := m.InitRegister(); err != nil {
		panic(err)
	}

	fmt.Println("init approval for all")
	if err := m.InitApprovalForAll(); err != nil {
		panic(err)
	}

	fmt.Println("init stake")
	if err := m.InitStake(); err != nil {
		panic(err)
	}
}

func (m *CeltManager) TransferOKTToAccount() {
	nonce := GetNonce(m.clientList[0], m.superAcc.ecdsaPriv)
	txs := make([]*types.Transaction, 0)
	for _, acc := range m.worker {
		if m.sendOKTToWorker {
			tx := transferOkt(m.superAcc.privateKey, acc.ethAddress, nonce, ether)
			nonce++
			txs = append(txs, tx)
		}
	}

	if err := SendTxs(m.clientList[0], txs); err != nil {
		panic(err)
	}
}

func (m *CeltManager) InitMint() error {
	for _, contract := range m.contracList {
		txList := make([]*types.Transaction, 0, len(m.worker)*2)
		nonce := GetNonce(m.clientList[0], m.superAcc.ecdsaPriv)
		for _, account := range m.worker {
			// supreme mintSudo
			payload, err := abi_bin.SurpemeBuilder.Build("mintSudo", account.ethAddress, big.NewInt(1))
			if err != nil {
				return err
			}
			txList = append(txList, SignTxWithNonce(m.superAcc.ecdsaPriv, contract.Supreme, payload, nonce))

			// common mintSudo
			nonce++
			payload, err = abi_bin.CommonBuilder.Build("mintSudo", account.ethAddress, big.NewInt(1), big.NewInt(100))
			if err != nil {
				return err
			}
			txList = append(txList, SignTxWithNonce(m.superAcc.ecdsaPriv, contract.Supreme, payload, nonce))

			nonce++
		}

		// send txs
		if err := SendTxs(m.clientList[0], txList); err != nil {
			return err
		}
	}

	return nil
}

func (m *CeltManager) InitRegister() error {
	for _, contract := range m.contracList {
		txList := make([]*types.Transaction, 0, len(m.worker))
		for _, account := range m.worker {
			nonce := GetNonce(m.clientList[0], account.ecdsaPriv)

			// registerInviter
			payload, err := abi_bin.InvitationCenterBuilder.Build("registerInviter", StringToBytes32("deadbeef"), big.NewInt(1), "bzh")
			if err != nil {
				return err
			}

			txList = append(txList, SignTxWithNonce(account.ecdsaPriv, contract.InvitationCenter, payload, nonce))
		}

		if err := SendTxs(m.clientList[0], txList); err != nil {
			return err
		}
	}

	return nil
}

func (m *CeltManager) InitApprovalForAll() error {
	for _, contract := range m.contracList {
		txList := make([]*types.Transaction, 0, len(m.worker)*2)
		for _, account := range m.worker {
			nonce := GetNonce(m.clientList[0], account.ecdsaPriv)

			// supreme setApprovalForAll
			payload, err := abi_bin.SurpemeBuilder.Build("setApprovalForAll", contract.CeltManager, true)
			if err != nil {
				return err
			}

			txList = append(txList, SignTxWithNonce(account.ecdsaPriv, contract.Supreme, payload, nonce))

			nonce++
			// common setApprovalForAll
			payload, err = abi_bin.CommonBuilder.Build("setApprovalForAll", contract.CeltManager, true)
			if err != nil {
				return err
			}

			txList = append(txList, SignTxWithNonce(account.ecdsaPriv, contract.Common, payload, nonce))
		}

		if err := SendTxs(m.clientList[0], txList); err != nil {
			return err
		}
	}

	return nil
}

func (m *CeltManager) InitStake() error {
	for _, contract := range m.contracList {
		txList := make([]*types.Transaction, 0, len(m.worker)*2)
		for _, account := range m.worker {
			nonce := GetNonce(m.clientList[0], account.ecdsaPriv)

			// nftpool setApprovalForAll
			payload, err := abi_bin.NftPoolBuilder.Build("stake", []*big.Int{big.NewInt(1)}, []*big.Int{big.NewInt(100)}, []*big.Int{big.NewInt(1)})
			if err != nil {
				return err
			}

			txList = append(txList, SignTxWithNonce(account.ecdsaPriv, contract.NftPool, payload, nonce))
		}

		if err := SendTxs(m.clientList[0], txList); err != nil {
			return err
		}
	}

	return nil
}

func (m *CeltManager) runPool(workIndex int, contractIndex int) error {
	account := m.worker[workIndex]
	contract := m.contracList[contractIndex]

	nonce := GetNonce(m.clientList[workIndex%len(m.clientList)], account.ecdsaPriv)
	txList := make([]*types.Transaction, 0)

	// nftpool getRewardAndBonus
	payload, err := abi_bin.NftPoolBuilder.Build("getRewardAndBonus")
	if err != nil {
		return err
	}

	txList = append(txList, SignTxWithNonce(account.ecdsaPriv, contract.NftPool, payload, nonce))

	if err := SendTxs(m.clientList[workIndex%len(m.clientList)], txList); err != nil {
		return err
	}

	return nil
}

func (m *CeltManager) run(tasks []int, contractIndex int) {

	//rand.Seed(time.Now().UnixNano())
	//sleepTime := rand.Intn(10)
	//time.Sleep(time.Duration(sleepTime) * time.Second)
	goId := goid.Get()
	for true {
		for _, workIndex := range tasks {
			fmt.Printf("goroutine %d run  workerIndex %d contractIndex %d\n", goId, workIndex, contractIndex)
			if err := m.runPool(workIndex, contractIndex); err != nil {
				log.Println(err)
			}
		}
	}
}
