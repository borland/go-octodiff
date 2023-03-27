package octodiff

import "io"

type DeltaWriter interface {
	WriteMetadata(hashAlgorithm HashAlgorithm, expectedNewFileHash []byte) error
	WriteCopyCommand(offset int64, length int64) error
	WriteDataCommand(source io.ReadSeeker, offset int64, length int64) error
}
