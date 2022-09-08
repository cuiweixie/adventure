package utils

import (
	"github.com/okex/adventure/common"
	"github.com/okex/adventure/evm/constant"
	"github.com/spf13/viper"
	"math"
)

type BasepParam struct {
	sleep       int
	concurrency int
	ips         []string

	privateKeys []string
	threshold   int
}

func NewBaseParam(sleep int, concurrency int, ips []string, privateKeyFile string, threshold int) BasepParam {
	privateKeys := constant.PrivateKeys
	if privateKeyFile != "" {
		privateKeys = common.ReadDataFromFile(privateKeyFile)
	}

	if threshold == 0 {
		threshold = math.MaxInt
	}

	return BasepParam{
		sleep,
		concurrency,
		ips,
		privateKeys,
		threshold,
	}
}

func DefaultBaseParamFromFlag() BasepParam {
	return NewBaseParam(
		viper.GetInt(constant.FlagSleep),
		viper.GetInt(constant.FlagConcurrency),
		viper.GetStringSlice(constant.FlagIPs),
		viper.GetString(constant.FlagPrivateKeyFile),
		viper.GetInt(constant.FlagThreshold),
	)
}

func (bParam BasepParam) GetSleep() int {
	return bParam.sleep
}

func (bParam BasepParam) GetConcurrency() int {
	return bParam.concurrency
}

func (bParam BasepParam) GetIPs() []string {
	return bParam.ips
}

func (bParam BasepParam) GetPrivateKeys() []string {
	return bParam.privateKeys
}

func (bParam BasepParam) GetThreshold() int {
	return bParam.threshold
}
