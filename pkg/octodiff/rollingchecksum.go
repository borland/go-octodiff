package octodiff

type RollingChecksum interface {
	Name() string
	Calculate(block []byte, offset int, count int) uint32
	Rotate(checksum uint32, remove byte, add byte, chunkSize int) uint32
}
