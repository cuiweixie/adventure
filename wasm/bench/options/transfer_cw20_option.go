package options

type CWTransferOption struct {
	PrivateKeysFile string   `json:"privateKeysFile"`
	RestUrls        []string `json:"restUrls"`
	TendermintUrls  []string `json:"tendermintUrls"`
	ContractAddress string   `json:"contractAddress"`
	WasmFilePath    string   `json:"wasmFilePath"`
	ConcurrentNum   int      `json:"concurrentNum"`
	Threshold       int      `json:"threshold"`
	ChainId         string   `json:"chainId"`
}
