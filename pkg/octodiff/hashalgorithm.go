package octodiff

import (
	"crypto/sha1"
)

type HashAlgorithm interface {
	Name() string
	HashLength() int
	//ComputeHash(reader io.Reader) []byte
	ComputeHash(buffer []byte) []byte
}

// the only hash algorithm octodiff seems to use is sha1

type Sha1HashAlgorithm struct {
}

func (s *Sha1HashAlgorithm) Name() string {
	return "SHA1"
}

func (s *Sha1HashAlgorithm) HashLength() int {
	return sha1.BlockSize
}

func (s *Sha1HashAlgorithm) ComputeHash(buffer []byte) []byte {
	sum := sha1.Sum(buffer)
	return sum[:] // convert fixed-length array to slice
}

var _ HashAlgorithm = (*Sha1HashAlgorithm)(nil)

var DefaultHashAlgorithm HashAlgorithm = &Sha1HashAlgorithm{}
