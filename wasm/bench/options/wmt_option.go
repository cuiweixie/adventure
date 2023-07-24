package options

type WmtContract struct {
	A           string `json:"Atoken"`
	B           string `json:"BToken"`
	Factory     string `json:"factoryAddress"`
	Pair        string `json:"pairAddress"`
	Router      string `json:"routerAddress"`
	RewardToken string `json:"mockRewardTokenAddress"`
	ABLPToken   string `json:"mockStakeTokenAddress"`
	LPStaking   string `json:"stakingAddress"`
}

type CWWMTOption struct {
	PrivateKeysFile    string        `json:"PrivateKeysFile"`
	RestUrls           []string      `json:"restUrls"`
	TendermintUrls     []string      `json:"tendermintUrls"`
	ContractAvailable  bool          `json:"contractAvailable"`
	ContractSetNum     int           `json:"contractSetNum"`
	ContractFolderPath string        `json:"contractFolderPath"`
	WmtContracts       []WmtContract `json:"wmtContracts"`
	ConcurrentNum      int           `json:"concurrentNum"`
	Threshold          int           `json:"threshold"`
	ChainId            string        `json:"chainId"`
}
