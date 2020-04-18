package zkstarks

import (
	"math/big"

	"github.com/actuallyachraf/algebra/ff"
	"github.com/actuallyachraf/algebra/poly"
	"github.com/actuallyachraf/go-merkle"
)

// FRI Layers construction
// We start with the evaluation domain generated during the domain parameters
// generation.
// We define a function called NextFRIDomain that given a domain returns
// half the elements of the input squared.
// NextFRIDomain(domain)
// return [x**2 for x in domain[:len(domain)/2]]
// The first FRI domain is a coset of a group of order 8192
// We take the first 4096 elements square them to find the next FRI domain.

// NextFRIDomain computes the next FRI domain.
func NextFRIDomain(evalDomain []ff.FieldElement) []ff.FieldElement {

	domainOrder := len(evalDomain)

	nextFRIDomainOrder := domainOrder / 2
	nextFRIDomain := make([]ff.FieldElement, nextFRIDomainOrder)

	for idx, elem := range evalDomain[:nextFRIDomainOrder] {
		nextFRIDomain[idx] = elem.Square()
	}

	return nextFRIDomain

}

// The FRI folding operator is a map on polynomial coefficients.
// For a given polynomial it folds it's coefficients by summing
// consecutive pairs of even/odd coefficients.
// First we extract a list of even and odd coefficients.
// We then sample a random field element trough the Fiat Shamir channel
// We multiply the odd coefficients by beta then sum them with the even
// coefficients (pair to pair) to produce a new a polynomial.

// evenCoeffs extracts a new polynomial using even index coefficients
func evenCoeffs(p poly.Polynomial) poly.Polynomial {

	var even []ff.FieldElement

	for i := 0; i < len(p); i += 2 {

		even = append(even, PrimeField.NewFieldElement(p[i]))

	}
	if len(even) == 0 {
		return poly.NewPolynomialInts(0)
	}
	return poly.NewPolynomial(even)
}

// oddCoeffs extracts a new polynomial using odd index coefficients
func oddCoeffs(p poly.Polynomial) poly.Polynomial {

	var odd []ff.FieldElement

	for i := 1; i < len(p); i += 2 {

		odd = append(odd, PrimeField.NewFieldElement(p[i]))

	}
	if len(odd) == 0 {
		return poly.NewPolynomialInts(0)
	}
	return poly.NewPolynomial(odd)
}

// NextFRIPolynomial creates the next FRI polynomial.
func NextFRIPolynomial(p poly.Polynomial, beta ff.FieldElement) poly.Polynomial {

	field := beta.Field()
	oddCoefficients := oddCoeffs(p)
	evenCoefficients := evenCoeffs(p)
	betaMonomial := poly.NewPolynomialBigInt(beta.Big())

	scaledCoeffs := oddCoefficients.Mul(betaMonomial, field.Modulus())

	nextFRIPoly := scaledCoeffs.Add(evenCoefficients, field.Modulus())

	return nextFRIPoly

}

// NextFRILayer constructs the next FRI layer
// a Layer is a tuple consisting of an evaluation domain and polynomial
// to create the next fri layer we evaluate the FRI-polynomial over the FRI-domain
func NextFRILayer(domain []ff.FieldElement, p poly.Polynomial, beta ff.FieldElement) ([]ff.FieldElement, poly.Polynomial, []ff.FieldElement) {

	domainEval := func(p poly.Polynomial, domain []ff.FieldElement) []ff.FieldElement {

		field := PrimeField
		evals := make([]ff.FieldElement, len(domain))

		for idx, elem := range domain {
			eval := p.Eval(elem.Big(), elem.Field().Modulus())
			evals[idx] = field.NewFieldElement(eval)
		}
		return evals
	}
	nextFRIDomain := NextFRIDomain(domain)
	nextFRIPoly := NextFRIPolynomial(p, beta)
	nextLayer := domainEval(nextFRIPoly, nextFRIDomain)

	return nextFRIDomain, nextFRIPoly, nextLayer
}

// DomainHash returns a merkle root of the domain elements
func DomainHash(domain []ff.FieldElement) []byte {

	domainBytes := make([][]byte, len(domain))

	for idx, elem := range domain {
		domainBytes[idx] = elem.Big().Bytes()
	}

	return merkle.Root(domainBytes)
}

// DomainBytes returns a byte serialized domain element set
func DomainBytes(domain []ff.FieldElement) [][]byte {

	domainBytes := make([][]byte, len(domain))

	for idx, elem := range domain {
		domainBytes[idx] = elem.Big().Bytes()
	}
	return domainBytes
}

// cosetDomainBytes returns a byte serialized domain element set
func cosetDomainBytes(domain []*big.Int) [][]byte {

	domainBytes := make([][]byte, len(domain))

	for idx, elem := range domain {
		domainBytes[idx] = elem.Bytes()
	}
	return domainBytes
}

// serializeAuditPath serializes a merkle audithash
func serializeAuditPath(ap []merkle.AuditHash) []byte {
	var auditPath = make([]byte, 0)

	for _, path := range ap {

		var b = make([]byte, 0, 33)
		copy(b, path.Val)
		if path.RightOperator {
			b = append(b, 1)
		} else {
			b = append(b, 0)
		}
		auditPath = append(auditPath, b...)
	}
	return auditPath

}

// GenerateFRICommitment given the composition polynomial
// the evaluation domain, the evaluations on said domain and
// the first commitment root.
func GenerateFRICommitment(compositionPoly poly.Polynomial, domain []ff.FieldElement, compositionEvals []ff.FieldElement, compositionRoot []byte, fs Channel) ([][]ff.FieldElement, []poly.Polynomial, [][]ff.FieldElement, [][]byte) {

	FRIPolynomials := []poly.Polynomial{compositionPoly}
	FRIDomains := [][]ff.FieldElement{domain}
	FRILayers := [][]ff.FieldElement{compositionEvals}
	FRIMerkleRoots := [][]byte{compositionRoot}

	iter := FRIPolynomials[len(FRIPolynomials)-1]
	field := PrimeField

	for iter.Degree() > 0 {

		beta := field.NewFieldElement(fs.RandFE(PrimeField.Modulus()))

		nextFRIDomain, nextFRIPoly, nextFRILayer := NextFRILayer(FRIDomains[len(FRIDomains)-1], FRIPolynomials[len(FRIPolynomials)-1], beta)

		root := DomainHash(nextFRILayer)

		FRIDomains = append(FRIDomains, nextFRIDomain)
		FRIPolynomials = append(FRIPolynomials, nextFRIPoly)
		FRILayers = append(FRILayers, nextFRILayer)
		FRIMerkleRoots = append(FRIMerkleRoots, root)

		fs.Send(FRIMerkleRoots[len(FRIMerkleRoots)-1])

		iter = FRIPolynomials[len(FRIPolynomials)-1]

	}
	fs.Send(FRIPolynomials[len(FRIPolynomials)-1][0].Bytes())

	return FRIDomains, FRIPolynomials, FRILayers, FRIMerkleRoots
}

// In order to verify the commitment proofs we need to implement to new functions
// the first will will send the FS channel data to verify that each FRI layer
// is consistent with the others ,the second will send the data required to
// decommit on the trace polynomial.
// This part deals mainly with non-interactiveness of our proof system.

// DecommitFRILayers iterates over the fri-layers and fri-roots (except the last ones)
// as those are constant and sends
// the following data trough the FS channel :
// - Element of the FRI layer at the given index
// - It's merkle proof
// - Sibling Element on the fri-layer if the element is cp_i(x) it's sibling
// is cp_i(-x)
// - The merkle proof of the sibling.
func DecommitFRILayers(index int, channel *Channel, friLayers [][]ff.FieldElement) {

	for i := 0; i < len(friLayers)-1; i++ {
		layer := friLayers[i]
		length := len(layer)
		index = index % length
		siblingIndex := (index + (length / 2)) % length

		elemBytes := layer[index].Big().Bytes()
		elemProof, err := merkle.Proof(DomainBytes(layer), index)

		if err != nil {
			panic(err)
		}
		siblingBytes := layer[siblingIndex].Big().Bytes()
		siblingProof, err := merkle.Proof(DomainBytes(layer), siblingIndex)
		if err != nil {
			panic(err)
		}
		elemProofBytes := serializeAuditPath(elemProof)
		siblingProofBytes := serializeAuditPath(siblingProof)

		channel.Send(elemBytes)
		channel.Send(elemProofBytes)
		channel.Send(siblingBytes)
		channel.Send(siblingProofBytes)

	}
	// Send the last layer element
	channel.Send(friLayers[len(friLayers)-1][0].Big().Bytes())
}

// Decommiting on the trace polynomial involves verifying the evaluation
// of the composition polynomial
// The value f(x) with its authentication path.
// The value f(gx) with its authentication path.
// The value f(g^2x) with its authentication path.
// The verifier, knowing the random coefficients of the composition polynomial,
// can compute its evaluation at x, and compare it with the first element sent from the first FRI layer.

// DecommitOnQuery takes an index, a channel, coset evaluations and sends
// the evaluations and their proofs at the given index
func DecommitOnQuery(index int, channel *Channel, cosetEval []*big.Int, friLayers [][]ff.FieldElement) {

	if index+16 > len(cosetEval) {
		panic("coset eval index out of range")
	}

	cosetBytes := cosetDomainBytes(cosetEval)

	firstEvalBytes := cosetBytes[index]
	firstEvalAP, err := merkle.Proof(cosetBytes, index)
	if err != nil {
		panic(err)
	}
	firstEvalAPSerialized := serializeAuditPath(firstEvalAP)

	channel.Send(firstEvalBytes)
	channel.Send(firstEvalAPSerialized)

	secondEvalBytes := cosetBytes[index+8]
	secondEvalAP, err := merkle.Proof(cosetBytes, index+8)
	if err != nil {
		panic(err)
	}
	secondEvalAPSerialized := serializeAuditPath(secondEvalAP)

	channel.Send(secondEvalBytes)
	channel.Send(secondEvalAPSerialized)

	thirdEvalBytes := cosetBytes[index+16]
	thirdEvalAP, err := merkle.Proof(cosetBytes, index+16)
	if err != nil {
		panic(err)
	}
	thirdEvalAPSerialized := serializeAuditPath(thirdEvalAP)

	channel.Send(thirdEvalBytes)
	channel.Send(thirdEvalAPSerialized)

	DecommitFRILayers(index, channel, friLayers)
}

// FRIDecommit receives random values from the verifier (using FS)
// and decommits on each query index.
func FRIDecommit(channel *Channel, cosetEval []*big.Int, friLayers [][]ff.FieldElement) {

	lb := big.NewInt(0)
	ub := big.NewInt(8196 - 16)

	for i := 0; i < 3; i++ {
		randIdx := channel.RandInt(lb, ub)

		DecommitOnQuery(int(randIdx.Int64()), channel, cosetEval, friLayers)
	}
}
