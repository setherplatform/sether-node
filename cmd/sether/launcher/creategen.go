package launcher

import (
	"fmt"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/setherplatform/sether-node/contracts/driver"
	"github.com/setherplatform/sether-node/genesis"
	"github.com/setherplatform/sether-node/genesis/genesisstore"
	"github.com/setherplatform/sether-node/genesis/gpos"
	"github.com/setherplatform/sether-node/genesis/makegenesis"
	"github.com/setherplatform/sether-node/inter"
	"github.com/setherplatform/sether-node/inter/ibr"
	"github.com/setherplatform/sether-node/inter/ier"
	"github.com/setherplatform/sether-node/inter/validatorpk"
	"github.com/setherplatform/sether-node/params"
	utils2 "github.com/setherplatform/sether-node/utils"
	"github.com/setherplatform/sether-node/utils/iodb"
	"gopkg.in/urfave/cli.v1"
	"io"
	"math/big"
	"os"
)

var genesisTypeFlag = cli.StringFlag{
	Name:  "genesis.type",
	Usage: "Type of genesis to generate: mainnet, testnet",
	Value: "testnet",
}

var (
	createGenesisCommand = cli.Command{
		Action:    utils.MigrateFlags(createGenesisCmd),
		Name:      "creategen",
		Usage:     "Create genesis",
		ArgsUsage: "",
		Category:  "MISCELLANEOUS COMMANDS",
	}
)

type GenesisValidator struct {
	addr    string
	pubkey  string
	stake   *big.Int
	balance *big.Int
}

type GenesisAccount struct {
	addr    string
	balance *big.Int
}

var (
	TestnetValidators = []GenesisValidator{
		{
			addr:    "0xf9215294250dF0D4Beb912C8e18F1e1416d6A398",
			pubkey:  "0xc004135f0d5860bc30341d22cc44f3007d0bf35ee815cc827215c96d7d9aba0fb906c5e8c6eb651cf24625b9151cb2e919259b4f006301292d158e5415d66564b81e",
			stake:   utils2.ToSeth(1_000_000), // min stake StakerConstants.sol -> minSelfStake
			balance: utils2.ToSeth(0),
		},
		{
			addr:    "0x594f344E2B6662C4b6c5A07B6b1287c6209c9c22",
			pubkey:  "0xc00459be00a14b8bd3b249ab7914e44f5c8e01be92aeafc51ec07f523c965cd74bc821cd7389c2ee9d87aaea327c13b6e4027c9cfcc651ef9b52a2adaeb69e6d6142",
			stake:   utils2.ToSeth(1_000_000), // min stake StakerConstants.sol -> minSelfStake
			balance: utils2.ToSeth(0),
		},
		{
			addr:    "0x280620317a56474ABCcc05d7Af612C8D11956611",
			pubkey:  "0xc0047a6230b289af747663ccff2c95edd2061b029b4e888847b2b8aed005e22daafe9a39f1bb5f12044adc25a3e8e733bb7b62088b746eec8ed2fee9e131ece5e907",
			stake:   utils2.ToSeth(1_000_000),
			balance: utils2.ToSeth(0), // min stake StakerConstants.sol -> minSelfStake
		},
	}

	TestnetAccounts = []GenesisAccount{
		{
			addr:    "0x01D4d20f19315D78f5E942029345dad1e85fce55",
			balance: utils2.ToSeth(10_000_000),
		},
		{
			addr:    "0x4a14c36f2A8D73525D44E70Fa2EaA2483A916690",
			balance: utils2.ToSeth(10_000_000),
		},
		{
			addr:    "0x349e543718458B46244f958e7BA4a5c2848F9c78",
			balance: utils2.ToSeth(10_000_000),
		},
	}

	MainnetValidators = []GenesisValidator{}
	MainnetAccounts   = []GenesisAccount{}

	GenesisTime = inter.FromUnix(1677067996)
)

func createGenesisCmd(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		utils.Fatalf("This command requires an argument.")
	}

	genesisType := "testnet"
	if ctx.GlobalIsSet(genesisTypeFlag.Name) {
		genesisType = ctx.GlobalString(genesisTypeFlag.Name)
	}

	fileName := ctx.Args().First()

	fmt.Println("Creating " + genesisType + " genesis")
	genesisStore, currentHash := CreateGenesis(genesisType)
	err := WriteGenesisStore(fileName, genesisStore, currentHash)
	if err != nil {
		return err
	}

	return nil
}

func CreateGenesis(genesisType string) (*genesisstore.Store, hash.Hash) {
	builder := makegenesis.NewGenesisBuilder(memorydb.NewProducer(""))

	validators := make(gpos.Validators, 0, 3)
	delegations := make([]driver.Delegation, 0, 3)

	var initialValidators = TestnetValidators
	var initialAccounts = TestnetAccounts
	if genesisType == "mainnet" {
		initialValidators = MainnetValidators
		initialAccounts = MainnetAccounts
	}

	// add initial validators, premine and lock their stake to get maximum rewards
	for i, v := range initialValidators {
		validators, delegations = AddValidator(
			uint8(i+1),
			v,
			validators, delegations, builder,
		)
	}

	// premine to genesis accounts
	for _, a := range initialAccounts {
		builder.AddBalance(
			common.HexToAddress(a.addr),
			a.balance,
		)
	}

	builder.DeployBaseContracts()

	rules := params.TestNetRules()
	if genesisType == "mainnet" {
		rules = params.MainNetRules()
	}

	builder.InitializeEpoch(1, 2, rules, GenesisTime)

	owner := validators[0].Address
	blockProc := makegenesis.DefaultBlockProc()
	genesisTxs := builder.GetGenesisTxs(0, validators, builder.TotalSupply(), delegations, owner)
	err := builder.ExecuteGenesisTxs(blockProc, genesisTxs)
	if err != nil {
		panic(err)
	}

	return builder.Build(genesis.Header{
		GenesisID:   builder.CurrentHash(),
		NetworkID:   rules.NetworkID,
		NetworkName: rules.Name,
	}), builder.CurrentHash()
}

func AddValidator(
	id uint8,
	v GenesisValidator,
	validators gpos.Validators,
	delegations []driver.Delegation,
	builder *makegenesis.GenesisBuilder,
) (gpos.Validators, []driver.Delegation) {
	validatorId := idx.ValidatorID(id)
	pk, _ := validatorpk.FromString(v.pubkey)
	ecdsaPubkey, _ := crypto.UnmarshalPubkey(pk.Raw)
	addr := crypto.PubkeyToAddress(*ecdsaPubkey)

	validator := gpos.Validator{
		ID:      validatorId,
		Address: addr,
		PubKey: validatorpk.PubKey{
			Raw:  pk.Raw,
			Type: validatorpk.Types.Secp256k1,
		},
		CreationTime:     GenesisTime,
		CreationEpoch:    0,
		DeactivatedTime:  0,
		DeactivatedEpoch: 0,
		Status:           0,
	}
	builder.AddBalance(validator.Address, v.balance)
	validators = append(validators, validator)

	delegations = append(delegations, driver.Delegation{
		Address:            validator.Address,
		ValidatorID:        validator.ID,
		Stake:              v.stake,
		LockedStake:        new(big.Int),
		LockupFromEpoch:    0,
		LockupEndTime:      0,
		LockupDuration:     0,
		EarlyUnlockPenalty: new(big.Int),
		Rewards:            new(big.Int),
	})

	return validators, delegations
}

func WriteGenesisStore(fn string, gs *genesisstore.Store, genesisHash hash.Hash) error {
	var plain io.WriteSeeker

	log.Info("GenesisID ", "hash", genesisHash.String())

	fh, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer fh.Close()
	plain = fh

	writer := newUnitWriter(plain)
	err = writer.Start(gs.Header(), genesisstore.EpochsSection, "/tmp/gentmp")
	if err != nil {
		return err
	}

	gs.Epochs().ForEach(func(epochRecord ier.LlrIdxFullEpochRecord) bool {
		b, _ := rlp.EncodeToBytes(epochRecord)
		_, err := writer.Write(b)
		if err != nil {
			panic(err)
		}
		return true
	})

	var epochsHash hash.Hash
	epochsHash, err = writer.Flush()
	if err != nil {
		return err
	}
	log.Info("Exported epochs", "hash", epochsHash.String())

	writer = newUnitWriter(plain)
	err = writer.Start(gs.Header(), genesisstore.BlocksSection, "/tmp/gentmp")
	if err != nil {
		return err
	}
	gs.Blocks().ForEach(func(blockRecord ibr.LlrIdxFullBlockRecord) bool {
		b, _ := rlp.EncodeToBytes(blockRecord)
		_, err := writer.Write(b)
		if err != nil {
			panic(err)
		}
		return true
	})

	var blocksHash hash.Hash
	blocksHash, err = writer.Flush()
	if err != nil {
		return err
	}
	log.Info("Exported blocks", "hash", blocksHash.String())

	writer = newUnitWriter(plain)
	err = writer.Start(gs.Header(), genesisstore.EvmSection, "/tmp/gentmp")
	if err != nil {
		return err
	}

	gs.RawEvmItems().(genesisstore.RawEvmItems).Iterator(func(it kvdb.Iterator) bool {
		defer it.Release()
		err = iodb.Write(writer, it)
		if err != nil {
			panic(err)
		}
		return true
	})

	var evmHash hash.Hash
	evmHash, err = writer.Flush()
	if err != nil {
		return err
	}
	log.Info("Exported EVM data", "hash", evmHash.String())

	fmt.Printf("Exported genesis to file %s\n", fn)
	return nil
}
