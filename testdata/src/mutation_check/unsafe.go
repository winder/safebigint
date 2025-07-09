package mutation_check

import "math/big"

func unsafeMutation() {
	x := big.NewInt(1)
	x.Add(x, big.NewInt(2)) // want "shared-object mutation: calling Add with receiver also passed as argument"
	x.Mul(x, x)             // want "shared-object mutation: calling Mul with receiver also passed as argument"

	y := x
	x.Sub(x, y) // want "shared-object mutation: calling Sub with receiver also passed as argument"

	z := big.NewInt(5)
	z.Exp(z, z, nil) // want "shared-object mutation: calling Exp with receiver also passed as argument"
}
