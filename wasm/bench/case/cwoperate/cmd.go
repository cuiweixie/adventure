package cwoperate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/okex/adventure/wasm/bench/options"
)

type cwoperateCmd struct {
	configPath string
	option     *options.CWOperateOption
}

func NewCWOperateCmd() *cwoperateCmd {
	return &cwoperateCmd{}
}

func (b *cwoperateCmd) NewBenchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cwoperate",
		Short: "operate wasm RO,WO and EO pressure test",
		Run:   b.Run,
	}

	cmd.Flags().StringVar(&b.configPath, "f", "", "the location of transfer config file")
	return cmd
}

func (b *cwoperateCmd) NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cwoperate-config",
		Short: "Generate a empty config for operate bench",
		Run: func(cmd *cobra.Command, args []string) {
			option := options.CWTransferOption{}
			bytes, _ := json.Marshal(option)
			fmt.Println(string(bytes))
		},
	}
	return cmd
}

func (b *cwoperateCmd) Run(cmd *cobra.Command, args []string) {
	if b.configPath == "" {
		panic(errors.New("configPath must be not empty "))
	}

	err := b.LoadConfig()
	if err != nil {
		panic(err)
	}

	bench, err := NewCWOperateBench(b.option)
	if err != nil {
		panic(err)
	}

	bench.StartBench()
}

func (b *cwoperateCmd) LoadConfig() error {
	file, err := os.Open(b.configPath)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	defer file.Close()

	var option options.CWOperateOption

	if err := json.Unmarshal(data, &option); err != nil {
		return err
	}

	b.option = &option

	return nil
}
