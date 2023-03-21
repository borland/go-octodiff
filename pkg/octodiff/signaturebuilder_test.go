package octodiff_test

import (
	"bytes"
	"encoding/hex"
	"github.com/OctopusDeploy/go-octodiff/pkg/octodiff"
	"github.com/OctopusDeploy/go-octodiff/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func buildSignature(builder *octodiff.SignatureBuilder, input []byte) []byte {
	var buf bytes.Buffer
	err := builder.Build(bytes.NewReader(input), int64(len(input)), &buf)
	if err != nil {
		panic(err) // should never fail under tests
	}
	return buf.Bytes()
}

func TestBuildSignatureWithDefaultSettingsAndTestData(t *testing.T) {
	b := octodiff.NewSignatureBuilder()
	result := buildSignature(b, test.TestData)

	assert.Equal(t, "4f43544f5349470104534841310741646c657233323e3e3e0802f79fa2f0330bd06982d3b5dbda6c1a6ad16687a0cdb03c0d", hex.EncodeToString(result))
}

func TestBuildSignatureWithSmallChunkSize(t *testing.T) {
	b := octodiff.NewSignatureBuilder()
	b.ChunkSize = octodiff.SignatureMinimumChunkSize // the smaller the chunk size, the larger the signature file needs to be
	result := buildSignature(b, test.TestData)

	assert.Equal(t, "4f43544f5349470104534841310741646c657233323e3e3e8000951f26e719f3978cb607e80a9aab3abbcac8bb1ecbcecf3e80001f18260f0f73196c2aa57877ee5e31291a59b5afca4493658000e035f42a42c4a73471dea3b9746e22dd93893fd8549f11bd8000dd2ff46b72e00e30ecae4c70ee07721d221a3b8a6d1847fa08008a02860c21d4023a8ba580ecdba742e7400aa40b6e449bb3", hex.EncodeToString(result))
}

func TestBuildSignatureWithLargeChunkSize(t *testing.T) {
	b := octodiff.NewSignatureBuilder()
	b.ChunkSize = octodiff.SignatureMaximumChunkSize // our input file is small so larger chunk size doesn't help here
	result := buildSignature(b, test.TestData)

	assert.Equal(t, "4f43544f5349470104534841310741646c657233323e3e3e0802f79fa2f0330bd06982d3b5dbda6c1a6ad16687a0cdb03c0d", hex.EncodeToString(result))
}

func TestBuildSignatureWithAdlerV2(t *testing.T) {
	b := octodiff.NewSignatureBuilder()
	b.RollingChecksumAlgorithm = octodiff.NewAdler32RollingChecksumV2()
	result := buildSignature(b, test.TestData)

	assert.Equal(t, "4f43544f5349470104534841310941646c6572333256323e3e3e0802f79fe5f8330bd06982d3b5dbda6c1a6ad16687a0cdb03c0d", hex.EncodeToString(result))
}
