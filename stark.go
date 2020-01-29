package zkstarks

import (
	"math/big"

	"github.com/actuallyachraf/algebra/ff"
	"github.com/actuallyachraf/algebra/nt"
)

// PrimeField in the remaining of the implementation we use a prime field
// with modulus q = 3221225473
var PrimeField, _ = ff.NewFiniteField(new(nt.Integer).SetUint64(3221225473))

// PrimeFieldGen is a generator of said field
var PrimeFieldGen = PrimeField.NewFieldElementFromInt64(5)

// Our goal is to construct a proof about the 1023rd element in the fibonacci
// sequence a_{n+2} = a_{n+1}^2 + a_{n}^2.
// The sequence starts with [1,3141592]

// GenSeq computes the actual sequence
func GenSeq() []ff.FieldElement {
	// FibSeq defines our fibonnaci sequence
	var FibSeq = make([]ff.FieldElement, 1023)

	fib := func(x, y ff.FieldElement) ff.FieldElement {
		return PrimeField.Add(x.Square(), y.Square())
	}

	FibSeq[0] = PrimeField.NewFieldElementFromInt64(1)
	FibSeq[1] = PrimeField.NewFieldElementFromInt64(3141592)

	for i := 2; i < len(FibSeq); i++ {
		FibSeq[i] = fib(FibSeq[i-1], FibSeq[i-2])
	}

	return FibSeq
}

// The unisolvence theorem states that given n+1 pairs of points (x_i,y_i) there
// exists a polynomial Q of degree at most n such as Q(x_i) = y_i
// Our Fibonacci Sequence of size 1023 can be represented as evaluations
// of a polynomial of degree 1022
// The primefield we use under multiplication (remove 0 and addition)
// is a cyclic group of order 3*2^20 so it contains subgroups of size 3*2^i
// for 0 <= i <= 30.
// We want to restrict our calculations to the subgroup of size 1024.
// To create the group in question we capture a generator of it and compute
// it's elements.

// GenElems returns the list of field elements of the subgroup G of order 1024
func GenElems(generator ff.FieldElement, order int) []ff.FieldElement {

	var subgroup = make([]ff.FieldElement, order)

	var i int64

	for i = 0; i < int64(order); i++ {
		subgroup[i] = generator.Exp(new(big.Int).SetInt64(i))
	}

	return subgroup
}
