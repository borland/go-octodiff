package octodiff_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/OctopusDeploy/go-octodiff/pkg/octodiff"
	"github.com/stretchr/testify/assert"
	"testing"
)

func logDeltaFile(input []byte) []string {
	actions := make([]string, 0)

	reader := octodiff.NewBinaryDeltaReader(bytes.NewReader(input))
	err := reader.Apply(
		func(bytes []byte) error {
			actions = append(actions, fmt.Sprintf("write %v", hex.EncodeToString(bytes)))
			return nil
		}, func(start int64, length int64) error {
			actions = append(actions, fmt.Sprintf("copy start=%v, length=%v", start, length))
			return nil
		})

	if err != nil {
		panic(err) // shouldn't happen under tests
	}

	return actions
}

func TestReadsNoOpDeltaFile(t *testing.T) {
	input, _ := hex.DecodeString("4f43544f44454c544101045348413114000000330bd06982d3b5dbda6c1a6ad16687a0cdb03c0d3e3e3e6000000000000000000802000000000000")

	assert.Equal(t, []string{
		"copy start=0, length=520",
	}, logDeltaFile(input))
}

func TestReadsFullDeltaFile(t *testing.T) {
	input, _ := hex.DecodeString("4f43544f44454c544101045348413114000000330bd06982d3b5dbda6c1a6ad16687a0cdb03c0d3e3e3e80080200000000000030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e0326453770")

	assert.Equal(t, []string{
		"write 30820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e0326453770",
	}, logDeltaFile(input))
}

func TestReadsDeltaFileWithSmallChanges(t *testing.T) {
	input, _ := hex.DecodeString("4f43544f44454c54410104534841311400000014ba64eeabad295dd60cfabd648ff176b64890773e3e3e80800000000000000030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7aaabac300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06036080000000000000008801000000000000")

	assert.Equal(t, []string{
		"write 30820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7aaabac300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b0603",
		"copy start=128, length=392",
	}, logDeltaFile(input))
}

func TestReadsDeltaFileWithSmallChanges_Prepend(t *testing.T) {
	input, _ := hex.DecodeString("4f43544f44454c544101045348413114000000e5ca5051b8cf462ed567a8f88802fd9e62a0f0e83e3e3e800100000000000000aa6000000000000000000802000000000000")

	assert.Equal(t, []string{
		"write aa",
		"copy start=0, length=520",
	}, logDeltaFile(input))
}

func TestReadsDeltaFileWithSmallChangesLargerFile(t *testing.T) {
	input, _ := hex.DecodeString("4f43544f44454c54410104534841311400000050f7ec0e6d4fe4ab8400b759e2b000f1d0aced8e3e3e3e80280000000000000030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7aaabac300a06082a6000780000000000000088010000000000800800000000000000a3bec4300a06082a600078000000000000007000000000000080d007000000000000060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7")

	assert.Equal(t, []string{
		"write 30820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7aaabac300a06082a",
		"copy start=30720, length=100352",
		"write a3bec4300a06082a",
		"copy start=30720, length=28672",
		"write 060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7",
	}, logDeltaFile(input))
}

func TestReadsDeltaFileWithSmallChangesLargerFile_DisjointChanges(t *testing.T) {
	input, _ := hex.DecodeString("4f43544f44454c544101045348413114000000645a41cab32226e8e9212c54db711c22653c00513e3e3e80280000000000000030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7aaabac300a06082a600078000000000000007800000000000080b00c00000000000061746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03aa0703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0cab522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652dac746174653117306000d00000000000000030010000000000800800000000000000a3bec4300a06082a600078000000000000004800000000000080200300000000000006082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7")

	assert.Equal(t, []string{
		"write 30820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7aaabac300a06082a",
		"copy start=30720, length=30720",
		"write 61746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03aa0703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0cab522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652dac74617465311730",
		"copy start=53248, length=77824",
		"write a3bec4300a06082a",
		"copy start=30720, length=18432",
		"write 06082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80314a57f37d0931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7",
	}, logDeltaFile(input))
}
