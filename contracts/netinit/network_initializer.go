package netinit

import (
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/setherplatform/sether-node/contracts/abis"
	"github.com/setherplatform/sether-node/utils"
	"math/big"
)

func InitializeAll(
	sealedEpoch idx.Epoch,
	totalSupply *big.Int,
	owner common.Address,
) []byte {
	data, _ := abis.NetworkInitializer.Pack(
		"initializeAll",
		utils.U64toBig(uint64(sealedEpoch)),
		totalSupply,
		owner,
	)
	return data
}
