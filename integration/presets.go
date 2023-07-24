package integration

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb/multidb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func Pbl1DBsConfig(scale func(uint64) uint64, fdlimit uint64) DBsConfig {
	return DBsConfig{
		Routing:      Pbl1RoutingConfig(),
		RuntimeCache: Pbl1RuntimeDBsCacheConfig(scale, fdlimit),
		GenesisCache: Pbl1GenesisDBsCacheConfig(scale, fdlimit),
	}
}

func Pbl1RoutingConfig() RoutingConfig {
	return RoutingConfig{
		Table: map[string]multidb.Route{
			"": {
				Type: "pebble-fsh",
			},
			"lachesis": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "C",
			},
			"gossip": {
				Type: "pebble-fsh",
				Name: "main",
			},
			"evm": {
				Type: "pebble-fsh",
				Name: "main",
			},
			"gossip/e": {
				Type: "pebble-fsh",
				Name: "events",
			},
			"evm/M": {
				Type: "pebble-drc",
				Name: "evm-data",
			},
			"evm-logs": {
				Type: "pebble-fsh",
				Name: "evm-logs",
			},
			"gossip-%d": {
				Type:  "leveldb-fsh",
				Name:  "epoch-%d",
				Table: "G",
			},
			"lachesis-%d": {
				Type:   "leveldb-fsh",
				Name:   "epoch-%d",
				Table:  "L",
				NoDrop: true,
			},
		},
	}
}

func Pbl1RuntimeDBsCacheConfig(scale func(uint64) uint64, fdlimit uint64) DBsCacheConfig {
	return DBsCacheConfig{
		Table: map[string]DBCacheConfig{
			"evm-data": {
				Cache:   scale(460 * opt.MiB),
				Fdlimit: fdlimit*460/1400 + 1,
			},
			"evm-logs": {
				Cache:   scale(260 * opt.MiB),
				Fdlimit: fdlimit*220/1400 + 1,
			},
			"main": {
				Cache:   scale(320 * opt.MiB),
				Fdlimit: fdlimit*280/1400 + 1,
			},
			"events": {
				Cache:   scale(240 * opt.MiB),
				Fdlimit: fdlimit*200/1400 + 1,
			},
			"epoch-%d": {
				Cache:   scale(100 * opt.MiB),
				Fdlimit: fdlimit*100/1400 + 1,
			},
			"": {
				Cache:   64 * opt.MiB,
				Fdlimit: fdlimit/100 + 1,
			},
		},
	}
}

func Pbl1GenesisDBsCacheConfig(scale func(uint64) uint64, fdlimit uint64) DBsCacheConfig {
	return DBsCacheConfig{
		Table: map[string]DBCacheConfig{
			"main": {
				Cache:   scale(1000 * opt.MiB),
				Fdlimit: fdlimit*1000/3000 + 1,
			},
			"evm-data": {
				Cache:   scale(1000 * opt.MiB),
				Fdlimit: fdlimit*1000/3000 + 1,
			},
			"evm-logs": {
				Cache:   scale(1000 * opt.MiB),
				Fdlimit: fdlimit*1000/3000 + 1,
			},
			"events": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3000 + 1,
			},
			"epoch-%d": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3000 + 1,
			},
			"": {
				Cache:   16 * opt.MiB,
				Fdlimit: fdlimit/100 + 1,
			},
		},
	}
}
