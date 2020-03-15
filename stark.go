package zkstarks

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/actuallyachraf/algebra/ff"
	"github.com/actuallyachraf/algebra/nt"
	"github.com/actuallyachraf/algebra/poly"
	"github.com/actuallyachraf/go-merkle"
)

// PrimeField in the remaining of the implementation we use a prime field
// with modulus q = 3221225473
var PrimeField, _ = ff.NewFiniteField(new(nt.Integer).SetUint64(3221225473))

// PrimeFieldGen is a generator of said field
var PrimeFieldGen = PrimeField.NewFieldElementFromInt64(5)

// Our goal is to construct a proof about the 1023rd element in the fibonacci
// sequence a_{n+2} = a_{n+1}^2 + a_{n}^2.
// The sequence starts with [1,3141592]

// DomainParameters represents the domain parameters of the proof generation
type DomainParameters struct {
	Trace                 []ff.FieldElement
	GeneratorG            ff.FieldElement
	SubgroupG             []ff.FieldElement
	GeneratorH            ff.FieldElement
	SubgroupH             []ff.FieldElement
	EvaluationDomain      []ff.FieldElement
	Polynomial            poly.Polynomial
	PolynomialEvaluations []*big.Int
	EvaluationRoot        []byte
}

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

// GenerateDomainParameters reproduces the domain parameters required
// for proof generation :
// a : the trace of FibSeq(1,3141592)
// g : generator of the subgroup of order 1024
// G : the subgroup elements
// h : generator of the larger evaluation domain of order 8192
// H : the subgroup elements
// f : interpolated polynomial over G
// fEval : evaluation of f over the elements of H
// fEvalCommitmentRoot : merkle commitment of the evaluations of over H
// fsChan : fiat shamir channel initiated with the commitment root
func GenerateDomainParameters() ([]ff.FieldElement, ff.FieldElement, []ff.FieldElement, ff.FieldElement, []ff.FieldElement, []ff.FieldElement, poly.Polynomial, []*big.Int, []byte, *Channel) {
	a := GenSeq()
	g := PrimeFieldGen.Exp(new(big.Int).SetInt64(3145728))
	G := GenElems(g, 1024)
	points := generatePoints(G[:len(G)-1], a)
	f := poly.Lagrange(points, PrimeField.Modulus())
	hGenerator := PrimeFieldGen.Exp(big.NewInt(393216))
	H := make([]ff.FieldElement, 8192)
	var i int64
	for i = 0; i < 8192; i++ {
		H[i] = hGenerator.Exp(big.NewInt(i))
	}
	evalDomain := make([]ff.FieldElement, 8192)
	for i = 0; i < 8192; i++ {
		evalDomain[i] = PrimeField.Mul(PrimeFieldGen, H[i])
	}
	h := PrimeFieldGen
	hInv := h.Inv()
	// Sanity checks
	for i = 0; i < 8192; i++ {
		if !PrimeField.Mul(PrimeField.Mul(hInv, evalDomain[1]).Exp(big.NewInt(i)), h).Equal(evalDomain[i]) {
			panic("error eval domain is incorrect")
		}
	}
	// the interpoled polynomial over the subgroup is evaluated
	// over the coset domain
	cosetEval := make([]*big.Int, len(evalDomain))
	cosetEvalBytes := make([][]byte, len(evalDomain))
	for i, v := range evalDomain {
		cosetEval[i] = f.Eval(v.Big(), PrimeField.Modulus())
		cosetEvalBytes[i] = cosetEval[i].Bytes()
	}

	// Commitments are cryptographic protocol used to commit to certain
	// values, hash functions are the most elementary of such protocols.
	// When committing to a range of values a more efficient way to do
	// so is to use merkle trees.
	commitmentRoot := merkle.Root(cosetEvalBytes)
	fmt.Println("Commitment to Coset Evaluation :", hex.EncodeToString(commitmentRoot))

	fsChan := NewChannel()
	fsChan.Send(commitmentRoot)

	return a, g, G, hGenerator, H, evalDomain, f, cosetEval, commitmentRoot, fsChan

}
func generatePoints(x []ff.FieldElement, y []ff.FieldElement) []poly.Point {

	if len(x) != len(y) {
		panic("Error : lists must be of the same length")
	}
	interpolationPoints := make([]poly.Point, len(x))

	for i := 0; i < len(x); i++ {
		interpolationPoints[i] = poly.NewPoint(x[i].Big(), y[i].Big())
	}

	return interpolationPoints
}
