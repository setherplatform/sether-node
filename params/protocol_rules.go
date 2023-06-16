package params

import (
	"encoding/json"
	"github.com/setherplatform/sether-node/contracts"
	"math/big"
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	ethparams "github.com/ethereum/go-ethereum/params"

	"github.com/setherplatform/sether-node/contracts/evmwriter"
	"github.com/setherplatform/sether-node/inter"
)

const (
	MainNetworkID   uint64 = 22222
	TestNetworkID   uint64 = 22223
	FakeNetworkID   uint64 = 22224
	DefaultEventGas uint64 = 28000
	berlinBit              = 1 << 0
	londonBit              = 1 << 1
	llrBit                 = 1 << 2
)

var DefaultVMConfig = vm.Config{
	StatePrecompiles: map[common.Address]vm.PrecompiledStateContract{
		contracts.EvmWriterSmartContractAddress: &evmwriter.PreCompiledContract{},
	},
}

type ProtocolRulesRLP struct {
	Name      string
	NetworkID uint64

	// Graph options
	Dag DagRules

	// Epochs options
	Epochs EpochsRules

	// Blockchain options
	Blocks BlocksRules

	// Economy options
	Economy EconomyRules

	Upgrades Upgrades `rlp:"-"`
}

// ProtocolRules describes sether net.
// Note keep track of all the non-copiable variables in Copy()
type ProtocolRules ProtocolRulesRLP

// GasPowerRules defines gas power rules in the consensus.
type GasPowerRules struct {
	AllocPerSec        uint64
	MaxAllocPeriod     inter.Timestamp
	StartupAllocPeriod inter.Timestamp
	MinStartupGas      uint64
}

type GasRulesRLPV1 struct {
	MaxEventGas  uint64
	EventGas     uint64
	ParentGas    uint64
	ExtraDataGas uint64
	// Post-LLR fields
	BlockVotesBaseGas    uint64
	BlockVoteGas         uint64
	EpochVoteGas         uint64
	MisbehaviourProofGas uint64
}

type GasRules GasRulesRLPV1

type EpochsRules struct {
	MaxEpochGas      uint64
	MaxEpochDuration inter.Timestamp
}

// DagRules of Lachesis DAG (directed acyclic graph).
type DagRules struct {
	MaxParents     idx.Event
	MaxFreeParents idx.Event // maximum number of parents with no gas cost
	MaxExtraData   uint32
}

// BlocksMissed is information about missed blocks from a staker
type BlocksMissed struct {
	BlocksNum idx.Block
	Period    inter.Timestamp
}

// EconomyRules contains economy constants
type EconomyRules struct {
	BlockMissedSlack idx.Block

	Gas GasRules

	MinGasPrice *big.Int

	ShortGasPower GasPowerRules
	LongGasPower  GasPowerRules
}

// BlocksRules contains blocks constants
type BlocksRules struct {
	MaxBlockGas             uint64 // technical hard limit, gas is mostly governed by gas power allocation
	MaxEmptyBlockSkipPeriod inter.Timestamp
}

type Upgrades struct {
	Berlin bool
	London bool
	Llr    bool
}

type UpgradeHeight struct {
	Upgrades Upgrades
	Height   idx.Block
}

// EvmChainConfig returns ChainConfig for transactions signing and execution
func (r ProtocolRules) EvmChainConfig(hh []UpgradeHeight) *ethparams.ChainConfig {
	cfg := *ethparams.AllEthashProtocolChanges
	cfg.ChainID = new(big.Int).SetUint64(r.NetworkID)
	cfg.BerlinBlock = nil
	cfg.LondonBlock = nil
	for i, h := range hh {
		height := new(big.Int)
		if i > 0 {
			height.SetUint64(uint64(h.Height))
		}
		if cfg.BerlinBlock == nil && h.Upgrades.Berlin {
			cfg.BerlinBlock = height
		}
		if !h.Upgrades.Berlin {
			cfg.BerlinBlock = nil
		}

		if cfg.LondonBlock == nil && h.Upgrades.London {
			cfg.LondonBlock = height
		}
		if !h.Upgrades.London {
			cfg.LondonBlock = nil
		}
	}
	return &cfg
}

func MainNetRules() ProtocolRules {
	return ProtocolRules{
		Name:      "main",
		NetworkID: MainNetworkID,
		Dag:       DefaultDagRules(),
		Epochs:    DefaultEpochsRules(),
		Economy:   DefaultEconomyRules(),
		Blocks: BlocksRules{
			MaxBlockGas:             20_500_000,
			MaxEmptyBlockSkipPeriod: inter.Timestamp(1 * time.Minute),
		},
		Upgrades: Upgrades{
			Berlin: true,
			London: true,
			Llr:    true,
		},
	}
}

func TestNetRules() ProtocolRules {
	return ProtocolRules{
		Name:      "test",
		NetworkID: TestNetworkID,
		Dag:       DefaultDagRules(),
		Epochs:    DefaultEpochsRules(),
		Economy:   DefaultEconomyRules(),
		Blocks: BlocksRules{
			MaxBlockGas:             20_500_000,
			MaxEmptyBlockSkipPeriod: inter.Timestamp(1 * time.Minute),
		},
		Upgrades: Upgrades{
			Berlin: true,
			London: true,
			Llr:    true,
		},
	}
}

func FakeNetRules() ProtocolRules {
	return ProtocolRules{
		Name:      "fake",
		NetworkID: FakeNetworkID,
		Dag:       DefaultDagRules(),
		Epochs:    FakeNetEpochsRules(),
		Economy:   FakeEconomyRules(),
		Blocks: BlocksRules{
			MaxBlockGas:             20_500_000,
			MaxEmptyBlockSkipPeriod: inter.Timestamp(3 * time.Second),
		},
		Upgrades: Upgrades{
			Berlin: true,
			London: true,
			Llr:    true,
		},
	}
}

// DefaultEconomyRules returns mainnet economy
func DefaultEconomyRules() EconomyRules {
	return EconomyRules{
		BlockMissedSlack: 50,
		Gas:              DefaultGasRules(),
		MinGasPrice:      big.NewInt(1e9), // 1 gwei
		ShortGasPower:    DefaultShortGasPowerRules(),
		LongGasPower:     DefaulLongGasPowerRules(),
	}
}

// FakeEconomyRules returns fakenet economy
func FakeEconomyRules() EconomyRules {
	cfg := DefaultEconomyRules()
	cfg.ShortGasPower = FakeShortGasPowerRules()
	cfg.LongGasPower = FakeLongGasPowerRules()
	return cfg
}

func DefaultDagRules() DagRules {
	return DagRules{
		MaxParents:     10,
		MaxFreeParents: 3,
		MaxExtraData:   128,
	}
}

func DefaultEpochsRules() EpochsRules {
	return EpochsRules{
		MaxEpochGas:      1500000000,
		MaxEpochDuration: inter.Timestamp(4 * time.Hour),
	}
}

func DefaultGasRules() GasRules {
	return GasRules{
		MaxEventGas:          10000000 + DefaultEventGas,
		EventGas:             DefaultEventGas,
		ParentGas:            2400,
		ExtraDataGas:         25,
		BlockVotesBaseGas:    1024,
		BlockVoteGas:         512,
		EpochVoteGas:         1536,
		MisbehaviourProofGas: 71536,
	}
}

func FakeNetEpochsRules() EpochsRules {
	cfg := DefaultEpochsRules()
	cfg.MaxEpochGas /= 5
	cfg.MaxEpochDuration = inter.Timestamp(1 * time.Minute)
	return cfg
}

// DefaulLongGasPowerRules is long-window config
func DefaulLongGasPowerRules() GasPowerRules {
	return GasPowerRules{
		AllocPerSec:        100 * DefaultEventGas,
		MaxAllocPeriod:     inter.Timestamp(60 * time.Minute),
		StartupAllocPeriod: inter.Timestamp(5 * time.Second),
		MinStartupGas:      DefaultEventGas * 20,
	}
}

// DefaultShortGasPowerRules is short-window config
func DefaultShortGasPowerRules() GasPowerRules {
	// 2x faster allocation rate, 6x lower max accumulated gas power
	cfg := DefaulLongGasPowerRules()
	cfg.AllocPerSec *= 2
	cfg.StartupAllocPeriod /= 2
	cfg.MaxAllocPeriod /= 2 * 6
	return cfg
}

// FakeLongGasPowerRules is fake long-window config
func FakeLongGasPowerRules() GasPowerRules {
	config := DefaulLongGasPowerRules()
	config.AllocPerSec *= 1000
	return config
}

// FakeShortGasPowerRules is fake short-window config
func FakeShortGasPowerRules() GasPowerRules {
	config := DefaultShortGasPowerRules()
	config.AllocPerSec *= 1000
	return config
}

func (r ProtocolRules) Copy() ProtocolRules {
	cp := r
	cp.Economy.MinGasPrice = new(big.Int).Set(r.Economy.MinGasPrice)
	return cp
}

func (r ProtocolRules) String() string {
	b, _ := json.Marshal(&r)
	return string(b)
}

func UpdateRules(src ProtocolRules, diff []byte) (res ProtocolRules, err error) {
	changed := src.Copy()
	err = json.Unmarshal(diff, &changed)
	if err != nil {
		return src, err
	}
	// protect readonly fields
	res = changed
	res.NetworkID = src.NetworkID
	res.Name = src.Name
	return
}
