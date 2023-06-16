package makefakegenesis

import (
	"crypto/ecdsa"
	"github.com/setherplatform/sether-node/contracts/driver"
	"github.com/setherplatform/sether-node/params"
	"math/big"
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/setherplatform/sether-node/evmcore"
	"github.com/setherplatform/sether-node/genesis"
	"github.com/setherplatform/sether-node/genesis/genesisstore"
	"github.com/setherplatform/sether-node/genesis/gpos"
	"github.com/setherplatform/sether-node/genesis/makegenesis"
	"github.com/setherplatform/sether-node/inter"
	"github.com/setherplatform/sether-node/inter/validatorpk"
)

var (
	FakeGenesisTime = inter.Timestamp(time.Now().UnixNano())
)

// FakeKey gets n-th fake private key.
func FakeKey(n idx.ValidatorID) *ecdsa.PrivateKey {
	return evmcore.FakeKey(int(n))
}

func FakeGenesisStore(num idx.Validator, balance, stake *big.Int) *genesisstore.Store {
	return FakeGenesisStoreWithRulesAndStart(num, balance, stake, params.FakeNetRules(), 2, 1)
}

func FakeGenesisStoreWithRulesAndStart(num idx.Validator, balance, stake *big.Int, rules params.ProtocolRules, epoch idx.Epoch, block idx.Block) *genesisstore.Store {
	builder := makegenesis.NewGenesisBuilder(memorydb.NewProducer(""))

	validators := GetFakeValidators(num)

	// add balances to validators
	var delegations []driver.Delegation
	for _, val := range validators {
		builder.AddBalance(val.Address, balance)
		delegations = append(delegations, driver.Delegation{
			Address:            val.Address,
			ValidatorID:        val.ID,
			Stake:              stake,
			LockedStake:        new(big.Int),
			LockupFromEpoch:    0,
			LockupEndTime:      0,
			LockupDuration:     0,
			EarlyUnlockPenalty: new(big.Int),
			Rewards:            new(big.Int),
		})
	}

	builder.DeployBaseContracts()
	builder.InitializeEpoch(block, epoch, rules, FakeGenesisTime)

	var owner common.Address
	if num != 0 {
		owner = validators[0].Address
	}

	blockProc := makegenesis.DefaultBlockProc()
	genesisTxs := builder.GetGenesisTxs(epoch-2, validators, builder.TotalSupply(), delegations, owner)
	err := builder.ExecuteGenesisTxs(blockProc, genesisTxs)
	if err != nil {
		panic(err)
	}

	return builder.Build(genesis.Header{
		GenesisID:   builder.CurrentHash(),
		NetworkID:   rules.NetworkID,
		NetworkName: rules.Name,
	})
}

func GetFakeValidators(num idx.Validator) gpos.Validators {
	validators := make(gpos.Validators, 0, num)

	for i := idx.ValidatorID(1); i <= idx.ValidatorID(num); i++ {
		key := FakeKey(i)
		addr := crypto.PubkeyToAddress(key.PublicKey)
		pubkeyraw := crypto.FromECDSAPub(&key.PublicKey)
		validators = append(validators, gpos.Validator{
			ID:      i,
			Address: addr,
			PubKey: validatorpk.PubKey{
				Raw:  pubkeyraw,
				Type: validatorpk.Types.Secp256k1,
			},
			CreationTime:     FakeGenesisTime,
			CreationEpoch:    0,
			DeactivatedTime:  0,
			DeactivatedEpoch: 0,
			Status:           0,
		})
	}

	return validators
}
