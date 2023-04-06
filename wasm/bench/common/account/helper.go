package account

import (
	"github.com/okex/adventure/common"
)

func ParseAccountsFromFile(file string) ([]*Account, error) {
	strs := common.ReadDataFromFile(file)
	return ParseAccountsFromSlice(strs)
}

func ParseAccountsFromSlice(strs []string) ([]*Account, error) {
	accounts := make([]*Account, 0, len(strs))
	for _, str := range strs {
		acc, err := NewAccount(str)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	return accounts, nil
}
