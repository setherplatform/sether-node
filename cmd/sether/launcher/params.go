package launcher

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/setherplatform/sether-node/params"

	"github.com/setherplatform/sether-node/genesis"
	"github.com/setherplatform/sether-node/genesis/genesisstore"
)

var (
	Bootnodes = map[string][]string{
		"main": {},
		"test": {
			"enode://4dbc94a60d0d5c91b0fcafd8dd931bb77a2de8b269c80a58da676af3a74fcf9fa5457c536aea40544080780a99b0dcf6629f34f0974d21da7a4c2f62a0074eec@167.235.203.218:6534",
			"enode://eae25d66e8fedd6e32801f627ae8ff35b74478460c3ea5d773bcf1417dc88e75b8284eb8ee9aa2c5d450f785591b63b440ea708a590d355ed57e1f51c7fc3082@135.181.90.176:6534",
			"enode://8cb1c9e3d93a88a9abb56aaa9c4fdd85bf6c90d6145236c696a6fc334890d334e462c406c6e72b2b463ef41844fc065f769fbc628a6371b3d0c70e3636d272bc@5.78.68.153:6534",
		},
	}

	mainnetHeader = genesis.Header{
		GenesisID:   hash.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		NetworkID:   params.MainNetworkID,
		NetworkName: "main",
	}

	testnetHeader = genesis.Header{
		GenesisID:   hash.HexToHash("0xb93bd2bdd0c13653fd0b636bdc865f14b2fd0f875340fe92ed4580ad774ed60d"),
		NetworkID:   params.TestNetworkID,
		NetworkName: "test",
	}

	AllowedGenesis = []GenesisTemplate{
		{
			Name:   "Mainnet",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection: hash.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
				genesisstore.BlocksSection: hash.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
				genesisstore.EvmSection:    hash.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
			},
		},

		{
			Name:   "Testnet",
			Header: testnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection: hash.HexToHash("0x2a4d5c5cdd0321be62ca25a37d358355773e2549baa4efd189c01d62ffd615ca"),
				genesisstore.BlocksSection: hash.HexToHash("0x3f490c6a2cc3c1916cae1e706b956e42b1b2e138a18655832a2fb78cdeb874ba"),
				genesisstore.EvmSection:    hash.HexToHash("0xb01838e5f9f0ab0a9e8940f68a6f71f001b972a7cf23c4aecd17278b518d617f"),
			},
		},
	}
)

func overrideParams() {
	ethparams.MainnetBootnodes = []string{}
	ethparams.RopstenBootnodes = []string{}
	ethparams.RinkebyBootnodes = []string{}
	ethparams.GoerliBootnodes = []string{}
}
