package options

type TransferOption struct {
	PrivateKeysFile string   `json:"PrivateKeysFile"`
	RestUrls        []string `json:"restUrls"`
	TendermintUrls  []string `json:"tendermintUrls"`
	CW20Address     string   `json:"CW20Address"`
	ConcurrentNum   int      `json:"concurrentNum"`
	Threshold       int      `json:"threshold"`
	ChainId         string   `json:"chainId"`
}
