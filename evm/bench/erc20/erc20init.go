package erc20

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

var TotalSupplyAmount = big.NewInt(100000000)

var (
	privateKey  string
	addressFile string
)

func erc20init(cmd *cobra.Command, args []string) {
	// 0. load env parameters
	cli, privateKey, addrs := loadEnv()
	// 1.1 query nonce
	nonce, err := cli.QueryNonce(common.GetEthAddressFromPK(privateKey).String())
	if err != nil {
		log.Println(fmt.Errorf("failed to query nonce, error", err))
		return
	}

	// 1.2 deploy BatchTransfer for NativeToken
	nativeAddr, err := deployBTNative(cli, privateKey, nonce)
	if err != nil {
		log.Println(fmt.Errorf("failed to deploy BatchTransfer for Native Token, error: %s", err))
		return
	}
	time.Sleep(time.Second * 5)

	amount, ok := new(big.Int).SetString(args[0], 10)
	if !ok {
		panic("failed to parse amount")
	}

	// 1.3 transfers Native Token
	if err := transfers(cli, privateKey, nonce+1, nativeAddr, amount, addrs); err != nil {
		log.Println(fmt.Errorf("failed to transfer Native Token, error: %s", err))
		return
	}

	// 2.1 query nonce again
	nonce, err = cli.QueryNonce(common.GetEthAddressFromPK(privateKey).String())
	if err != nil {
		log.Println(fmt.Errorf("failed to query nonce, error", err))
		return
	}

	// 2.1 deploy ERC20, BatchTransfer for ERC20

	bterc20Addr, err := deployBTERC20(cli, privateKey, nonce)
	if err != nil {
		log.Println(fmt.Errorf("failed to deploy BatchTransfer for ERC20, error: %s", err))
		return
	}
	time.Sleep(time.Second * 5)

	erc20Addr, err := deployERC20(cli, privateKey, nonce+1)
	if err != nil {
		log.Println(fmt.Errorf("failed to deploy ERC20, error: %s", err))
		return
	}
	time.Sleep(time.Second * 5)

	err = sendApprove(cli, privateKey, nonce+2, erc20Addr, TotalSupplyAmount, bterc20Addr)
	if err != nil {
		log.Println(fmt.Errorf("failed to approve ERC20 for BatchTransfer Contract, error: %s", err))
		return
	}
	time.Sleep(time.Second * 5)

	// 2.2 query nonce again
	nonce, err = cli.QueryNonce(common.GetEthAddressFromPK(privateKey).String())
	if err != nil {
		log.Println(fmt.Errorf("failed to query nonce, error", err))
		return
	}

	// 2.3 transfers ERC20
	accBalance := TotalSupplyAmount.Int64() / int64(len(addrs))
	if err := transferERC20(cli, privateKey, nonce, bterc20Addr, erc20Addr, big.NewInt(accBalance), addrs); err != nil {
		log.Println(fmt.Errorf("failed to transfer ERC20, error: %s", err))
		return
	}

	log.Printf("Finish! ERC20 Address: %s\n", erc20Addr)
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

func deployBTNative(cli client.Client, privateKey *ecdsa.PrivateKey, nonce uint64) (ethcmn.Address, error) {
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

func deployERC20(cli client.Client, privateKey *ecdsa.PrivateKey, nonce uint64) (ethcmn.Address, error) {
	// Query GasPrice
	var gasPrice *big.Int
	ethClient, ok := cli.(*client.EthClient)
	if ok {
		gasPrice = getGasPrice(ethClient.Client)
	}

	txhash, err := cli.CreateContract(privateKey, nonce, nil, 10000000, gasPrice, ethcmn.Hex2Bytes(ERC20Hex))
	if err != nil {
		return ethcmn.Address{}, err
	}
	contractAddr := crypto.CreateAddress(common.GetEthAddressFromPK(privateKey), nonce)
	log.Printf("ERC20 contract: caller: %s, nonce: %d, contract: %s, txhash: %s\n", common.GetEthAddressFromPK(privateKey), nonce, contractAddr, txhash)

	return contractAddr, nil
}

func deployBTERC20(cli client.Client, privateKey *ecdsa.PrivateKey, nonce uint64) (ethcmn.Address, error) {
	// Query GasPrice
	var gasPrice *big.Int
	ethClient, ok := cli.(*client.EthClient)
	if ok {
		gasPrice = getGasPrice(ethClient.Client)
	}

	txhash, err := cli.CreateContract(privateKey, nonce, nil, 500000, gasPrice, ethcmn.Hex2Bytes(BatchTransferHex))
	if err != nil {
		return ethcmn.Address{}, err
	}
	contractAddr := crypto.CreateAddress(common.GetEthAddressFromPK(privateKey), nonce)
	log.Printf("BatchTransfer for ERC20: caller: %s, nonce: %d, contract: %s, txhash: %s\n", common.GetEthAddressFromPK(privateKey), nonce, contractAddr, txhash)

	return contractAddr, nil
}

func sendApprove(cli client.Client, privateKey *ecdsa.PrivateKey, nonce uint64, to ethcmn.Address, totalSupply *big.Int, bterc20Addrs ethcmn.Address) error {
	// Query GasPrice
	var gasPrice *big.Int
	ethClient, ok := cli.(*client.EthClient)
	if ok {
		gasPrice = getGasPrice(ethClient.Client)
	}

	// load abi
	tABI, err := abi.JSON(strings.NewReader(ERC20ABI))
	if err != nil {
		return fmt.Errorf("failed to initialize BatchTransfer abi, error: %s", err)
	}
	txdata, err := tABI.Pack("approve", bterc20Addrs, totalSupply)
	txhash, err := cli.SendEthereumTx(privateKey, nonce, to, nil, uint64(3000000), gasPrice, txdata)
	if err != nil {
		return err
	}
	log.Printf("Approve to BatchTransfer: caller: %s, nonce: %d, txhash: %s\n", common.GetEthAddressFromPK(privateKey), nonce, txhash)
	return nil
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
		log.Printf("[BatchTransfer Native] caller: %s, nonce: %d, to[%d:%d] txhash: %s\n", common.GetEthAddressFromPK(privateKey), nonce, start, end-1, txhash)

		nonce++
		//time.Sleep(time.Second)
	}

	return nil
}

func transferERC20(cli client.Client, privateKey *ecdsa.PrivateKey, nonce uint64, bterc20Addr, tokenAddr ethcmn.Address, amount *big.Int, addrs []ethcmn.Address) error {
	// load abi
	tABI, err := abi.JSON(strings.NewReader(BatchTransferABI))
	if err != nil {
		return fmt.Errorf("failed to initialize BatchTransferERC20 abi, error: %s", err)
	}

	batchNum := 200 // 40,000 gas per address

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
		txdata, err := tABI.Pack("batchTransferERC20", addrs[start:end], tokenAddr, amount)
		if err != nil {
			return fmt.Errorf("failed to pack BatchTransferERC20 parameters, error: %s", err)
		}
		txhash, err := cli.SendEthereumTx(privateKey, nonce, bterc20Addr, nil, uint64(100000*batchNum), gasPrice, txdata)
		if err != nil {
			return err
		}
		log.Printf("[BatchTransfer ERC20] caller: %s, nonce: %d, to[%d:%d] txhash: %s\n", common.GetEthAddressFromPK(privateKey), nonce, start, end-1, txhash)

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
