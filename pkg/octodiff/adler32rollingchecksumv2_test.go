package octodiff_test

import (
	"github.com/OctopusDeploy/go-octodiff/pkg/octodiff"
	"github.com/OctopusDeploy/go-octodiff/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAdler32RollingChecksumV2_Name(t *testing.T) {
	c := &octodiff.Adler32RollingChecksumV2{}

	assert.Equal(t, "Adler32V2", c.Name())
}

func TestAdler32RollingChecksumV2_Calculate(t *testing.T) {
	c := &octodiff.Adler32RollingChecksumV2{}
	block := test.TestData

	assert.Equal(t, uint32(2760448612), c.Calculate(block, 0, 100))
	assert.Equal(t, uint32(2892962471), c.Calculate(block, 1, 100))
	assert.Equal(t, uint32(2481658437), c.Calculate(block, 2, 100))
	assert.Equal(t, uint32(595858050), c.Calculate(block, 93, 100))
	assert.Equal(t, uint32(4175798263), c.Calculate(block, 0, len(block)))
}

func TestAdler32RollingChecksumV2_Rotate(t *testing.T) {

	c := &octodiff.Adler32RollingChecksumV2{}
	assert.Equal(t, uint32(3209698067), c.Rotate(uint32(2755533412), 0, 0xAF, 8))
	assert.Equal(t, uint32(3209698067), c.Rotate(uint32(2755533412), 0, 0xAF, 16))
	assert.Equal(t, uint32(3209698067), c.Rotate(uint32(2755533412), 0, 0xAF, 24))
	assert.Equal(t, uint32(3209698067), c.Rotate(uint32(2755533412), 0, 0xAF, 32))

	assert.Equal(t, uint32(3577289570), c.Rotate(uint32(3209698067), 0xAF, 0xFE, 8))
	assert.Equal(t, uint32(3485539170), c.Rotate(uint32(3209698067), 0xAF, 0xFE, 16))
	assert.Equal(t, uint32(3393788770), c.Rotate(uint32(3209698067), 0xAF, 0xFE, 24))
	assert.Equal(t, uint32(3302038370), c.Rotate(uint32(3209698067), 0xAF, 0xFE, 32))
}
