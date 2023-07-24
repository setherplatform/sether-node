package integration

import (
	"github.com/setherplatform/sether-node/genesis/makefakegenesis"
	"io/ioutil"
	"os"
	"testing"

	"github.com/Fantom-foundation/lachesis-base/abft"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/utils/cachescale"
	"github.com/ethereum/go-ethereum/common"

	"github.com/setherplatform/sether-node/gossip"
	"github.com/setherplatform/sether-node/inter"
	"github.com/setherplatform/sether-node/utils"
	"github.com/setherplatform/sether-node/vecmt"
)

func BenchmarkFlushDBs(b *testing.B) {
	dir := tmpDir("flush_bench")
	defer os.RemoveAll(dir)
	genStore := makefakegenesis.FakeGenesisStore(1, utils.ToSeth(1), utils.ToSeth(1))
	g := genStore.Genesis()
	_, _, store, s2, _, closeDBs := MakeEngine(dir, &g, Configs{
		Sether:        gossip.DefaultConfig(cachescale.Identity),
		SetherStore:   gossip.DefaultStoreConfig(cachescale.Identity),
		Lachesis:      abft.DefaultConfig(),
		LachesisStore: abft.DefaultStoreConfig(cachescale.Identity),
		VectorClock:   vecmt.DefaultConfig(cachescale.Identity),
		DBs:           Pbl1DBsConfig(cachescale.Identity.U64, 512),
	})
	defer closeDBs()
	defer store.Close()
	defer s2.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		n := idx.Block(0)
		randUint32s := func() []uint32 {
			arr := make([]uint32, 128)
			for i := 0; i < len(arr); i++ {
				arr[i] = uint32(i) ^ (uint32(n) << 16) ^ 0xd0ad884e
			}
			return []uint32{uint32(n), uint32(n) + 1, uint32(n) + 2}
		}
		for !store.IsCommitNeeded() {
			store.SetBlock(n, &inter.Block{
				Time:        inter.Timestamp(n << 32),
				Atropos:     hash.Event{},
				Events:      hash.Events{},
				Txs:         []common.Hash{},
				InternalTxs: []common.Hash{},
				SkippedTxs:  randUint32s(),
				GasUsed:     uint64(n) << 24,
				Root:        hash.Hash{},
			})
			n++
		}
		b.StartTimer()
		err := store.Commit()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func tmpDir(name string) string {
	dir, err := ioutil.TempDir("", name)
	if err != nil {
		panic(err)
	}
	return dir
}
