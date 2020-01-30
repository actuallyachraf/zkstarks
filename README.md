# zkstarks

Status : **WIP**

This is a literate implementation of zkSTARKs based on the recent tutorials
by [StarkWare](https://github.com/starkware-industries/stark101).

## Usage

The program will take some time to run polynomial interpolation and evaluation
are the most costly operations

```sh
go run cmd/main.go
```

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
