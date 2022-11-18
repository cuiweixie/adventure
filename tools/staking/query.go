package staking

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/okex/adventure/common/client"
	sdk "github.com/okex/exchain/libs/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

const (
	flagMode = "mode"
	flagIp   = "ip"
)

const (
	unboding uint8 = iota
	unbonded
	bonded
	all
)

func QueryValidatorsCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:  "staking-query",
		Long: "query validators",
		RunE: runQueryScript,
	}
	queryCmd.Flags().Uint8P(flagMode, "m", all, "the mode of query vals")
	queryCmd.Flags().StringP(flagIp, "i", "http://exchaintmrpc.okex.org", "the ip of cosmos port")
	return queryCmd
}

func runQueryScript(cmd *cobra.Command, args []string) error {
	ip, err := cmd.Flags().GetString(flagIp)
	if err != nil {
		return err
	}
	cli, err := client.NewCosmosClient(ip)
	if err != nil {
		return err
	}

	mode, err := cmd.Flags().GetUint8(flagMode)
	if err != nil {
		return err
	}
	valid := checkFlagModeValid(mode)
	if !valid {
		return fmt.Errorf("flag mode %d is invaild", mode)
	}

	vals, err := cli.Staking().QueryValidators()
	if err != nil {
		return err
	}

	total := sdk.ZeroDec()
	for _, val := range vals {
		if val.Status == sdk.BondStatus(mode) || all == mode {
			total = total.Add(val.DelegatorShares)
		}
	}
	base, _ := sdk.NewDecFromStr("100")

	var vlist validators
	for _, val := range vals {
		if val.Status == sdk.BondStatus(mode) || all == mode {
			bechConsPubkey, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, val.ConsPubKey)
			if err != nil {
				return err
			}

			consAddr := sdk.GetConsAddress(val.ConsPubKey)
			v := &validator{
				status:      byte(val.Status),
				jailed:      val.Jailed,
				shareRate:   val.DelegatorShares.Quo(total).Mul(base),
				shares:      val.DelegatorShares,
				name:        val.Description.Moniker,
				operAddr:    val.OperatorAddress.String(),
				consAddr:    consAddr.String(),
				consPubkey:  bechConsPubkey,
				fingerPrint: consAddr[:],
			}
			vlist = append(vlist, v)
		}
	}

	if len(vlist) == 0 {
		fmt.Println("failed to query validators, the result is nil")
	} else {
		sort.Sort(vlist)
		fmt.Println("val num:", vlist.Len())
		fmt.Println(vlist.String())
	}

	return nil
}

func checkFlagModeValid(mode uint8) bool {
	if mode == unboding || mode == unbonded || mode == bonded || mode == all {
		return true
	}
	return false
}

type validator struct {
	status      byte //unbonding, unboned, bonded
	jailed      bool
	shareRate   sdk.Dec
	shares      sdk.Dec
	name        string
	operAddr    string
	consAddr    string
	consPubkey  string
	fingerPrint sdk.ConsAddress
}

func (val *validator) String() string {
	//return fmt.Sprintf(`%s: %s`, val.name, val.shares)
	return fmt.Sprintf(`Validator
  Name:                       %s
  Jailed:                     %v
  Status:                     %d
  Share Percentage:           %s
  Delegator Shares:           %s
  Operator Address:           %s
  Validator Consensus Address:%s
  Validator Consensus Pubkey: %s
  FingerPrint:                %v`,
		val.name, val.jailed, val.status, val.shareRate.String()+"%", val.shares, val.operAddr, val.consAddr, val.consPubkey, val.fingerPrint)
}

type validators []*validator

func (vals validators) Len() int {
	return len(vals)
}

func (vals validators) Less(i, j int) bool {
	if !(vals[i].jailed != vals[j].jailed) { //bool XNOR expression
		if vals[i].shares.GT(vals[j].shares) {
			return true
		} else if vals[i].shares.Equal(vals[j].shares) {
			if strings.Compare(vals[i].name, vals[j].name) <= 0 {
				return true
			}
			return false
		} else {
			return false
		}
	} else if vals[i].jailed == true {
		return false
	} else {
		return true
	}
}

func (vals validators) Swap(i, j int) {
	vals[i], vals[j] = vals[j], vals[i]
}

func (vals validators) String() (out string) {
	out = ""
	for i, val := range vals {
		out += strconv.Itoa(i+1) + "  " + val.String() + "\n"
	}
	return strings.TrimSpace(out)
}
