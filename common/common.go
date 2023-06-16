package common

import (
	"github.com/setherplatform/sether-node/secp256k1"
	"math/big"
)

type Key string

type Point struct {
	X big.Int
	Y big.Int
}

type HexPoint struct {
	X string
	Y string
}

func (p HexPoint) ToPoint() Point {
	return Point{
		X: *secp256k1.HexToBigInt(p.X),
		Y: *secp256k1.HexToBigInt(p.Y),
	}
}

func (p Point) ToHex() HexPoint {
	return HexPoint{
		X: p.X.Text(16),
		Y: p.Y.Text(16),
	}
}
