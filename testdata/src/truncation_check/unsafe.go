package truncation_check

import "math/big"

func unsafeCases() {
	x := big.NewInt(1)
	_ = x.Uint64() // want "calling Uint64 on \\*big.Int may silently truncate or overflow"
	_ = x.Int64()  // want "calling Int64 on \\*big.Int may silently truncate or overflow"
}
