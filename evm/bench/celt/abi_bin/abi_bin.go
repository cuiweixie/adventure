package abi_bin

import (
	_ "embed"
	"github.com/okex/exchain-go-sdk/utils"
)

var (
	//go:embed common_abi
	CommonAbi string
	//go:embed common_bin
	CommonBin string

	//go:embed supreme_abi
	SupremeAbi string
	//go:embed supreme_bin
	SupremeBin string

	//go:embed nftpool_abi
	NftPoolAbi string
	//go:embed nftpool_bin
	NftPoolBin string

	//go:embed invitationcenter_abi
	InvitationCenterAbi string
	//go:embed invitationcenter_bin
	InvitationCenterBin string

	//go:embed linearunlock_abi
	LinearUnlockAbi string
	//go:embed linearunlock_bin
	LinearUnlockBin string
)

var (
	CommonBuilder           utils.PayloadBuilder
	SurpemeBuilder          utils.PayloadBuilder
	NftPoolBuilder          utils.PayloadBuilder
	InvitationCenterBuilder utils.PayloadBuilder
	LinearUnlockBuilder     utils.PayloadBuilder
)

func InitBuilder() {
	var err error

	CommonBuilder, err = utils.NewPayloadBuilder(CommonBin, CommonAbi)
	if err != nil {
		panic(err)
	}

	SurpemeBuilder, err = utils.NewPayloadBuilder(SupremeBin, SupremeAbi)
	if err != nil {
		panic(err)
	}

	NftPoolBuilder, err = utils.NewPayloadBuilder(NftPoolBin, NftPoolAbi)
	if err != nil {
		panic(err)
	}

	InvitationCenterBuilder, err = utils.NewPayloadBuilder(InvitationCenterBin, InvitationCenterAbi)
	if err != nil {
		panic(err)
	}

	LinearUnlockBuilder, err = utils.NewPayloadBuilder(LinearUnlockBin, LinearUnlockAbi)
	if err != nil {
		panic(err)
	}
}
