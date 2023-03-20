package octodiff

// TODO Obsolete: This is non standard implimentation of Adler32, Adler32RollingChecksumV2 should be used instead.

type Adler32RollingChecksum struct{}

func (_ *Adler32RollingChecksum) Name() string {
	return "Adler32"
}

func (_ *Adler32RollingChecksum) Calculate(block []byte, offset int, count int) uint32 {
	a := uint32(1)
	b := uint32(0)

	for i := offset; i < offset+count; i++ {
		z := block[i]
		a = uint32(z) + a
		b = b + a
	}
	return (b << 16) | a
}

func (_ *Adler32RollingChecksum) Rotate(checksum uint32, remove byte, add byte, chunkSize int) uint32 {
	b := checksum >> 16 & 0xffff
	a := checksum & 0xffff

	a = a - uint32(remove) + uint32(add)
	b = b - (uint32(chunkSize) * uint32(remove)) + a - 1

	return (b << 16) | a
}

var _ RollingChecksum = (*Adler32RollingChecksum)(nil)
