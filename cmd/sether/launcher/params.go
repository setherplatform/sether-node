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
			"enode://82e8d25cb18ab2349cd6cd55e9eb3b201ddb3d107b129adde535fbf476d0750460c412e15cca6cb58230a7363d6c913e73fef0da896c3144eec6b50d946d2cf1@95.216.219.209:22220",
			"enode://8098c7e1bdf47971abe83ef311be55e1753f09744d48965d6534612d9d5bf6eafe62cc0ca9189985236eb8dc21cb3a55aad1dcb0b94019b0aa4d16f7f12a9765@91.107.204.78:22220",
			"enode://384432cb9cc3bd7e7758336b69efbfb21a4d3526695bffb6f0139e51322e86b254a7979c2126757f8cf88f2a9369926d426104d2370215a2d4e6fcffa743388d@195.201.118.191:22220",
		},
	}

	mainnetHeader = genesis.Header{
		GenesisID:   hash.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		NetworkID:   params.MainNetworkID,
		NetworkName: "main",
	}

	testnetHeader = genesis.Header{
		GenesisID:   hash.HexToHash("0xf17d24f77e7c45b9be3516b5607a52891711eac46a8bfb11d5104b4d9b387f5d"),
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
