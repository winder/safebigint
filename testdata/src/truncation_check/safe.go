package truncation_check

import "math/big"

func safeCases() {
	x := big.NewInt(123)
	_ = x.BitLen() // not a truncating method
	_ = x.Cmp(big.NewInt(1))
}
