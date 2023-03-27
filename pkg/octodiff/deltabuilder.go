package octodiff

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"sort"
)

type DeltaBuilder struct {
	ProgressReporter ProgressReporter
}

func NewDeltaBuilder() *DeltaBuilder {
	return &DeltaBuilder{
		ProgressReporter: NopProgressReporter(),
	}
}

const readBufferSize = 4 * 1024 * 1024

// confusing naming: "newFile" isn't a new file that we are creating, but rather an existing file which is
// "new" in that we haven't created a delta for it yet.
func (d *DeltaBuilder) Build(newFile io.ReadSeeker, newFileLength int64, signatureFile io.Reader, signatureFileLength int64, output io.Writer) error {
	signatureReader := NewSignatureReader()
	signatureReader.ProgressReporter = d.ProgressReporter

	signature, err := signatureReader.ReadSignature(signatureFile, signatureFileLength)
	if err != nil {
		return err
	}

	chunks := signature.Chunks
	hash, err := signature.HashAlgorithm.HashOverReader(newFile)
	if err != nil {
		return err
	}
	_, err = newFile.Seek(0, io.SeekStart) // HashOverReader reads the entire newFile so we need to seek back to the start to process it
	if err != nil {
		return err
	}

	err = writeMetadata(output, signature.HashAlgorithm, hash)
	if err != nil {
		return err
	}

	sort.SliceStable(chunks, func(i, j int) bool {
		// C# ChunkSignatureChecksumComparer has secondary comparison based on chunk StartOffset but we don't need that as we are using a stable sort
		return chunks[i].RollingChecksum > chunks[j].RollingChecksum
	})

	chunkMap, minChunkSize, maxChunkSize := d.createChunkMap(chunks)

	lastMatchPosition := int64(0)
	buffer := make([]byte, readBufferSize)
	d.ProgressReporter.ReportProgress("Building delta", int64(0), newFileLength)

	startPosition := int64(0)

	for {
		bytesRead, fileReadErr := newFile.Read(buffer)
		if bytesRead > 0 { // we got some bytes, process them
			checksumAlgorithm := signature.RollingChecksumAlgorithm
			checksum := uint32(0)

			remainingPossibleChunkSize := maxChunkSize

			// slide a window over the buffer, looking for anything that matches our known list of chunks
			for i := 0; i < (bytesRead - minChunkSize + 1); i++ {
				readSoFar := startPosition + int64(i)

				remainingBytes := bytesRead - i
				if remainingBytes < maxChunkSize {
					remainingPossibleChunkSize = minChunkSize
				}

				if i == 0 || remainingBytes < maxChunkSize { // we are either at the start or end of buffer; calculate a full checksum
					checksum = checksumAlgorithm.Calculate(buffer[i : i+remainingPossibleChunkSize])
				} else { // we are stepping through the buffer, just rotate the existing checksum
					remove := buffer[i-1]
					add := buffer[i+remainingPossibleChunkSize-1]
					checksum = checksumAlgorithm.Rotate(checksum, remove, add, remainingPossibleChunkSize)
				}

				d.ProgressReporter.ReportProgress("Building delta", readSoFar, newFileLength)

				if readSoFar-(lastMatchPosition-int64(remainingPossibleChunkSize)) < int64(remainingPossibleChunkSize) {
					continue
				}

				startIndex, ok := chunkMap[checksum]
				if !ok {
					continue // we didn't match any known chunks. Skip, and the skipped data will be picked up later in a Data command based on lastMatchPosition
				}

				for j := startIndex; j < len(chunks) && chunks[j].RollingChecksum == checksum; j++ {
					chunk := chunks[j]

					sha := signature.HashAlgorithm.HashOverData(buffer[i : i+remainingPossibleChunkSize])

					if bytes.Equal(sha, chunk.Hash) {
						// we matched a chunk. Write any data in between it and the previous match as data, then write the 'copy' command for a chunk
						readSoFar = readSoFar + int64(remainingPossibleChunkSize)

						missing := readSoFar - lastMatchPosition
						if missing > int64(remainingPossibleChunkSize) {
							err = writeDataCommand(output, newFile, lastMatchPosition, missing-int64(remainingPossibleChunkSize))
							if err != nil {
								return err
							}
						}

						err = writeCopyCommand(output, chunk.StartOffset, int64(chunk.Length))
						if err != nil {
							return err
						}
						lastMatchPosition = readSoFar
						break
					}
				}
			}
		}
		if fileReadErr != nil {
			if fileReadErr == io.EOF { // all done
				break
			}
			return fileReadErr // something else went wrong after processing bytes, fail!
		}
		// If we didn't read a full buffer size, then assume we reached the end of newFile and exit the loop.
		// Note that Go's reader interface doesn't promise that it will always give you N bytes, even if N
		// bytes are available, in practice for file readers it does, and because we jump around seeking within
		// newFile there's not a great way of otherwise determining EOF without rewriting this whole thing
		if bytesRead < len(buffer) {
			break
		}

		// seek backwards by maxChunkSize+1 so we can read that stuff, and continue sliding the window over the file
		// note we mutate startPosition so it is ready for the next time round the loop
		startPosition, err = newFile.Seek(-int64(maxChunkSize)+1, io.SeekCurrent)
		if err != nil {
			return err
		}
	}

	// we've reached the end of the file. Write any trailing data as a 'Data' command
	if newFileLength != lastMatchPosition {
		err = writeDataCommand(output, newFile, lastMatchPosition, newFileLength-lastMatchPosition)
		if err != nil {
			return err
		}
	}

	return nil
}

// returns chunkMap, minChunkSize, maxChunkSize
func (d *DeltaBuilder) createChunkMap(chunks []*ChunkSignature) (map[uint32]int, int, int) {
	d.ProgressReporter.ReportProgress("Creating chunk map", 0, int64(len(chunks)))

	maxChunkSize := uint16(0)
	minChunkSize := uint16(math.MaxUint16)

	chunkMap := make(map[uint32]int)

	for chunkIdx, chunk := range chunks {
		if chunk.Length > maxChunkSize {
			maxChunkSize = chunk.Length
		}
		if chunk.Length < minChunkSize {
			minChunkSize = chunk.Length
		}

		if _, ok := chunkMap[chunk.RollingChecksum]; !ok {
			chunkMap[chunk.RollingChecksum] = chunkIdx
		}
		d.ProgressReporter.ReportProgress("Creating chunk map", int64(chunkIdx), int64(len(chunks)))
	}
	return chunkMap, int(minChunkSize), int(maxChunkSize)
}

func writeMetadata(output io.Writer, hashAlgorithm HashAlgorithm, expectedNewFileHash []byte) error {
	_, err := output.Write(BinaryDeltaHeader)
	if err != nil {
		return err
	}
	_, err = output.Write(BinaryVersion)
	if err != nil {
		return err
	}
	err = writeLengthPrefixedString(output, hashAlgorithm.Name())
	if err != nil {
		return err
	}
	err = binary.Write(output, binary.LittleEndian, int32(len(expectedNewFileHash)))
	if err != nil {
		return err
	}
	_, err = output.Write(expectedNewFileHash)
	if err != nil {
		return err
	}
	_, err = output.Write(BinaryEndOfMetadata)
	return err
}

// writes the "Copy Command" header to `output`
// followed by offset and length; There's no data
func writeCopyCommand(output io.Writer, offset int64, length int64) error {
	_, err := output.Write(BinaryCopyCommand)
	if err != nil {
		return err
	}
	err = binary.Write(output, binary.LittleEndian, offset)
	if err != nil {
		return err
	}
	return binary.Write(output, binary.LittleEndian, length)
}

// writes the "Data Command" header to `output`
// then proceeds to read `length` bytes from `source`, seeking to `offset` and write those to `output`
func writeDataCommand(output io.Writer, source io.ReadSeeker, offset int64, length int64) (err error) {
	_, err = output.Write(BinaryDataCommand)
	if err != nil {
		return
	}
	err = binary.Write(output, binary.LittleEndian, length)
	if err != nil {
		return
	}

	var originalPosition int64
	originalPosition, err = source.Seek(0, io.SeekCurrent) // doing a no-op seek is how you find out the current position of a Go reader
	if err != nil {
		return
	}
	// we need to ensure we seek back to originalPosition before exiting the function.
	defer func() {
		_, seekBackErr := source.Seek(originalPosition, io.SeekStart)
		if seekBackErr != nil {
			err = seekBackErr // this causes writeDataCommand to return this error
		}
	}()

	_, err = source.Seek(0, io.SeekStart)
	if err != nil {
		return
	}

	bufSize := 1024 * 1024
	if int(length) < bufSize {
		bufSize = int(length)
	}

	buffer := make([]byte, bufSize)
	soFar := int64(0)

	readNextBytes := func() (int, []byte, error) {
		subBufSize := bufSize
		if (length - soFar) < int64(subBufSize) {
			subBufSize = int(length - soFar)
		}
		subBuffer := buffer[:subBufSize]
		bytesRead, err := source.Read(subBuffer)

		if bytesRead != len(subBuffer) {
			return bytesRead, subBuffer[:bytesRead], nil
		}

		return bytesRead, subBuffer, err
	}

	bytesRead := 1 // placeholder to start loop
	for bytesRead > 0 {
		var subBuffer []byte
		bytesRead, subBuffer, err = readNextBytes()
		if err != nil {
			return
		}
		soFar += int64(bytesRead)

		_, err = output.Write(subBuffer)
		if err != nil {
			return
		}
	}

	return
}
