package zkstarks

import (
	"encoding/hex"
	"math/big"
	"strings"

	"golang.org/x/crypto/sha3"
)

// This file contains an implementation of the Fiat-Shamir heuristic channel
// The prover-verifier interaction is made non-interactive by evaluation
// and sending proofs to random points i.e leafs on the tree.
// On each call to send the channel state is update by transcript hash
// and the proof contains the transcript operations.

// Pragmatically this implementation simulates merlin transcripts where
// each "computation" adds some randomness to the state's hash which we
// use to simulate randomness when emulating prover-verifier interaction.
// More : https://merlin.cool/

var (
	sendOperator   = "send:"
	receiveRandInt = "receiveRandInt:"
	receiveRandFE  = "receiveRandFE:"
)

// Channel represents a FS transcript cache
type Channel struct {
	State []byte
	Proof []string
}

// NewChannel creates a new instance of the FS channel
func NewChannel() *Channel {
	return &Channel{
		State: []byte{0},
		Proof: make([]string, 0, 64),
	}
}

// Send appends items to the channel state by hashing them
func (ch *Channel) Send(s []byte) {
	var builder strings.Builder
	builder.WriteString(sendOperator)
	builder.WriteString(hex.EncodeToString(s))

	ch.Proof = append(ch.Proof, builder.String())
	ch.State = hash(concat(ch.State, s))
}

// RandInt emulates a random integer scalar in the range [min,max]
// sent by the verifier
func (ch *Channel) RandInt(min, max *big.Int) *big.Int {

	stateAsInt := new(big.Int).SetBytes(ch.State)
	diff := new(big.Int).Sub(max, min)
	diff = diff.Add(diff, big.NewInt(1))
	reduced := new(big.Int).Mod(stateAsInt, diff)

	num := new(big.Int).Add(min, reduced)

	var builder strings.Builder
	builder.WriteString(receiveRandInt)
	builder.WriteString(num.String())

	ch.Proof = append(ch.Proof, builder.String())
	ch.State = hash(ch.State)
	return num

}

// RandFE emulates a random field element sent by the verifier given the field's
// modulus.
func (ch *Channel) RandFE(m *big.Int) *big.Int {
	max := new(big.Int).Sub(m, big.NewInt(1))
	num := ch.RandInt(big.NewInt(0), new(big.Int).Set(max))

	var builder strings.Builder
	builder.WriteString(receiveRandFE)
	builder.WriteString(num.String())

	return num

}
func concat(a, b []byte) []byte {
	return append(a, b...)
}
func hash(b []byte) []byte {
	h := sha3.Sum256(b)
	return h[:]
}
