package cw20

import (
	"encoding/json"
	"errors"
	"github.com/okex/adventure/wasm/bench/options"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

type cw20Case struct {
	configPath string
}

func NewCW20Case() *cw20Case {
	return &cw20Case{}
}

func (b *cw20Case) NewBenchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-cw20",
		Short: "send cw20 token to address",
		Run:   b.run,
	}

	cmd.Flags().StringVar(&b.configPath, "f", "", "the location of transfer config file")
	return cmd
}

func (b *cw20Case) NewConfigCmd() *cobra.Command {
	return nil
}

func (b *cw20Case) run(cmd *cobra.Command, args []string) {
	if b.configPath == "" {
		panic(errors.New("configPath must be not empty "))
	}

	option, err := loadConfig(b.configPath)
	if err != nil {
		panic(err)
	}

	bench, err := NewCW20Bench(option)
	if err != nil {
		panic(err)
	}

	bench.StartBench()
}

func loadConfig(configPath string) (*options.CW20TransferOption, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var option options.CW20TransferOption

	if err := json.Unmarshal(data, &option); err != nil {
		return nil, err
	}

	return &option, nil
}
