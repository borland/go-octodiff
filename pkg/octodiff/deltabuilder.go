package octodiff

import "io"

type DeltaBuilder struct {
	ProgressReporter ProgressReporter
}

func NewDeltaBuilder() *DeltaBuilder {
	return &DeltaBuilder{
		ProgressReporter: NopProgressReporter(),
	}
}

const readBufferSize = 4 * 1024 * 1024

func (s *DeltaBuilder) Build(newFile io.Reader, signatureFile io.Reader, signatureFileLength int64, output io.Writer) error {
	signatureReader := NewSignatureReader()
	signatureReader.ProgressReporter = s.ProgressReporter

	signature, err := signatureReader.ReadSignature(signatureFile, signatureFileLength)
	if err != nil {
		return err
	}

	chunks := signature.Chunks
	hash, err := signature.HashAlgorithm.HashFromReader(newFile)
	if err != nil {
		return err
	}

	// deltaWriter.writeMetadata

	// suppress unused variable warnings
	print(chunks, hash)

	return nil
}
