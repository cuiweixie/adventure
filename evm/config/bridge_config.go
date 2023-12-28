package config

var (
	Bridgecfg BridgeConfig
)

type BridgeConfig struct {
	L1RPC            []string `json:"L1RPC"`
	L2RPC            []string `json:"L2RPC"`
	Concurrency      int      `json:"concurrency"`
	Threshold        int      `json:"threshold"`
	AccountsFilePath string   `json:"accountsFilePath"`
	BridgeAddress    string   `json:"bridgeAddress"`
	OKBTokenAddress  string   `json:"okbtokenAddress"`
	PrivateKeys      []string
}
