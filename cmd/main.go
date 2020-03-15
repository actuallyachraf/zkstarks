package main

func main() {

}

/*
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
func GenerateDomainParameters() ([]ff.FieldElement, ff.FieldElement, []ff.FieldElement, ff.FieldElement, []ff.FieldElement, []ff.FieldElement, poly.Polynomial, []*big.Int, []byte, *zkstarks.Channel) {
	a := zkstarks.GenSeq()
	g := zkstarks.PrimeFieldGen.Exp(new(big.Int).SetInt64(3145728))
	G := zkstarks.GenElems(g, 1024)
	points := generatePoints(G[:len(G)-1], a)
	f := poly.Lagrange(points, zkstarks.PrimeField.Modulus())
	hGenerator := zkstarks.PrimeFieldGen.Exp(big.NewInt(393216))
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
		cosetEval[i] = f.Eval(v.Big(), zkstarks.PrimeField.Modulus())
		cosetEvalBytes[i] = cosetEval[i].Bytes()
	}

	// Commitments are cryptographic protocol used to commit to certain
	// values, hash functions are the most elementary of such protocols.
	// When committing to a range of values a more efficient way to do
	// so is to use merkle trees.
	commitmentRoot := merkle.Root(cosetEvalBytes)
	fmt.Println("Commitment to Coset Evaluation :", hex.EncodeToString(commitmentRoot))

	fsChan := zkstarks.NewChannel()
	fsChan.Send(commitmentRoot)

	return a, g, G, hGenerator, H, evalDomain, f, cosetEval, commitmentRoot, fsChan

}

// GenerateProgramConstraints generates the polynomial constraints for the proof.
func GenerateProgramConstraints(f poly.Polynomial) {

	// Each constraint (see /constraint.go) is represented by a polynomial u(x)
	// that evaluates to 0 for a certain group element x in G
	// When a polynomial evaluates to zero for a group element
	// it means that u(x) is divisble by Prod(0,k) of (x - g_i) where g_i
	// are the group elements.
	// Therefore each constraint can be encoded to a rational function
	// i.e a quotient of polynomials if the constraint are correct
	// the quotient is itself a polynomial (quotient can be irreducible).
	// A constraint is valid becomes simply a check that u(x)/r(x) is
	// a polynomial.

	// The first constraint :
	// f(x) - 1 = 0 <=> f(x)-1/(X-1)
	num0 := f.Sub(poly.NewPolynomialInts(1))
	dem0 := poly.NewPolynomialInts(1, 1)
	quoPolyConstraint1, remPolyConstraint1 := num0.Div(dem0)

	// Validate the first constraint
	if num0.Eval(nt.FromInt64(1), nt.FromInt64(0)).Cmp(nt.Zero) != 0 {

	}

}
*/
