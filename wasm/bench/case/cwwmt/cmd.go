package cwwmt

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/okex/adventure/wasm/bench/options"
)

type cwwmtCmd struct {
	configPath string
	option     *options.CWWMTOption
}

func NewCWWMTCmd() *cwwmtCmd {
	return &cwwmtCmd{}
}

func (b *cwwmtCmd) NewBenchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wmt",
		Short: "send multi-set of wmt txs for pressure test",
		Run:   b.Run,
	}

	cmd.Flags().StringVar(&b.configPath, "f", "", "the location of transfer config file")
	return cmd
}

func (b *cwwmtCmd) NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wmt-config",
		Short: "Generate a empty config for wmt bench",
		Run: func(cmd *cobra.Command, args []string) {
			option := options.CWWMTOption{}
			bytes, _ := json.Marshal(option)
			fmt.Println(string(bytes))
		},
	}
	return cmd
}

func (b *cwwmtCmd) Run(cmd *cobra.Command, args []string) {
	if b.configPath == "" {
		panic(errors.New("configPath must be not empty "))
	}

	err := b.LoadConfig()
	if err != nil {
		panic(err)
	}

	bench, err := NewCWWMTBench(b.option)
	if err != nil {
		panic(err)
	}

	bench.StartBench()
}

func (b *cwwmtCmd) LoadConfig() error {
	file, err := os.Open(b.configPath)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	defer file.Close()

	var option options.CWWMTOption

	if err := json.Unmarshal(data, &option); err != nil {
		return err
	}

	b.option = &option

	return nil
}
