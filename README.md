# zkstarks

This is a test-driven implementation of zkSTARKs based on the recent tutorials
by [StarkWare](https://github.com/starkware-industries/stark101).

To follow trough the implementations are separated into files and their respective
tests reading the files and following the Python example will be more fruitful
to your understanding.

You can see an example of the execution log [here](EXAMPLE.md).

The package implements both the primitives necessary for proof generation
and we use tests to generates a full uncompressed proof of the following statement :

```sh
I know a field element *X* such that the 1023rd element of the FibonacciSq sequence is 2338775057.
```

## Usage

The program will take some time to run polynomial interpolation and evaluation
are the most costly operations

```sh
go test -v -gcflags=all=-d=checkptr=0
```

* P.S : The checkptr flag crashes due to an unsafe conversion in Go's SHA3 implementation [issue](https://github.com/golang/go/issues/37644)

## Notes

Due to some intricacies and differences between languages, the hash values are different
from this implementation and the starkware one, we also use sha3 (Keccak-FIPS)
instead of sha256.
The difference in hash values is due to the internal encodings (Values to Bytes)
used by Python and Go.

## References

- [Arithmetization I](https://medium.com/starkware/arithmetization-i-15c046390862)
- [Arithmetization II](https://medium.com/starkware/arithmetization-ii-403c3b3f4355)
- [Scalable Transparent and Post-Quantum Proofs](https://eprint.iacr.org/2018/046)
