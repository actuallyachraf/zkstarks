package main

import (
	"fmt"
	"math/big"

	"github.com/actuallyachraf/algebra/ff"
	"github.com/actuallyachraf/algebra/poly"
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

	fmt.Println("Generator to power 1024", generator.Exp(new(big.Int).SetInt64(1024)).Big().String())
	// We are going to interpolate the lagrange polynomial
	// where the x_s are the subgroup elements and their evaluations are the
	// fib sequence terms.
	points := generatePoints(subgroup[:len(subgroup)-1], fib)
	fmt.Println("Subgroup order :", len(subgroup))
	fmt.Println("Eval points : ", len(points))

	polynomial := poly.Lagrange(points, zkstarks.PrimeField.Modulus())

	v := polynomial.Eval(new(big.Int).SetInt64(2), zkstarks.PrimeField.Modulus())

	if v.Cmp(new(big.Int).SetInt64(1302089273)) != 0 {
		fmt.Println("Expected P(2) = 1302089273 but Got P(2) : ", v)
		panic("Error : polynomial evaluation failed")
	}

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
