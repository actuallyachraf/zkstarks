package zkstarks

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/actuallyachraf/algebra/ff"
	"github.com/actuallyachraf/algebra/nt"
	"github.com/actuallyachraf/algebra/poly"
)

func TestZKGen(t *testing.T) {

	a, g, G, h, H, evalDomain, f, fEvals, fCommitment, fsChannel := GenerateDomainParameters()

	domainParams := &DomainParameters{
		a,
		g,
		G,
		h,
		H,
		evalDomain,
		f,
		fEvals,
		fCommitment,
	}

	domainParamsJSON, err := json.MarshalIndent(domainParams, "", " ")
	if err != nil {
		t.Fatal("failed to serialize domain params to JSON")
	}
	err = ioutil.WriteFile("./domainparams.json", domainParamsJSON, 0711)
	if err != nil {
		t.Fatal("failed to serialize domain params to JSON")
	}

	t.Run("TestParamGen", func(t *testing.T) {
		t.Log("Trace length :", len(a))
		t.Log("Subgroup G generator :", g)
		t.Log("Subgroup order:", len(G))
		t.Log("Subgroup H generator : ", h)
		t.Log("Subgroup order :", len(H))
		t.Log("Polynomial :", f.String())
		t.Log("Eval domain order :", len(evalDomain))
		t.Log("Merkle Commitment of evaluations :", hex.EncodeToString(fCommitment))
		t.Log("Chanel ", hex.EncodeToString(fsChannel.State))
		t.Log("Polynomial evaluations :", fEvals[0])
	})
	t.Run("TestConstraintEncoding", func(t *testing.T) {
		f := f.Clone(0)
		// The first constraint :
		// f(x) - 1 = 0 <=> f(x)-1/(X-1)
		num0 := f.Sub(poly.NewPolynomialInts(1), PrimeField.Modulus())
		dem0 := poly.NewPolynomialInts(-1, 1)
		quoPolyConstraint1, _ := num0.Div(dem0, PrimeField.Modulus())

		// Validate the first constraint
		if num0.Eval(nt.FromInt64(1), PrimeField.Modulus()).Cmp(nt.FromInt64(0)) != 0 {
			t.Fatal("first constraint not verified : wrong evaluation at x = 0")

		}
		if quoPolyConstraint1.Eval(nt.FromInt64(2718), PrimeField.Modulus()).Cmp(nt.FromInt64(2509888982)) != 0 {
			t.Fatal("first constraint not verified : wrong evaluation at x = 2718")
		}
		f = f.Clone(0)

		// The second constraint
		// f(x) - 2338775057 = 0 <=> f(x0) - 2338775057 / X - g^1022
		num1 := f.Sub(poly.NewPolynomialInts(2338775057), PrimeField.Modulus())
		dem1 := poly.NewPolynomialInts(0, 1).Sub(poly.NewPolynomial([]ff.FieldElement{g.Exp(nt.FromInt64(1022))}), PrimeField.Modulus())

		quoPolyConstraint2, _ := num1.Div(dem1, PrimeField.Modulus())

		if quoPolyConstraint2.Eval(nt.FromInt64(5772), PrimeField.Modulus()).Cmp(nt.FromInt64(232961446)) != 0 {
			t.Fatal("second constraint not verified : wrong evaluation at 5772")
		}

		// The third constraint requires polynomial composition
		// f(g^2.x) - f(g.x^2) - f(x)^2 / (X - g^k)
		fcompGSquared := f.Compose(poly.NewPolynomialBigInt(nt.FromInt64(0), g.Exp(nt.FromInt64(2)).Big()), PrimeField.Modulus())
		fcompG := f.Compose(poly.NewPolynomialBigInt(nt.FromInt64(0), g.Big()).Pow(nt.FromInt64(2), PrimeField.Modulus()), PrimeField.Modulus())
		fSquared := f.Pow(nt.FromInt64(2), PrimeField.Modulus())

		num2 := fcompGSquared.Sub(fcompG, PrimeField.Modulus()).Sub(fSquared, PrimeField.Modulus())
		dem2num := poly.NewPolynomialInts(0, 1).Clone(1023).Sub(poly.NewPolynomialInts(1), nil)

		coeffs := []ff.FieldElement{
			g.Exp(nt.FromInt64(1021)),
			g.Exp(nt.FromInt64(1022)),
			g.Exp(nt.FromInt64(1023)),
		}

		var terms []poly.Polynomial

		for _, coeff := range coeffs {
			monomial := poly.NewPolynomialInts(0, 1).Sub(poly.NewPolynomialBigInt(coeff.Big()), nil)
			terms = append(terms, monomial)
		}

		dem2dem := poly.NewPolynomialInts(1)

		for _, term := range terms {
			dem2dem = dem2dem.Mul(term, PrimeField.Modulus())
		}

		dem2, _ := dem2num.Div(dem2dem, PrimeField.Modulus())

		quoPolyConstraint3, _ := num2.Div(dem2, PrimeField.Modulus())

		expected := nt.FromInt64(2090051528)
		actual := quoPolyConstraint3.Eval(nt.FromInt64(31415), PrimeField.Modulus())
		if actual.Cmp(expected) != 0 {
			t.Fatal("third constraint not verified : wrong evaluation at 31415 , expected :", expected, " got :", actual)
		}

		// To generate succint proofs we transform the three polynomial validity checks
		// into one by applying a linear transform [a0,a1,a2]
		// the composition polynomial is written a0p0 + a1p1 + a2p2
		// where a0,a1,a2 are random field elements in this case extracted
		// from the fiat shamir channel

		constraints := []poly.Polynomial{quoPolyConstraint1, quoPolyConstraint2, quoPolyConstraint3}
		compositionPoly := poly.NewPolynomialInts(0)
		for i := 0; i < 3; i++ {

			randomFE := fsChannel.RandFE(PrimeField.Modulus())
			comb := constraints[i].Mul(poly.NewPolynomialBigInt(randomFE), PrimeField.Modulus())
			compositionPoly = compositionPoly.Add(comb, PrimeField.Modulus())
		}

	})
}
