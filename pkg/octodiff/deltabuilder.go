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

	d.ProgressReporter.ReportProgress("Building delta", int64(0), newFileLength)

	startPosition, err := newFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	iter := NewReaderIteratorSize(newFile, readBufferSize)
	for iter.Next() {
		buffer := iter.Current
		bytesRead := len(buffer)

		checksumAlgorithm := signature.RollingChecksumAlgorithm
		checksum := uint32(0)

		remainingPossibleChunkSize := maxChunkSize

		for i := 0; i < (bytesRead - minChunkSize + 1); i++ {
			readSoFar := startPosition + int64(i)

			remainingBytes := bytesRead - i
			if remainingBytes < maxChunkSize {
				remainingPossibleChunkSize = minChunkSize
			}

			if i == 0 || remainingBytes < maxChunkSize {
				checksum = checksumAlgorithm.Calculate(buffer[i : i+remainingPossibleChunkSize])
			} else {
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
				continue
			}

			for j := startIndex; j < len(chunks) && chunks[j].RollingChecksum == checksum; j++ {
				chunk := chunks[j]

				sha := signature.HashAlgorithm.HashOverData(buffer[i : i+remainingPossibleChunkSize])
				if bytes.Equal(sha, chunk.Hash) {
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
					//deltaWriter.WriteCopyCommand(new DataRange(chunk.StartOffset, chunk.Length));
					lastMatchPosition = readSoFar
					break
				}
			}
		}

		if bytesRead < len(buffer) {
			break
		}

		// why are we seeking backwards? Need to debug the C# code and figure out what's going on here
		startPosition, err = newFile.Seek(-int64(maxChunkSize)+1, io.SeekCurrent)
		if err != nil {
			return err
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}

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
	err = binary.Write(output, binary.LittleEndian, len(expectedNewFileHash))
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
	originalPosition, err = source.Seek(offset, io.SeekCurrent)
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
