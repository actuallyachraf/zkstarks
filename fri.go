package zkstarks

import (
	"github.com/actuallyachraf/algebra/ff"
	"github.com/actuallyachraf/algebra/nt"
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

// NextFRILayer constructs the next FRI layer
// a Layer is a tuple consisting of an evaluation domain and polynomial
// to create the next fri layer we evaluate the FRI-polynomial over the FRI-domain
func NextFRILayer(domain []ff.FieldElement, p poly.Polynomial, beta ff.FieldElement) ([]ff.FieldElement, poly.Polynomial, []ff.FieldElement) {

	domainEval := func(p poly.Polynomial, domain []ff.FieldElement) []ff.FieldElement {

		field := domain[0].Field()
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

// NextFRIPolynomial creates the next FRI polynomial.
func NextFRIPolynomial(p poly.Polynomial, beta ff.FieldElement) poly.Polynomial {

	field := beta.Field()

	evenCoeffs := func(p poly.Polynomial) poly.Polynomial {

		var even []ff.FieldElement
		var two = nt.FromInt64(2)
		var zero = nt.FromInt64(0)

		for _, coeff := range p {

			if nt.Mod(coeff, two).Cmp(zero) == 0 {
				even = append(even, field.NewFieldElement(coeff))
			}
		}

		return poly.NewPolynomial(even)
	}
	oddCoeffs := func(p poly.Polynomial) poly.Polynomial {

		var odd []ff.FieldElement
		var two = nt.FromInt64(2)
		var zero = nt.FromInt64(0)

		for _, coeff := range p {

			if nt.Mod(coeff, two).Cmp(zero) != 0 {
				odd = append(odd, field.NewFieldElement(coeff))
			}
		}

		return poly.NewPolynomial(odd)
	}

	oddCoefficients := oddCoeffs(p)
	evenCoefficients := evenCoeffs(p)
	betaMonomial := poly.NewPolynomialBigInt(beta.Big())

	scaledCoeffs := oddCoefficients.Mul(betaMonomial, field.Modulus())

	nextFRIPoly := scaledCoeffs.Add(evenCoefficients, field.Modulus())

	return nextFRIPoly

}

// DomainHash returns a merkle root of the domain elements
func DomainHash(domain []ff.FieldElement) []byte {

	domainBytes := make([][]byte, len(domain))

	for idx, elem := range domain {
		domainBytes[idx] = elem.Big().Bytes()
	}

	return merkle.Root(domainBytes)
}
