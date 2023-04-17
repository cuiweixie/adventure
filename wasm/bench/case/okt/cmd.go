package okt

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/okex/adventure/wasm/bench/options"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

type oktCmd struct {
	configPath string
	option     *options.CW20TransferOption
}

func NewCW20Cmd() *oktCmd {
	return &oktCmd{}
}

func (b *oktCmd) NewBenchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "okt",
		Short: "send native token to address",
		Run:   b.Run,
	}

	cmd.Flags().StringVar(&b.configPath, "f", "", "the location of transfer config file")
	return cmd
}

func (b *oktCmd) NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "okt-config",
		Short: "Generate a empty config for okt bench",
		Run: func(cmd *cobra.Command, args []string) {
			option := options.CW20TransferOption{}
			bytes, _ := json.Marshal(option)
			fmt.Println(string(bytes))
		},
	}
	return cmd
}

func (b *oktCmd) Run(cmd *cobra.Command, args []string) {
	if b.configPath == "" {
		panic(errors.New("configPath must be not empty "))
	}

	err := b.LoadConfig()
	if err != nil {
		panic(err)
	}

	bench, err := NewOKTBench(b.option)
	if err != nil {
		panic(err)
	}

	bench.StartBench()
}

func (b *oktCmd) LoadConfig() error {
	file, err := os.Open(b.configPath)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	defer file.Close()

	var option options.CW20TransferOption

	if err := json.Unmarshal(data, &option); err != nil {
		return err
	}

	b.option = &option

	return nil
}
