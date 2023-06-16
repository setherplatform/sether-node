package makegenesis

import (
	"bytes"
	"errors"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/setherplatform/sether-node/contracts"
	"github.com/setherplatform/sether-node/contracts/driver"
	"github.com/setherplatform/sether-node/contracts/netinit"
	"github.com/setherplatform/sether-node/genesis/gpos"
	"github.com/setherplatform/sether-node/inter/drivertype"
	"github.com/setherplatform/sether-node/params"
	"io"
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/setherplatform/sether-node/evmcore"
	"github.com/setherplatform/sether-node/genesis"
	"github.com/setherplatform/sether-node/genesis/genesisstore"
	"github.com/setherplatform/sether-node/gossip/blockproc"
	"github.com/setherplatform/sether-node/gossip/blockproc/drivermodule"
	"github.com/setherplatform/sether-node/gossip/blockproc/eventmodule"
	"github.com/setherplatform/sether-node/gossip/blockproc/evmmodule"
	"github.com/setherplatform/sether-node/gossip/blockproc/sealmodule"
	"github.com/setherplatform/sether-node/gossip/evmstore"
	"github.com/setherplatform/sether-node/inter"
	"github.com/setherplatform/sether-node/inter/iblockproc"
	"github.com/setherplatform/sether-node/inter/ibr"
	"github.com/setherplatform/sether-node/inter/ier"
	"github.com/setherplatform/sether-node/utils/iodb"
)

type GenesisBuilder struct {
	dbs kvdb.DBProducer

	tmpEvmStore *evmstore.Store
	tmpStateDB  *state.StateDB

	totalSupply *big.Int

	blocks       []ibr.LlrIdxFullBlockRecord
	epochs       []ier.LlrIdxFullEpochRecord
	currentEpoch ier.LlrIdxFullEpochRecord
}

type BlockProc struct {
	SealerModule     blockproc.SealerModule
	TxListenerModule blockproc.TxListenerModule
	PreTxTransactor  blockproc.TxTransactor
	PostTxTransactor blockproc.TxTransactor
	EventsModule     blockproc.ConfirmedEventsModule
	EVMModule        blockproc.EVM
}

func DefaultBlockProc() BlockProc {
	return BlockProc{
		SealerModule:     sealmodule.New(),
		TxListenerModule: drivermodule.NewDriverTxListenerModule(),
		PreTxTransactor:  drivermodule.NewDriverTxPreTransactor(),
		PostTxTransactor: drivermodule.NewDriverTxTransactor(),
		EventsModule:     eventmodule.New(),
		EVMModule:        evmmodule.New(),
	}
}

func (b *GenesisBuilder) GetStateDB() *state.StateDB {
	if b.tmpStateDB == nil {
		tmpEvmStore := evmstore.NewStore(b.dbs, evmstore.LiteStoreConfig())
		b.tmpStateDB, _ = tmpEvmStore.StateDB(hash.Zero)
	}
	return b.tmpStateDB
}

func (b *GenesisBuilder) AddBalance(acc common.Address, balance *big.Int) {
	b.tmpStateDB.AddBalance(acc, balance)
	b.totalSupply.Add(b.totalSupply, balance)
}

func (b *GenesisBuilder) SetCode(acc common.Address, code []byte) {
	b.tmpStateDB.SetCode(acc, code)
}

func (b *GenesisBuilder) SetNonce(acc common.Address, nonce uint64) {
	b.tmpStateDB.SetNonce(acc, nonce)
}

func (b *GenesisBuilder) SetStorage(acc common.Address, key, val common.Hash) {
	b.tmpStateDB.SetState(acc, key, val)
}

func (b *GenesisBuilder) AddBlock(br ibr.LlrIdxFullBlockRecord) {
	b.blocks = append(b.blocks, br)
}

func (b *GenesisBuilder) AddEpoch(er ier.LlrIdxFullEpochRecord) {
	b.epochs = append(b.epochs, er)
}

func (b *GenesisBuilder) SetCurrentEpoch(er ier.LlrIdxFullEpochRecord) {
	b.currentEpoch = er
}

func (b *GenesisBuilder) GetCurrentEpoch() ier.LlrIdxFullEpochRecord {
	return b.currentEpoch
}

func (b *GenesisBuilder) TotalSupply() *big.Int {
	return b.totalSupply
}

func (b *GenesisBuilder) CurrentHash() hash.Hash {
	er := b.epochs[len(b.epochs)-1]
	return er.Hash()
}

func NewGenesisBuilder(dbs kvdb.DBProducer) *GenesisBuilder {
	tmpEvmStore := evmstore.NewStore(dbs, evmstore.LiteStoreConfig())
	statedb, _ := tmpEvmStore.StateDB(hash.Zero)
	return &GenesisBuilder{
		dbs:         dbs,
		tmpEvmStore: tmpEvmStore,
		tmpStateDB:  statedb,
		totalSupply: new(big.Int),
	}
}

type dummyHeaderReturner struct {
}

func (d dummyHeaderReturner) GetHeader(common.Hash, uint64) *evmcore.EvmHeader {
	return &evmcore.EvmHeader{}
}

func (b *GenesisBuilder) ExecuteGenesisTxs(blockProc BlockProc, genesisTxs types.Transactions) error {
	bs, es := b.currentEpoch.BlockState.Copy(), b.currentEpoch.EpochState.Copy()

	blockCtx := iblockproc.BlockCtx{
		Idx:     bs.LastBlock.Idx + 1,
		Time:    bs.LastBlock.Time + 1,
		Atropos: hash.Event{},
	}

	sealer := blockProc.SealerModule.Start(blockCtx, bs, es)
	sealing := true
	txListener := blockProc.TxListenerModule.Start(blockCtx, bs, es, b.tmpStateDB)
	evmProcessor := blockProc.EVMModule.Start(blockCtx, b.tmpStateDB, dummyHeaderReturner{}, func(l *types.Log) {
		txListener.OnNewLog(l)
	}, es.Rules, params.DefaultVMConfig, es.Rules.EvmChainConfig([]params.UpgradeHeight{
		{
			Upgrades: es.Rules.Upgrades,
			Height:   0,
		},
	}))

	// Execute genesis transactions
	evmProcessor.Execute(genesisTxs)
	bs = txListener.Finalize()

	// Execute pre-internal transactions
	preInternalTxs := blockProc.PreTxTransactor.PopInternalTxs(blockCtx, bs, es, sealing, b.tmpStateDB)
	evmProcessor.Execute(preInternalTxs)
	bs = txListener.Finalize()

	// Seal epoch if requested
	if sealing {
		sealer.Update(bs, es)
		bs, es = sealer.SealEpoch()
		txListener.Update(bs, es)
	}

	// Execute post-internal transactions
	internalTxs := blockProc.PostTxTransactor.PopInternalTxs(blockCtx, bs, es, sealing, b.tmpStateDB)
	evmProcessor.Execute(internalTxs)

	evmBlock, skippedTxs, receipts := evmProcessor.Finalize()
	for _, r := range receipts {
		if r.Status == 0 {
			return errors.New("genesis transaction reverted")
		}
	}
	if len(skippedTxs) != 0 {
		return errors.New("genesis transaction is skipped")
	}
	bs = txListener.Finalize()
	bs.FinalizedStateRoot = hash.Hash(evmBlock.Root)

	bs.LastBlock = blockCtx

	prettyHash := func(root hash.Hash) hash.Event {
		e := inter.MutableEventPayload{}
		// for nice-looking ID
		e.SetEpoch(es.Epoch)
		e.SetLamport(1)
		// actual data hashed
		e.SetExtra(root[:])

		return e.Build().ID()
	}
	receiptsStorage := make([]*types.ReceiptForStorage, len(receipts))
	for i, r := range receipts {
		receiptsStorage[i] = (*types.ReceiptForStorage)(r)
	}
	// add block
	b.blocks = append(b.blocks, ibr.LlrIdxFullBlockRecord{
		LlrFullBlockRecord: ibr.LlrFullBlockRecord{
			Atropos:  prettyHash(bs.FinalizedStateRoot),
			Root:     bs.FinalizedStateRoot,
			Txs:      evmBlock.Transactions,
			Receipts: receiptsStorage,
			Time:     blockCtx.Time,
			GasUsed:  evmBlock.GasUsed,
		},
		Idx: blockCtx.Idx,
	})
	// add epoch
	b.currentEpoch = ier.LlrIdxFullEpochRecord{
		LlrFullEpochRecord: ier.LlrFullEpochRecord{
			BlockState: bs,
			EpochState: es,
		},
		Idx: es.Epoch,
	}
	b.epochs = append(b.epochs, b.currentEpoch)

	return b.tmpEvmStore.Commit(bs.LastBlock.Idx, bs.FinalizedStateRoot, true)
}

type memFile struct {
	*bytes.Buffer
}

func (f *memFile) Close() error {
	*f = memFile{}
	return nil
}

func (b *GenesisBuilder) Build(head genesis.Header) *genesisstore.Store {
	return genesisstore.NewStore(func(name string) (io.Reader, error) {
		buf := &memFile{bytes.NewBuffer(nil)}
		if name == genesisstore.BlocksSection {
			for i := len(b.blocks) - 1; i >= 0; i-- {
				_ = rlp.Encode(buf, b.blocks[i])
			}
			return buf, nil
		}
		if name == genesisstore.EpochsSection {
			for i := len(b.epochs) - 1; i >= 0; i-- {
				_ = rlp.Encode(buf, b.epochs[i])
			}
			return buf, nil
		}
		if name == genesisstore.EvmSection {
			it := b.tmpEvmStore.EvmDb.NewIterator(nil, nil)
			defer it.Release()
			_ = iodb.Write(buf, it)
		}
		if buf.Len() == 0 {
			return nil, errors.New("not found")
		}
		return buf, nil
	}, head, func() error {
		*b = GenesisBuilder{}
		return nil
	})
}

func (b *GenesisBuilder) DeployBaseContracts() {
	// deploy essential contracts
	// pre deploy NetworkInitializer
	b.SetCode(contracts.NetworkInitializerSmartContractAddress, contracts.NetworkInitializerBytecode)
	// pre deploy NodeDriver
	b.SetCode(contracts.NodeDriverSmartContractAddress, contracts.NodeDriverBytecode)
	// pre deploy NodeDriverAuth
	b.SetCode(contracts.NodeDriverAuthSmartContractAddress, contracts.NodeDriverAuthBytecode)
	// pre deploy Staking
	b.SetCode(contracts.StakingSmartContractAddress, contracts.StakingBytecode)
	b.SetCode(contracts.ValidatorInfoSmartContractAddress, contracts.ValidatorInfoBytecode)
	// set non-zero code for pre-compiled contracts
	b.SetCode(contracts.EvmWriterSmartContractAddress, []byte{0})
}

func (b *GenesisBuilder) InitializeEpoch(block idx.Block, epoch idx.Epoch, rules params.ProtocolRules, timestamp inter.Timestamp) {
	b.SetCurrentEpoch(ier.LlrIdxFullEpochRecord{
		LlrFullEpochRecord: ier.LlrFullEpochRecord{
			BlockState: iblockproc.BlockState{
				LastBlock: iblockproc.BlockCtx{
					Idx:     block - 1,
					Time:    timestamp,
					Atropos: hash.Event{},
				},
				FinalizedStateRoot:    hash.Hash{},
				EpochGas:              0,
				EpochCheaters:         lachesis.Cheaters{},
				CheatersWritten:       0,
				ValidatorStates:       make([]iblockproc.ValidatorBlockState, 0),
				NextValidatorProfiles: make(map[idx.ValidatorID]drivertype.Validator),
				DirtyRules:            nil,
				AdvanceEpochs:         0,
			},
			EpochState: iblockproc.EpochState{
				Epoch:             epoch - 1,
				EpochStart:        timestamp,
				PrevEpochStart:    timestamp - 1,
				EpochStateRoot:    hash.Zero,
				Validators:        pos.NewBuilder().Build(),
				ValidatorStates:   make([]iblockproc.ValidatorEpochState, 0),
				ValidatorProfiles: make(map[idx.ValidatorID]drivertype.Validator),
				Rules:             rules,
			},
		},
		Idx: epoch - 1,
	})
}

func (b *GenesisBuilder) GetGenesisTxs(sealedEpoch idx.Epoch, validators gpos.Validators, totalSupply *big.Int, delegations []driver.Delegation, driverOwner common.Address) types.Transactions {
	buildTx := txBuilder()
	internalTxs := make(types.Transactions, 0, 15)
	// initialization
	calldata := netinit.InitializeAll(sealedEpoch, totalSupply, driverOwner)
	internalTxs = append(internalTxs, buildTx(calldata, contracts.NetworkInitializerSmartContractAddress))
	// push genesis validators
	for _, v := range validators {
		calldata := driver.SetGenesisValidator(v)
		internalTxs = append(internalTxs, buildTx(calldata, contracts.NodeDriverSmartContractAddress))
	}
	// push genesis delegations
	for _, delegation := range delegations {
		calldata := driver.SetGenesisDelegation(delegation)
		internalTxs = append(internalTxs, buildTx(calldata, contracts.NodeDriverSmartContractAddress))
	}
	return internalTxs
}

func txBuilder() func(calldata []byte, addr common.Address) *types.Transaction {
	nonce := uint64(0)
	return func(calldata []byte, addr common.Address) *types.Transaction {
		tx := types.NewTransaction(nonce, addr, common.Big0, 1e10, common.Big0, calldata)
		nonce++
		return tx
	}
}
