package mutation_check

import "math/big"

func safeMutation() {
	x := big.NewInt(1)
	y := big.NewInt(2)
	z := big.NewInt(3)

	// Valid: receiver is different from args
	x.Add(y, z)

	// Unrelated method
	x.BitLen()

	// Struct selector field (should not match big.Int directly)
	type wrapper struct{ b *big.Int }
	var w wrapper
	w.b = big.NewInt(5)
	w.b.Add(y, z)
}

func mixedArgs() {
	x := big.NewInt(1)
	x.Add(nil, nil) // nil arg — getReferencedObject will return nil
}

func notAMutation() {
	x := big.NewInt(1)
	_ = x.BitLen() // triggers checkForMutation, skips early
}
