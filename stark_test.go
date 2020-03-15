package zkstarks

import (
	"encoding/hex"
	"testing"

	"github.com/actuallyachraf/algebra/nt"
	"github.com/actuallyachraf/algebra/poly"
	"github.com/stretchr/testify/assert"
)

func TestZKGen(t *testing.T) {

	a, g, G, h, H, evalDomain, f, fEvals, fCommitment, fsChannel := GenerateDomainParameters()

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
		assert.Equal(t, num0.Eval(nt.FromInt64(1), PrimeField.Modulus()), nt.FromInt64(0))
		assert.Equal(t, quoPolyConstraint1.Eval(nt.FromInt64(2718), PrimeField.Modulus()), 2509888982)

	})
}
