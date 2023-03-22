package octodiff

import (
	"crypto/sha1"
	"io"
)

type HashAlgorithm interface {
	Name() string
	HashLength() int
	HashFromData(data []byte) []byte
	HashFromReader(reader io.Reader) ([]byte, error)
}

// the only hash algorithm octodiff seems to use is sha1

type Sha1HashAlgorithm struct {
}

func (s *Sha1HashAlgorithm) Name() string {
	return "SHA1"
}

// returns the length in bytes of a Hash computed with this aglorithm
func (s *Sha1HashAlgorithm) HashLength() int {
	return sha1.Size
}

func (s *Sha1HashAlgorithm) HashFromData(data []byte) []byte {
	h := sha1.Sum(data)
	return h[:] // convert from fixed-length array to slice
}

// This will issue lots of 1k reads into the reader.
// It's up to the caller to pass us a bufio if performance is of concern
func (s *Sha1HashAlgorithm) HashFromReader(reader io.Reader) ([]byte, error) {

	sha := sha1.New()
	block := make([]byte, 1024)

	bytesRead := 1
	var err error

	for bytesRead > 0 {
		bytesRead, err = reader.Read(block)
		if err == io.EOF {
			// possible last block. In practice this shouldn't happen as the last read should return success with 0 bytes instead, but good to be defensive
			return sha.Sum(block[:bytesRead]), nil
		}
		if err != nil {
			return nil, err
		}
		_, err = sha.Write(block[:bytesRead])
		if err != nil {
			return nil, err
		}
	}

	return sha.Sum(nil), nil
}

var DefaultHashAlgorithm HashAlgorithm = &Sha1HashAlgorithm{}
