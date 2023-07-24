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
			addr:    "0x020A61B6922FEe79C0A8f0A5E342eae359cC1733",
			pubkey:  "0xc0043a2d7348d04d235d381240cee43b9f0a042422223bc9068b3a4a106a6d39cb41c97648f703ebfa6444c104e3ec116d761873360db7fd79d2072af53a447519af",
			stake:   utils2.ToSeth(100_000),
			balance: utils2.ToSeth(10_000_000),
		},
		{
			addr:    "0x11eF7f4B6B10b1a70E33B54cDe7b87F7e13e5400",
			pubkey:  "0xc00403d3747695ba5806f7222d0aaacc9c44c238c6a5cfba55aadd53a414b42215142ac2a03d6597c59e22eb155ce2402d6fba0b9a3bb70e57c7409667050ae192ab",
			stake:   utils2.ToSeth(100_000),
			balance: utils2.ToSeth(10_000_000),
		},
		{
			addr:    "0xF1864f7268F51D4a0D472dEFD80B697FCCbff99D",
			pubkey:  "0xc004e10005419e6baeeaf780e89035bdffaea88f9e2565a7cfb07fa715a573da7f50ec67669decad1dba586c2b6812e095fe886c2ee275e43780dbac6ceba6e7104c",
			stake:   utils2.ToSeth(100_000),
			balance: utils2.ToSeth(10_000_000),
		},
	}

	TestnetAccounts = []GenesisAccount{
		{
			addr:    "0x32f6CCc1aBFb13c2515d403D9e80eD2205E57Af0",
			balance: utils2.ToSeth(20_000_000),
		},
		{
			addr:    "0x15C13D817c531035930fE24a68a6D1CC1A2eA4ed",
			balance: utils2.ToSeth(20_000_000),
		},
		{
			addr:    "0x619a877Db76824734B066f4275CDa78E02910567",
			balance: utils2.ToSeth(20_000_000),
		},
	}

	MainnetValidators = []GenesisValidator{}
	MainnetAccounts   = []GenesisAccount{}

	GenesisTime = inter.FromUnix(1690179442)
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
