package config

var (
	TransferCfg TransferConfig
)

type TransferConfig struct {
	Rpc              []string `json:"rpc"`
	TenderMint       []string `json:"tenderMint"`
	Concurrency      int      `json:"concurrency"`
	Threshold        int      `json:"threshold"`
	AccountsFilePath string   `json:"accountsFilePath"`
	PrivateKeys      []string
	GasPrice         float64 `json:"gasprice"`
}
