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

func (s *DeltaBuilder) Build(newFile io.Reader, signatureFile io.Reader, output io.Writer) error {
	return nil
}
