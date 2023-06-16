package utils

import "math/big"

// ToSeth number of SETHN to Wei
func ToSeth(sethn uint64) *big.Int {
	return new(big.Int).Mul(new(big.Int).SetUint64(sethn), big.NewInt(1e18))
}
