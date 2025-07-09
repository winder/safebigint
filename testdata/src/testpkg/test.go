package testpkg

import (
	"math/big"
)

type myBigInt struct{}

func (m *myBigInt) Uint64() uint64 { return 42 } // Should NOT trigger

func safeConversion(b *big.Int) uint64 {
	if b.Cmp(big.NewInt(0)) < 0 {
		return 0
	}
	return b.Uint64() // want "calling Uint64\\(\\) on \\*big.Int may truncate large values silently"
}

func testMixed() {
	x := big.NewInt(123)
	_ = x.Uint64() // want "calling Uint64\\(\\) on \\*big.Int may truncate large values silently"

	y := new(myBigInt)
	_ = y.Uint64() // OK: user-defined type

	z := big.NewInt(999)
	_ = z.Int64() // OK: not Uint64
}

func testIgnore() {
	var i int
	_ = uint64(i) // OK: not big.Int
}
