package options

type CW20TransferOption struct {
	PrivateKeysFile string   `json:"privateKeysFile"`
	RestUrls        []string `json:"restUrls"`
	TendermintUrls  []string `json:"tendermintUrls"`
	CW20Address     string   `json:"cw20Address"`
	WasmFilePath    string   `json:"wasmFilePath"`
	ConcurrentNum   int      `json:"concurrentNum"`
	Threshold       int      `json:"threshold"`
	ChainId         string   `json:"chainId"`
}
