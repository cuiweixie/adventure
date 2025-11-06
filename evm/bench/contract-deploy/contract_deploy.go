package contractdeploy

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"os"
	"strings"

	ethcmm "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/okex/adventure/common"
	"github.com/okex/adventure/evm/bench/utils"
	"github.com/okex/adventure/evm/config"
)

var (
	// configurable gasPrice
	gasPrice = new(big.Int).SetUint64(1)
	// global variables
	eParam utils.TxParam
)

func contractDeploy(cmd *cobra.Command, args []string) {
	if configPath == "" {
		panic(errors.New("configPath must be not empty"))
	}

	// 获取合约字节码
	bytecode, err := getContractBytecode()
	if err != nil {
		panic(err)
	}

	if err := loadConfig(configPath); err != nil {
		panic(err)
	}

	// 合约部署交易：to地址为nil，data为合约字节码
	eParam = utils.NewTxParam(
		ethcmm.Address{},           // 合约部署时to地址为空
		nil,                        // 不发送ETH
		gasLimit,                   // 使用指定的gas limit
		gasPrice,                   // 使用配置的gas price
		ethcmm.Hex2Bytes(bytecode), // 合约字节码
	)

	utils.RunTxs(
		utils.DefaultBaseParamFromFlag(),
		func(_ ethcmm.Address) []utils.TxParam {
			return []utils.TxParam{eParam}
		},
	)
}

// getContractBytecode 获取合约字节码，优先从文件读取，其次从命令行参数读取
func getContractBytecode() (string, error) {
	// 优先从文件读取
	if bytecodeFile != "" {
		data, err := ioutil.ReadFile(bytecodeFile)
		if err != nil {
			return "", err
		}
		bytecode := strings.TrimSpace(string(data))
		// 移除可能的0x前缀
		if strings.HasPrefix(bytecode, "0x") {
			bytecode = bytecode[2:]
		}
		return bytecode, nil
	}

	// 从命令行参数读取
	if contractBytecode != "" {
		bytecode := strings.TrimSpace(contractBytecode)
		// 移除可能的0x前缀
		if strings.HasPrefix(bytecode, "0x") {
			bytecode = bytecode[2:]
		}
		return bytecode, nil
	}

	return "", errors.New("contract bytecode must be provided either via --bytecode or --bytecode-file")
}

func loadConfig(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	defer file.Close()

	if err := json.Unmarshal(data, &config.TransferCfg); err != nil {
		return err
	}

	privateKeys := common.ReadDataFromFile(config.TransferCfg.AccountsFilePath)
	config.TransferCfg.PrivateKeys = privateKeys
	gasPrice = utils.ParseGasPriceToBigInt(config.TransferCfg.GasPrice, 9)

	return nil
}

