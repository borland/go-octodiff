package octodiff

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type mockReader struct {
	Callbacks []func(p []byte) (n int, err error)
}

func newMockReader(Callbacks ...func(p []byte) (n int, err error)) *mockReader {
	return &mockReader{Callbacks: Callbacks}
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if len(m.Callbacks) == 0 {
		return 0, errors.New("Read past end of providers")
	}
	callback := m.Callbacks[0]
	m.Callbacks = m.Callbacks[1:]
	return callback(p)
}

func (m *mockReader) AllCallbacksConsumed() bool {
	return len(m.Callbacks) == 0
}

func returnData(data []byte) func(p []byte) (n int, err error) {
	return func(p []byte) (int, error) {
		copy(p, data)
		return len(data), nil
	}
}

func returnDataWithError(data []byte, err error) func(p []byte) (n int, err error) {
	return func(p []byte) (int, error) {
		copy(p, data)
		return len(data), err
	}
}

func returnError(err error) func(p []byte) (n int, err error) {
	return returnDataWithError(nil, err)
}

func returnDataWithEof(data []byte) func(p []byte) (n int, err error) {
	return returnDataWithError(data, io.EOF)
}

var returnEof = returnDataWithEof(nil)

func TestReaderIterator_OneShotExactBufferSize(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnEof)

	var received []byte
	iter := NewReaderIteratorSize(reader, 5)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcde", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_OneShotExactBufferSizeInlineEof(t *testing.T) {
	reader := newMockReader(
		returnDataWithEof([]byte("abcde")))

	var received []byte
	iter := NewReaderIteratorSize(reader, 5)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcde", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_MultiShotExactBufferSize(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("fghij")),
		returnData([]byte("klmno")),
		returnEof)

	var received []byte
	iter := NewReaderIteratorSize(reader, 5)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcdefghijklmno", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_MultiShotExactBufferSizeInlineEof(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("fghij")),
		returnDataWithEof([]byte("klmno")))

	var received []byte
	iter := NewReaderIteratorSize(reader, 5)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcdefghijklmno", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_OneShotLargerBufferSize(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnEof)

	var received []byte
	iter := NewReaderIteratorSize(reader, 500)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcde", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_OneShotLargerBufferSizeInlineEof(t *testing.T) {
	reader := newMockReader(
		returnDataWithEof([]byte("abcde")))

	var received []byte
	iter := NewReaderIteratorSize(reader, 500)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcde", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_MultiShotLargerBufferSizeInlineEof(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("fghij")),
		returnDataWithEof([]byte("klmno")))

	var received []byte
	iter := NewReaderIteratorSize(reader, 500)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcdefghijklmno", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_DifferentBlockSizes(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("f")),
		returnData([]byte("ghijklmnop")),
		// Note deliberate zero byte block in the middle which io.Reader explicitly allows but discourages
		// https://github.com/golang/go/issues/27531
		returnData(make([]byte, 0)),
		returnData([]byte("qr")),
		returnEof)

	var received []byte
	iter := NewReaderIteratorSize(reader, 500)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcdefghijklmnopqr", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_Failure(t *testing.T) {
	reader := newMockReader(
		returnError(errors.New("x")))

	var received []byte
	iter := NewReaderIteratorSize(reader, 500)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.EqualError(t, iter.Err(), "x")
	assert.Equal(t, []byte(nil), received)

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_FailureAfterReading(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("f")),
		returnError(errors.New("x")))

	var received []byte
	iter := NewReaderIteratorSize(reader, 500)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.EqualError(t, iter.Err(), "x")
	assert.Equal(t, []byte("abcdef"), received)

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_NextAfterFailure(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("f")),
		returnError(errors.New("x")))

	var received []byte
	iter := NewReaderIteratorSize(reader, 500)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.EqualError(t, iter.Err(), "x")
	assert.Equal(t, []byte("abcdef"), received)

	assert.False(t, iter.Next())

	// doesn't impact the outcome
	assert.EqualError(t, iter.Err(), "x")
	assert.Equal(t, []byte("abcdef"), received)

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_FailureWhileAlsoProvidingData(t *testing.T) {
	// quoting https://pkg.go.dev/io#Reader
	// > Callers should always process the n > 0 bytes returned before considering the error err.
	// > Doing so correctly handles I/O errors that happen after reading some bytes and also both of the allowed EOF behaviors.
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("f")),
		returnDataWithError([]byte("ghi"), errors.New("x")))

	var received []byte
	iter := NewReaderIteratorSize(reader, 500)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.EqualError(t, iter.Err(), "x")
	assert.Equal(t, []byte("abcdefghi"), received)

	assert.True(t, reader.AllCallbacksConsumed())
}
