package octodiff

type Signature struct {
	HashAlgorithm            HashAlgorithm
	RollingChecksumAlgorithm RollingChecksum
	Chunks                   []*ChunkSignature
}

type ChunkSignature struct {
	Length          uint16
	Hash            []byte
	RollingChecksum uint32
}
