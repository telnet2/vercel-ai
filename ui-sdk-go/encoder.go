package uimessagestream

import (
	"fmt"
	"io"
	"net/http"
)

// SSE format constants.
const (
	// ProtocolHeader is the header key for the protocol version.
	ProtocolHeader = "X-Vercel-AI-UI-Message-Stream"
	// ProtocolVersion is the current protocol version.
	ProtocolVersion = "v1"
	// DoneMarker is the marker that signals the end of the stream.
	DoneMarker = "[DONE]"
)

// Encoder writes UI message stream chunks in SSE format.
type Encoder struct {
	w       io.Writer
	flusher http.Flusher
}

// NewEncoder creates a new Encoder that writes to w.
// If w implements http.Flusher, the encoder will flush after each write.
func NewEncoder(w io.Writer) *Encoder {
	flusher, _ := w.(http.Flusher)
	return &Encoder{
		w:       w,
		flusher: flusher,
	}
}

// WriteHeaders writes the required SSE headers to an http.ResponseWriter.
// This should be called before any Encode calls.
func WriteHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set(ProtocolHeader, ProtocolVersion)
	w.Header().Set("X-Accel-Buffering", "no")
}

// Encode writes a chunk in SSE format: "data: {json}\n\n"
func (e *Encoder) Encode(chunk Chunk) error {
	data, err := MarshalChunk(chunk)
	if err != nil {
		return fmt.Errorf("failed to marshal chunk: %w", err)
	}

	if _, err := fmt.Fprintf(e.w, "data: %s\n\n", data); err != nil {
		return fmt.Errorf("failed to write chunk: %w", err)
	}

	if e.flusher != nil {
		e.flusher.Flush()
	}

	return nil
}

// EncodeRaw writes raw bytes as an SSE data event.
// The data should be a valid JSON-encoded chunk.
func (e *Encoder) EncodeRaw(data []byte) error {
	if _, err := fmt.Fprintf(e.w, "data: %s\n\n", data); err != nil {
		return fmt.Errorf("failed to write chunk: %w", err)
	}

	if e.flusher != nil {
		e.flusher.Flush()
	}

	return nil
}

// Done writes the stream termination marker "[DONE]".
func (e *Encoder) Done() error {
	if _, err := fmt.Fprintf(e.w, "data: %s\n\n", DoneMarker); err != nil {
		return fmt.Errorf("failed to write done marker: %w", err)
	}

	if e.flusher != nil {
		e.flusher.Flush()
	}

	return nil
}

// Flush flushes any buffered data to the underlying writer.
func (e *Encoder) Flush() {
	if e.flusher != nil {
		e.flusher.Flush()
	}
}

// StreamWriter provides a convenient interface for writing a complete stream.
type StreamWriter struct {
	enc     *Encoder
	started bool
	err     error
}

// NewStreamWriter creates a new StreamWriter that writes to w.
// It automatically writes SSE headers if w is an http.ResponseWriter.
func NewStreamWriter(w http.ResponseWriter) *StreamWriter {
	WriteHeaders(w)
	return &StreamWriter{
		enc: NewEncoder(w),
	}
}

// Start writes the start chunk. Should be called first.
func (sw *StreamWriter) Start(messageID string, metadata any) error {
	if sw.err != nil {
		return sw.err
	}
	sw.started = true

	chunk := &StartChunk{
		MessageID:       messageID,
		MessageMetadata: metadata,
	}
	sw.err = sw.enc.Encode(chunk)
	return sw.err
}

// WriteChunk writes a chunk to the stream.
func (sw *StreamWriter) WriteChunk(chunk Chunk) error {
	if sw.err != nil {
		return sw.err
	}
	sw.err = sw.enc.Encode(chunk)
	return sw.err
}

// WriteText writes a complete text part (start, delta, end).
func (sw *StreamWriter) WriteText(id string, text string) error {
	if sw.err != nil {
		return sw.err
	}

	// Start
	if err := sw.enc.Encode(&TextStartChunk{ID: id}); err != nil {
		sw.err = err
		return err
	}

	// Delta
	if err := sw.enc.Encode(&TextDeltaChunk{ID: id, Delta: text}); err != nil {
		sw.err = err
		return err
	}

	// End
	if err := sw.enc.Encode(&TextEndChunk{ID: id}); err != nil {
		sw.err = err
		return err
	}

	return nil
}

// WriteTextDelta writes a text delta chunk.
func (sw *StreamWriter) WriteTextDelta(id string, delta string) error {
	if sw.err != nil {
		return sw.err
	}
	sw.err = sw.enc.Encode(&TextDeltaChunk{ID: id, Delta: delta})
	return sw.err
}

// Finish writes the finish chunk and done marker.
func (sw *StreamWriter) Finish(reason FinishReason, metadata any) error {
	if sw.err != nil {
		return sw.err
	}

	chunk := &FinishChunk{
		FinishReason:    reason,
		MessageMetadata: metadata,
	}
	if err := sw.enc.Encode(chunk); err != nil {
		sw.err = err
		return err
	}

	sw.err = sw.enc.Done()
	return sw.err
}

// Abort writes the abort chunk and done marker.
func (sw *StreamWriter) Abort() error {
	if sw.err != nil {
		return sw.err
	}

	if err := sw.enc.Encode(&AbortChunk{}); err != nil {
		sw.err = err
		return err
	}

	sw.err = sw.enc.Done()
	return sw.err
}

// Error writes an error chunk, finish chunk, and done marker.
func (sw *StreamWriter) Error(errorText string) error {
	if sw.err != nil {
		return sw.err
	}

	if err := sw.enc.Encode(&ErrorChunk{ErrorText: errorText}); err != nil {
		sw.err = err
		return err
	}

	return sw.Finish(FinishReasonError, nil)
}

// Err returns any error that occurred during writing.
func (sw *StreamWriter) Err() error {
	return sw.err
}
