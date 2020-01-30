package main

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/actuallyachraf/algebra/ff"
	"github.com/actuallyachraf/algebra/poly"
	"github.com/actuallyachraf/go-merkle"
	"github.com/actuallyachraf/zkstarks"
)

func print(elems []ff.FieldElement) {
	for i, v := range elems {
		fmt.Printf("Fib : (%d) = %s\n", i, v.String())
	}
}

func main() {

	fib := zkstarks.GenSeq()
	print(fib)
	// generator = 5**(3*2**20)
	generator := zkstarks.PrimeFieldGen.Exp(new(big.Int).SetInt64(3145728))
	subgroup := zkstarks.GenElems(generator, 1024)
	fmt.Println("Subgroup G of order  :", len(subgroup))
	fmt.Println("Generator to power 1024", generator.Exp(new(big.Int).SetInt64(1024)).Big().String())
	// We are going to interpolate the lagrange polynomial
	// where the x_s are the subgroup elements and their evaluations are the
	// fib sequence terms.
	// Interpolation is a heavy computational task that takes the longest time
	// when building the stark .

	points := generatePoints(subgroup[:len(subgroup)-1], fib)
	fmt.Println("Subgroup order :", len(subgroup))
	fmt.Println("Eval points : ", len(points))

	polynomial := poly.Lagrange(points, zkstarks.PrimeField.Modulus())

	v := polynomial.Eval(new(big.Int).SetInt64(2), zkstarks.PrimeField.Modulus())

	if v.Cmp(new(big.Int).SetInt64(1302089273)) != 0 {
		fmt.Println("Expected P(2) = 1302089273 but Got P(2) : ", v)
		panic("Error : polynomial evaluation failed")
	}

	// The computational trace that is used to build the proof is considered here
	// as polynomial evaluations on G where |G| = 1024.
	// The trace can be extended to a larger group by evaluating said polynomial
	// over a group |H| = 8192 creating a Reed-Solomon EC.
	// To generate such group we take a generator of an order 8192 group
	// and shift it by the G's generator to obtain a coset that'll represent the
	// evaluation domain.
	// 5 ** ((2 ** 30 * 3) // 8192)
	hGenerator := zkstarks.PrimeFieldGen.Exp(big.NewInt(393216))
	fmt.Println(hGenerator.String())
	H := make([]ff.FieldElement, 8192)
	var i int64
	for i = 0; i < 8192; i++ {
		H[i] = hGenerator.Exp(big.NewInt(i))
	}
	evalDomain := make([]ff.FieldElement, 8192)
	for i = 0; i < 8192; i++ {
		evalDomain[i] = zkstarks.PrimeField.Mul(zkstarks.PrimeFieldGen, H[i])
	}

	h := zkstarks.PrimeFieldGen
	hInv := h.Inv()
	// Sanity checks
	for i = 0; i < 8192; i++ {
		if !zkstarks.PrimeField.Mul(zkstarks.PrimeField.Mul(hInv, evalDomain[1]).Exp(big.NewInt(i)), h).Equal(evalDomain[i]) {
			panic("error eval domain is incorrect")
		}
	}
	// the interpoled polynomial over the subgroup is evaluated
	// over the coset domain
	cosetEval := make([]*big.Int, len(evalDomain))
	cosetEvalBytes := make([][]byte, len(evalDomain))
	for i, v := range evalDomain {
		cosetEval[i] = polynomial.Eval(v.Big(), zkstarks.PrimeField.Modulus())
		cosetEvalBytes[i] = cosetEval[i].Bytes()
	}

	// Commitments are cryptographic protocol used to commit to certain
	// values, hash functions are the most elementary of such protocols.
	// When committing to a range of values a more efficient way to do
	// so is to use merkle trees.
	commitmentRoot := merkle.Root(cosetEvalBytes)
	fmt.Println("Commitment to Coset Evaluation :", hex.EncodeToString(commitmentRoot))
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
