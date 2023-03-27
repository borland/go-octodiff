package octodiff_test

import (
	"bytes"
	"encoding/hex"
	"github.com/OctopusDeploy/go-octodiff/pkg/octodiff"
	"github.com/OctopusDeploy/go-octodiff/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func buildDelta(newFile []byte, signatureFile []byte) []byte {
	d := octodiff.NewDeltaBuilder()
	newFileReader := bytes.NewReader(newFile)
	signatureFileReader := bytes.NewReader(signatureFile)

	var output bytes.Buffer
	err := d.Build(newFileReader, int64(len(newFile)), signatureFileReader, int64(len(signatureFile)), &output)
	if err != nil {
		panic(err) // should never fail under tests
	}
	return output.Bytes()
}

func buildSignature(input []byte) []byte {
	var output bytes.Buffer
	err := octodiff.NewSignatureBuilder().Build(bytes.NewReader(input), int64(len(input)), &output)
	if err != nil {
		panic(err) // should never fail under tests
	}
	return output.Bytes()
}

func TestBuildsNoOpDeltaForSameInput(t *testing.T) {
	signature := buildSignature(test.TestData())

	deltaFile := buildDelta(test.TestData(), signature)

	assert.Equal(t, "4f43544f44454c544101045348413114000000330bd06982d3b5dbda6c1a6ad16687a0cdb03c0d3e3e3e6000000000000000000802000000000000", hex.EncodeToString(deltaFile))
}

func TestBuildsFullDeltaForNoInput(t *testing.T) {
	signature := buildSignature(nil)

	deltaFile := buildDelta(test.TestData(), signature)

	assert.Equal(t, "4f43544f44454c544101045348413114000000330bd06982d3b5dbda6c1a6ad16687a0cdb03c0d3e3e3e80080200000000000030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e0326453770", hex.EncodeToString(deltaFile))
}
