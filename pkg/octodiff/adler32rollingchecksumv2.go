package octodiff

type Adler32RollingChecksumV2 struct{}

func NewAdler32RollingChecksumV2() *Adler32RollingChecksumV2 {
	return &Adler32RollingChecksumV2{}
}

const Modulus = uint32(65521)

func (_ *Adler32RollingChecksumV2) Name() string {
	return "Adler32V2"
}

func (_ *Adler32RollingChecksumV2) Calculate(block []byte) uint32 {
	a := uint32(1)
	b := uint32(0)

	for _, z := range block {
		a = (uint32(z) + a) % Modulus
		b = (b + a) % Modulus
	}
	return (b << 16) | a
}

func (_ *Adler32RollingChecksumV2) Rotate(checksum uint32, remove byte, add byte, chunkSize int) uint32 {
	b := checksum >> 16 & 0xffff
	a := checksum & 0xffff

	a = ((a - uint32(remove) + uint32(add)) % Modulus) & 0xffff
	b = ((b - (uint32(chunkSize) * uint32(remove)) + a - 1) % Modulus) & 0xffff

	return (b << 16) | a
}

var _ RollingChecksum = (*Adler32RollingChecksumV2)(nil)
