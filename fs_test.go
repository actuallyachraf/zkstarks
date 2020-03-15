package zkstarks

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/actuallyachraf/algebra/nt"
)

func TestFiatShamirChannel(t *testing.T) {

	t.Run("TestFiatShamirReproducible", func(t *testing.T) {
		c := NewChannel()

		c.Send([]byte("Yes"))
		r1 := c.RandInt(nt.FromInt64(0), nt.FromInt64(math.MaxUint32))

		d := NewChannel()

		d.Send([]byte("Yes"))
		r2 := d.RandInt(nt.FromInt64(0), nt.FromInt64(math.MaxUint32))

		if r1.Cmp(r2) != 0 {
			t.Fatal("error FiatShamir channel should be reproducible")
		}
	})

	t.Run("TestFiatShamirUniformity", func(t *testing.T) {

		var rangeSize int64 = 10
		var upperBound int64 = 1048576
		var numTries int64 = 1024

		c := NewChannel()
		n := rand.Int63n(upperBound)

		var intBuf = new(bytes.Buffer)
		err := binary.Write(intBuf, binary.LittleEndian, n)
		if err != nil {
			t.Fatal("failed to write integer to buffer")
		}
		c.Send(intBuf.Bytes())

		dist := make([]*big.Int, 0, numTries)
		distRand := make([]*big.Int, 0, numTries)

		var i int64 = 0
		for i = 0; i < numTries; i++ {
			dist = append(dist, c.RandInt(nt.FromInt64(0), nt.FromInt64(rangeSize-1)))
			distRand = append(distRand, big.NewInt(rand.Int63n(rangeSize-1)))
		}

		mean := numTries / rangeSize

		var yMeanChannel = big.NewInt(0)
		var yMeanRand = big.NewInt(0)

		for _, y := range dist {

			yNorm := nt.ModExp(nt.Sub(y, nt.FromInt64(mean)), nt.FromInt64(2), nil)
			yMeanChannel = nt.Add(yMeanChannel, yNorm)
		}

		for _, y := range distRand {

			yNorm := nt.ModExp(nt.Sub(y, nt.FromInt64(mean)), nt.FromInt64(2), nil)
			yMeanRand = nt.Add(yMeanRand, yNorm)
		}

		normalizedStdDevChannel := nt.Div(new(big.Int).Sqrt(yMeanChannel), nt.FromInt64(rangeSize))
		normalizedStdDevRand := nt.Div(new(big.Int).Sqrt(yMeanRand), nt.FromInt64(rangeSize))

		if (nt.Sub(normalizedStdDevChannel, normalizedStdDevRand)).Cmp(nt.FromInt64(4)) > 0 {
			t.Error("fiat shamir channel is not uniform")
		}
	})

}
