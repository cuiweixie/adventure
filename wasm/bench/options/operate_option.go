package options

type CWOperateOption struct {
	PrivateKeysFile    string   `json:"PrivateKeysFile"`
	RestUrls           []string `json:"restUrls"`
	TendermintUrls     []string `json:"tendermintUrls"`
	ContractAddress    string   `json:"contractAddress"`
	ContractPath       string   `json:"contractPath"`
	RouterContractPath string   `json:"routerContractPath"`
	WasmOperType       string   `json:"WasmOperType"`
	WasmOperRouter     bool     `json:"WasmOperRouter"`
	ConcurrentNum      int      `json:"concurrentNum"`
	Threshold          int      `json:"threshold"`
	ChainId            string   `json:"chainId"`
}
