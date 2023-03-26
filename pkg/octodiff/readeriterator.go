package octodiff

import "io"

// ReaderIterator lets you treat reading bytes from an io.Reader as a for-loop.
// The logic of "read all the byes from an io.Reader" is surprisingly complex in Go.
// this struct wraps it up into a "Next/Current" iterator style object as used by bufio.Scanner, or IEnumerator in C#
// Important: Like bufio.Scanner, if Next() returns false you MUST check Err() to see if it failed
// Important: Calling Next() mutates the io.Reader so you can't create more than one ReaderIterator per reader
// Note: This is designed to be stack-allocated by the caller, so the New functions don't return pointers
type ReaderIterator struct {
	reader      io.Reader
	buffer      []byte
	isCompleted bool
	err         error

	Current []byte
}

func (b *ReaderIterator) Err() error {
	return b.err
}

// Next calls `Read` on the underlying reader, returning true
func (b *ReaderIterator) Next() bool {
	if b.isCompleted {
		return false // already completed
	}

	bytesRead, err := b.reader.Read(b.buffer)
	if err != nil {
		// last block. May or may not have data depending on underlying reader
		b.isCompleted = true
		if err != io.EOF {
			b.err = err
		}
	}
	// even if an error was returned (whether EOF or not), the reader can still provide data
	if bytesRead == len(b.buffer) { // don't slice the buffer if we read the whole thing
		b.Current = b.buffer
	} else {
		b.Current = b.buffer[:bytesRead]
	}
	// if we hit the last block AND there's no data to return, tell the caller we're done
	return bytesRead > 0 || !b.isCompleted
}

// NewReaderIterator creates an iterator, allocating a buffer of a default size
func NewReaderIterator(reader io.Reader) ReaderIterator {
	return NewReaderIteratorSize(reader, 1024*1024)
}

// NewReaderIteratorSize creates an iterator, allocating a buffer of `bufferSize`
func NewReaderIteratorSize(reader io.Reader, bufferSize int) ReaderIterator {
	return NewReaderIteratorBuffer(reader, make([]byte, bufferSize))
}

// NewReaderIteratorBuffer creates an iterator, referencing an already-allocated buffer
func NewReaderIteratorBuffer(reader io.Reader, buffer []byte) ReaderIterator {
	return ReaderIterator{
		reader:      reader,
		buffer:      buffer,
		isCompleted: false,
		err:         nil,
		Current:     nil,
	}
}
