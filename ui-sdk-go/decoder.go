package uimessagestream

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
)

// ErrStreamDone is returned when the stream has been terminated with [DONE].
var ErrStreamDone = errors.New("stream done")

// Decoder reads UI message stream chunks from an SSE stream.
type Decoder struct {
	reader *bufio.Reader
	done   bool
}

// NewDecoder creates a new Decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		reader: bufio.NewReader(r),
	}
}

// Decode reads and returns the next chunk from the stream.
// Returns ErrStreamDone when the [DONE] marker is encountered.
// Returns io.EOF when the stream ends without [DONE].
func (d *Decoder) Decode() (Chunk, error) {
	if d.done {
		return nil, ErrStreamDone
	}

	for {
		line, err := d.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF && len(line) == 0 {
				return nil, io.EOF
			}
			// Try to process any remaining data
			if len(line) == 0 {
				return nil, err
			}
		}

		// Trim whitespace
		line = bytes.TrimSpace(line)

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Check for "data:" prefix
		if !bytes.HasPrefix(line, []byte("data:")) {
			// Skip lines that don't start with "data:" (like comments or event types)
			continue
		}

		// Extract the data after "data:"
		data := bytes.TrimPrefix(line, []byte("data:"))
		data = bytes.TrimSpace(data)

		// Check for [DONE] marker
		if string(data) == DoneMarker {
			d.done = true
			return nil, ErrStreamDone
		}

		// Parse the JSON chunk
		chunk, err := ParseChunk(data)
		if err != nil {
			return nil, err
		}

		return chunk, nil
	}
}

// IsDone returns true if the stream has been terminated with [DONE].
func (d *Decoder) IsDone() bool {
	return d.done
}

// DecodeAll reads all chunks from the stream until [DONE] or EOF.
func (d *Decoder) DecodeAll() ([]Chunk, error) {
	var chunks []Chunk

	for {
		chunk, err := d.Decode()
		if err != nil {
			if err == ErrStreamDone || err == io.EOF {
				return chunks, nil
			}
			return chunks, err
		}
		chunks = append(chunks, chunk)
	}
}

// StreamReader provides a convenient interface for reading a complete stream.
type StreamReader struct {
	dec       *Decoder
	messageID string
	metadata  any
	err       error
}

// NewStreamReader creates a new StreamReader that reads from r.
func NewStreamReader(r io.Reader) *StreamReader {
	return &StreamReader{
		dec: NewDecoder(r),
	}
}

// Next reads and returns the next chunk from the stream.
// Returns ErrStreamDone when the stream is complete.
// The first chunk should be a StartChunk which sets the message ID and metadata.
func (sr *StreamReader) Next() (Chunk, error) {
	if sr.err != nil {
		return nil, sr.err
	}

	chunk, err := sr.dec.Decode()
	if err != nil {
		sr.err = err
		return nil, err
	}

	// Track message ID and metadata from start chunk
	if start, ok := chunk.(*StartChunk); ok {
		sr.messageID = start.MessageID
		sr.metadata = start.MessageMetadata
	}

	// Update metadata from metadata chunk
	if meta, ok := chunk.(*MessageMetadataChunk); ok {
		sr.metadata = meta.MessageMetadata
	}

	// Update metadata from finish chunk
	if finish, ok := chunk.(*FinishChunk); ok {
		if finish.MessageMetadata != nil {
			sr.metadata = finish.MessageMetadata
		}
	}

	return chunk, nil
}

// MessageID returns the message ID from the start chunk.
func (sr *StreamReader) MessageID() string {
	return sr.messageID
}

// Metadata returns the current message metadata.
func (sr *StreamReader) Metadata() any {
	return sr.metadata
}

// Err returns any error that occurred during reading.
func (sr *StreamReader) Err() error {
	return sr.err
}

// IsDone returns true if the stream has been terminated.
func (sr *StreamReader) IsDone() bool {
	return sr.dec.IsDone()
}

// ChunkIterator provides an iterator interface for reading chunks.
type ChunkIterator struct {
	dec   *Decoder
	chunk Chunk
	err   error
}

// NewChunkIterator creates a new iterator for reading chunks.
func NewChunkIterator(r io.Reader) *ChunkIterator {
	return &ChunkIterator{
		dec: NewDecoder(r),
	}
}

// Next advances to the next chunk. Returns false when done or on error.
func (ci *ChunkIterator) Next() bool {
	if ci.err != nil {
		return false
	}

	chunk, err := ci.dec.Decode()
	if err != nil {
		ci.err = err
		ci.chunk = nil
		return false
	}

	ci.chunk = chunk
	return true
}

// Chunk returns the current chunk.
func (ci *ChunkIterator) Chunk() Chunk {
	return ci.chunk
}

// Err returns any error that occurred during iteration.
// ErrStreamDone is not considered an error and will return nil.
func (ci *ChunkIterator) Err() error {
	if ci.err == ErrStreamDone {
		return nil
	}
	return ci.err
}

// DecodeFromString is a convenience function for decoding chunks from a string.
func DecodeFromString(s string) ([]Chunk, error) {
	decoder := NewDecoder(strings.NewReader(s))
	return decoder.DecodeAll()
}
