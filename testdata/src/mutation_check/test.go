package mutationcheck

import (
	"math/big"
)

type customBig struct{}

func (c *customBig) Add(x, y *customBig) {} // shadowing method

func sharedMutationCases() {
	// Unsafe: receiver is same as first argument
	a := big.NewInt(10)
	b := big.NewInt(2)
	a.Add(a, b) // want "shared-object mutation: calling Add with receiver also passed as argument"

	// Unsafe: receiver is reused in both argument slots
	c := big.NewInt(5)
	c.Mul(c, c) // want "shared-object mutation: calling Mul with receiver also passed as argument"

	// Unsafe: alias of receiver
	x := big.NewInt(100)
	y := x
	x.Sub(x, y) // want "shared-object mutation: calling Sub with receiver also passed as argument"

	// Unsafe: even with three args
	e := big.NewInt(3)
	e.Exp(e, e, nil) // want "shared-object mutation: calling Exp with receiver also passed as argument"

	// Safe: all different
	f := big.NewInt(2)
	g := big.NewInt(3)
	h := big.NewInt(4)
	f.Add(g, h) // OK

	// Safe: non-mutating method
	result := new(big.Int)
	xor := new(big.Int)
	yor := new(big.Int)
	result.Xor(xor, yor) // OK

	// Safe: unrelated type with similar method name
	var custom customBig
	custom.Add(&custom, &custom) // OK

	// Unsafe: reassigned variable
	m := big.NewInt(10)
	n := m
	n.And(n, m) // want "shared-object mutation: calling And with receiver also passed as argument"

	// Unsafe: all arguments are the same variable
	q := big.NewInt(7)
	q.Mod(q, q) // want "shared-object mutation: calling Mod with receiver also passed as argument"

	// Safe: selector field on different variable
	type S struct{ x *big.Int }
	var s S
	s.x = big.NewInt(20)
	other := big.NewInt(5)
	s.x.Add(other, other) // OK
}
