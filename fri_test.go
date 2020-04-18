package zkstarks

import (
	"testing"

	"github.com/actuallyachraf/algebra/ff"
	"github.com/actuallyachraf/algebra/poly"
	"github.com/stretchr/testify/assert"
)

func TestFRIOperations(t *testing.T) {

	t.Run("TestPolynomialIndexer", func(t *testing.T) {

		p := poly.NewPolynomialInts(1, 2, 3, 4, 5, 6)
		evenP := evenCoeffs(p)
		oddP := oddCoeffs(p)

		actualOddP := poly.NewPolynomialInts(2, 4, 6)
		actualEvenP := poly.NewPolynomialInts(1, 3, 5)

		assert.Equal(t, evenP, actualEvenP)
		assert.Equal(t, oddP, actualOddP)

	})
	t.Run("TestNextFRILayer", func(t *testing.T) {

		testPoly := poly.NewPolynomialInts(2, 3, 0, 1)
		testDomain := []ff.FieldElement{PrimeField.NewFieldElementFromInt64(3), PrimeField.NewFieldElementFromInt64(5)}
		beta := PrimeField.NewFieldElementFromInt64(7)

		nextDomain, nextPoly, nextLayer := NextFRILayer(testDomain, testPoly, beta)

		actualNextPoly := poly.NewPolynomialInts(23, 7)
		actualNextDomain := []ff.FieldElement{PrimeField.NewFieldElementFromInt64(9)}
		actualNextLayer := []ff.FieldElement{PrimeField.NewFieldElementFromInt64(86)}
		assert.Equal(t, nextPoly, actualNextPoly)
		assert.Equal(t, nextDomain, actualNextDomain)
		assert.Equal(t, nextLayer, actualNextLayer)
	})

}
