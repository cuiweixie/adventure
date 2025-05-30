package batch_transfer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/okex/adventure/common"
	"github.com/okex/adventure/common/client"
	"github.com/okex/adventure/evm/constant"
)

var (
	privateKey  string
	addressFile string
	amount      int
)

func batchTransfer(cmd *cobra.Command, args []string) {

	// 0.1 load env parameters
	cli, privateKey, addrs := loadEnv()
	// query nonce
	nonce, err := cli.QueryNonce(common.GetEthAddressFromPK(privateKey).String())
	if err != nil {
		log.Println(fmt.Errorf("failed to query nonce, error", err))
		return
	}

	// 1. deploy
	contractAddr, err := deploy(cli, privateKey, nonce)
	if err != nil {
		log.Println(fmt.Errorf("failed to deploy, error: %s", err))
		return
	}
	time.Sleep(time.Second * 5)
	// 2. transfers
	amount, ok := new(big.Int).SetString(args[0], 10)
	if !ok {
		panic("failed to parse amount")
	}
	if err := transfers(cli, privateKey, nonce+1, contractAddr, amount, addrs); err != nil {
		log.Println(fmt.Errorf("failed to transfer, error: %s", err))
		return
	}

	//file, _ := os.Open("/Users/finefine/workspace/ok/adventure/config/devnet/addr_50000_transfer")
	//defer file.Close()
	//scanner := bufio.NewScanner(file)
	//targetFile, _ := os.Create("/Users/finefine/workspace/ok/adventure/config/devnet/address_50000_transfer")
	//writer := bufio.NewWriter(targetFile)
	//defer targetFile.Close()
	//
	//for scanner.Scan() {
	//	key := scanner.Text()
	//	privateKey, err := crypto.HexToECDSA(key)
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	publicKey := privateKey.Public()
	//	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	//	if !ok {
	//		panic("keyToAcc")
	//	}
	//	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	//	fmt.Println(key, address.String())
	//
	//	writer.WriteString(address.String())
	//	writer.WriteString("\n")
	//}

}

func loadEnv() (client.Client, *ecdsa.PrivateKey, []ethcmn.Address) {
	ips := viper.GetStringSlice(constant.FlagIPs)
	if len(ips) == 0 {
		panic(fmt.Errorf("ip list is nil, please set them in flag %s", constant.FlagIPs))
	}
	cli := client.NewClient(ips[0])

	privateKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		panic(fmt.Errorf("failed to unencrypted private key [%s]: %s", privateKey, err))
	}

	addresses := constant.HexAddresses
	if addressFile != "" {
		addresses = common.ReadDataFromFile(addressFile)
		// make sure addressFile has at least one address or one private key
		if len(addresses) == 0 {
			addresses = constant.HexAddresses
		}
	}

	hexAddrs := make([]ethcmn.Address, len(addresses), len(addresses))
	if !strings.HasPrefix(addresses[0], "0x") {
		// support private key file input
		for i, addr := range addresses {
			privKey, err := crypto.HexToECDSA(addr)
			if err != nil {
				log.Println(fmt.Errorf("failed to convert private key string %s, error: %s", addr, err))
				break
			}
			hexAddrs[i] = common.GetEthAddressFromPK(privKey)
		}
	} else {
		for i, addr := range addresses {
			hexAddrs[i] = ethcmn.HexToAddress(addr)
		}
	}

	return cli, privateKey, hexAddrs
}

func deploy(cli client.Client, privateKey *ecdsa.PrivateKey, nonce uint64) (ethcmn.Address, error) {
	// Query GasPrice
	var gasPrice *big.Int
	ethClient, ok := cli.(*client.EthClient)
	if ok {
		gasPrice = getGasPrice(ethClient.Client)
	}

	// deploy contract BatchTransfer contract
	txhash, err := cli.CreateContract(privateKey, nonce, nil, 300000, gasPrice, ethcmn.Hex2Bytes(constant.BatchTransferHex))
	if err != nil {
		return ethcmn.Address{}, err
	}
	contractAddr := crypto.CreateAddress(common.GetEthAddressFromPK(privateKey), nonce)
	log.Printf("caller: %s, nonce: %d, contract: %s, txhash: %s\n", common.GetEthAddressFromPK(privateKey), nonce, contractAddr, txhash)

	return contractAddr, nil
}

func transfers(cli client.Client, privateKey *ecdsa.PrivateKey, nonce uint64, to ethcmn.Address, amount *big.Int, addrs []ethcmn.Address) error {
	// load abi
	tABI, err := abi.JSON(strings.NewReader(constant.BatchTransferABI))
	if err != nil {
		return fmt.Errorf("failed to initialize BatchTransfer abi, error: %s", err)
	}

	batchNum := 200 // 40,000 gas per address
	totalAmount := big.NewInt(1).Mul(amount, big.NewInt(int64(batchNum)))

	// Query GasPrice
	var gasPrice *big.Int
	ethClient, ok := cli.(*client.EthClient)
	if ok {
		gasPrice = getGasPrice(ethClient.Client)
	}

	for i := 0; i <= len(addrs)/batchNum && i*batchNum < len(addrs); i++ {
		start, end := i*batchNum, (i+1)*batchNum
		if end > len(addrs) {
			end = len(addrs)
		}
		txdata, err := tABI.Pack("transfers", addrs[start:end], amount)
		if err != nil {
			return fmt.Errorf("failed to pack BatchTransfer parameters, error: %s", err)
		}
		if end-start < batchNum {
			totalAmount = big.NewInt(1).Mul(amount, big.NewInt(int64(end-start)))
		}
		txhash, err := cli.SendEthereumTx(privateKey, nonce, to, totalAmount, uint64(41000*batchNum), gasPrice, txdata)
		if err != nil {
			return err
		}
		log.Printf("caller: %s, nonce: %d, to[%d:%d] txhash: %s\n", common.GetEthAddressFromPK(privateKey), nonce, start, end-1, txhash)

		nonce++
		//time.Sleep(time.Second)
	}

	return nil
}

var defaultGasPrice = big.NewInt(10000000000)

func getGasPrice(client *ethclient.Client) *big.Int {
	return defaultGasPrice
	var gp *big.Int
	var err error
	var incAmount = new(big.Int).SetUint64(10000000000)

	for {
		gp, err = client.SuggestGasPrice(context.Background())
		if err != nil {
			time.Sleep(10 * time.Microsecond)
		} else {
			break
		}
	}
	gp.Add(gp, incAmount)
	return gp
}
