package cwokt

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/okex/adventure/wasm/bench/options"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

type cwoktCmd struct {
	configPath string
	option     *options.CWTransferOption
}

func NewCW20Cmd() *cwoktCmd {
	return &cwoktCmd{}
}

func (b *cwoktCmd) NewBenchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cwokt",
		Short: "send native token to address",
		Run:   b.Run,
	}

	cmd.Flags().StringVar(&b.configPath, "f", "", "the location of transfer config file")
	return cmd
}

func (b *cwoktCmd) NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cwokt-config",
		Short: "Generate a empty config for okt bench",
		Run: func(cmd *cobra.Command, args []string) {
			option := options.CWTransferOption{}
			bytes, _ := json.Marshal(option)
			fmt.Println(string(bytes))
		},
	}
	return cmd
}

func (b *cwoktCmd) Run(cmd *cobra.Command, args []string) {
	if b.configPath == "" {
		panic(errors.New("configPath must be not empty "))
	}

	err := b.LoadConfig()
	if err != nil {
		panic(err)
	}

	bench, err := NewCWOKTBench(b.option)
	if err != nil {
		panic(err)
	}

	bench.StartBench()
}

func (b *cwoktCmd) LoadConfig() error {
	file, err := os.Open(b.configPath)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	defer file.Close()

	var option options.CWTransferOption

	if err := json.Unmarshal(data, &option); err != nil {
		return err
	}

	b.option = &option

	return nil
}
