package zkstarks

import (
	"github.com/actuallyachraf/algebra/poly"
)

// The FibSeq program we want to prove validity statements can be
// proven correct if some constraints over it's output are valid.
// Mainly if FibSeq represents the trace of computation of the Fibonnaci
// sequence then :
// FibSeq(0) = 1
// FibSeq(1022) = 2338775057
// For all i, FibSeq(i+2) = FibSeq(i+1)^2 + FibSeq(i)^2
// Our goal in this part is two folds :
// - Set the constraints of the program (we just did that)
// - Encode the constraint into statements about polynomials (Algebraic transformation)
// - Encode the polynomial constraints into rational functions
// Since the trace of the computation is just a list of numbers
// and due to the unisolvence theorem only one polynomial goes trough them
// which we found trough lagrange interpolation what remains is
// encoding the three constraitns about into algebraic statements
// about the interpolated polynomial which we'll call f
// The constraints in Algebraic form :
// FibSeq(0) = 1 => q(x) = f(x) - 1 = 0 for x = g^0 = 1
// FibSeq(1022) = 2338775057 => r(x) = f(x) - 2338775057 = 0 for x = g^1022
// FibSeq(i+2) = FibSeq(i+1)^2 + FibSeq(i)^2 => f(g(x)^2) - f(g(x))^2 - f(x)^2.

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

}
