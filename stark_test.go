package zkstarks

import (
	"encoding/hex"
	"io/ioutil"
	"testing"

	"github.com/actuallyachraf/algebra/ff"
	"github.com/actuallyachraf/algebra/nt"
	"github.com/actuallyachraf/algebra/poly"
	"github.com/stretchr/testify/assert"
)

func TestZKGen(t *testing.T) {

	/*
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

		domainParamsJSON, err := domainparamsInstance.MarshalJSON()
		if err != nil {
			t.Fatal("failed to serialize domain params to JSON")
		}
		err = ioutil.WriteFile("./domainparamsInstance.json", domainParamsJSON, 0711)
		if err != nil {
			t.Fatal("failed to serialize domain params to JSON")
		}
	*/
	paramBytes, err := ioutil.ReadFile("domainparams.json")

	paramsInstance := &DomainParameters{}
	err = paramsInstance.UnmarshalJSON(paramBytes)
	if err != nil {
		t.Fatal("failed to unmarshal domain params with error :", err)
	}
	_, g, _, _, _, _, f, _, _ := paramsInstance.Trace, paramsInstance.GeneratorG, paramsInstance.SubgroupG, paramsInstance.GeneratorH, paramsInstance.SubgroupH, paramsInstance.EvaluationDomain, paramsInstance.Polynomial, paramsInstance.PolynomialEvaluations, paramsInstance.EvaluationRoot
	fsChannel := NewChannel()
	fsChannel.Send(paramsInstance.EvaluationRoot)
	t.Run("TestParamGen", func(t *testing.T) {
		t.Log("Trace length :", len(paramsInstance.Trace))
		t.Log("Subgroup G generator :", paramsInstance.GeneratorG)
		t.Log("Subgroup order:", len(paramsInstance.SubgroupG))
		t.Log("Subgroup H generator : ", paramsInstance.GeneratorH)
		t.Log("Subgroup order :", len(paramsInstance.SubgroupH))
		t.Log("Polynomial :", paramsInstance.Polynomial.String())
		t.Log("Eval domain order :", len(paramsInstance.EvaluationDomain))
		t.Log("Merkle Commitment of evaluations :", hex.EncodeToString(paramsInstance.EvaluationRoot))
		t.Log("Chanel ", hex.EncodeToString(fsChannel.State))
	})
	t.Run("TestProve", func(t *testing.T) {
		f := f.Clone(0)

		quoPolyConstraint1, quoPolyConstraint2, quoPolyConstraint3 := GenerateProgramConstraints(f, g)

		if quoPolyConstraint1.Eval(nt.FromInt64(2718), PrimeField.Modulus()).Cmp(nt.FromInt64(2509888982)) != 0 {
			t.Fatal("first constraint not verified : wrong evaluation at x = 2718")
		}

		if quoPolyConstraint2.Eval(nt.FromInt64(5772), PrimeField.Modulus()).Cmp(nt.FromInt64(232961446)) != 0 {
			t.Fatal("second constraint not verified : wrong evaluation at 5772")
		}

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
		t.Log("Composition Polynomial :", compositionPoly)

		// Now we evaluate the composition polynomial on the evaluation domain
		// and commit to the evaluation
		compositionPolyEvals := make([]ff.FieldElement, len(paramsInstance.EvaluationDomain))
		for idx, elem := range paramsInstance.EvaluationDomain {

			eval := compositionPoly.Eval(elem.Big(), PrimeField.Modulus())
			compositionPolyEvals[idx] = PrimeField.NewFieldElement(eval)
		}
		compositionPolyEvalsRoot := DomainHash(compositionPolyEvals)

		t.Log("Composition Polynomial Evaluations Root :", hex.EncodeToString(compositionPolyEvalsRoot))
		fsChannel.Send(compositionPolyEvalsRoot)

		friDomains, friPolys, friLayers, friRoots := GenerateFRICommitment(compositionPoly, paramsInstance.EvaluationDomain, compositionPolyEvals, compositionPolyEvalsRoot, *fsChannel)

		assert.Len(t, friLayers, 11)
		assert.Len(t, friLayers[len(friLayers)-1], 8)
		t.Log("FRI Domains Count :", len(friDomains))
		t.Log("Last Layer :", friLayers[len(friLayers)-1])
		assert.Equal(t, friPolys[len(friPolys)-1].Degree(), 0)
		t.Log("Last Layer Root :", hex.EncodeToString(friRoots[len(friRoots)-1]))
		t.Log("Channel Proof", fsChannel.Proof)

	})
	t.Run("TestFRICommitment", func(t *testing.T) {
		testPoly := poly.NewPolynomialInts(2, 3, 0, 1)
		testDomain := []ff.FieldElement{PrimeField.NewFieldElementFromInt64(3), PrimeField.NewFieldElementFromInt64(5)}
		testBeta := PrimeField.NewFieldElementFromInt64(7)

		nextDomain, nextPoly, nextLayer := NextFRILayer(testDomain, testPoly, testBeta)

		actualPoly := poly.NewPolynomialInts(23, 7)
		actualDomain := []ff.FieldElement{PrimeField.NewFieldElementFromInt64(9)}
		actualLayer := []ff.FieldElement{PrimeField.NewFieldElementFromInt64(86)}

		assert.Equal(t, actualPoly, nextPoly)
		assert.Equal(t, actualDomain, nextDomain)
		assert.Equal(t, actualLayer, nextLayer)

	})
}
