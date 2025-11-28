package uimessagestream

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEncoder_Encode(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	chunk := &TextDeltaChunk{ID: "text-1", Delta: "Hello"}
	if err := enc.Encode(chunk); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	output := buf.String()
	if !strings.HasPrefix(output, "data: ") {
		t.Error("output should start with 'data: '")
	}
	if !strings.HasSuffix(output, "\n\n") {
		t.Error("output should end with '\\n\\n'")
	}
	if !strings.Contains(output, `"type":"text-delta"`) {
		t.Error("output should contain type field")
	}
	if !strings.Contains(output, `"delta":"Hello"`) {
		t.Error("output should contain delta field")
	}
}

func TestEncoder_Done(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	if err := enc.Done(); err != nil {
		t.Fatalf("Done() error = %v", err)
	}

	expected := "data: [DONE]\n\n"
	if got := buf.String(); got != expected {
		t.Errorf("Done() = %q, want %q", got, expected)
	}
}

func TestEncoder_EncodeRaw(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	raw := []byte(`{"type":"text-delta","id":"text-1","delta":"raw"}`)
	if err := enc.EncodeRaw(raw); err != nil {
		t.Fatalf("EncodeRaw() error = %v", err)
	}

	output := buf.String()
	expected := "data: {\"type\":\"text-delta\",\"id\":\"text-1\",\"delta\":\"raw\"}\n\n"
	if output != expected {
		t.Errorf("EncodeRaw() = %q, want %q", output, expected)
	}
}

func TestWriteHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteHeaders(rec)

	tests := []struct {
		header string
		want   string
	}{
		{"Content-Type", "text/event-stream"},
		{"Cache-Control", "no-cache"},
		{"Connection", "keep-alive"},
		{ProtocolHeader, ProtocolVersion},
		{"X-Accel-Buffering", "no"},
	}

	for _, tt := range tests {
		if got := rec.Header().Get(tt.header); got != tt.want {
			t.Errorf("Header %s = %q, want %q", tt.header, got, tt.want)
		}
	}
}

func TestStreamWriter_SimpleResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewStreamWriter(rec)

	// Start
	if err := sw.Start("msg-123", nil); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Text
	if err := sw.WriteText("text-1", "Hello, world!"); err != nil {
		t.Fatalf("WriteText() error = %v", err)
	}

	// Finish
	if err := sw.Finish(FinishReasonStop, nil); err != nil {
		t.Fatalf("Finish() error = %v", err)
	}

	// Verify output
	output := rec.Body.String()

	expectedParts := []string{
		`"type":"start"`,
		`"messageId":"msg-123"`,
		`"type":"text-start"`,
		`"type":"text-delta"`,
		`"delta":"Hello, world!"`,
		`"type":"text-end"`,
		`"type":"finish"`,
		`"finishReason":"stop"`,
		"[DONE]",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("output should contain %q", part)
		}
	}
}

func TestStreamWriter_Error(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewStreamWriter(rec)

	if err := sw.Start("msg-123", nil); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if err := sw.Error("Something went wrong"); err != nil {
		t.Fatalf("Error() error = %v", err)
	}

	output := rec.Body.String()
	if !strings.Contains(output, `"type":"error"`) {
		t.Error("output should contain error chunk")
	}
	if !strings.Contains(output, `"errorText":"Something went wrong"`) {
		t.Error("output should contain error text")
	}
	if !strings.Contains(output, `"finishReason":"error"`) {
		t.Error("output should contain error finish reason")
	}
}

func TestStreamWriter_Abort(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewStreamWriter(rec)

	if err := sw.Start("msg-123", nil); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if err := sw.Abort(); err != nil {
		t.Fatalf("Abort() error = %v", err)
	}

	output := rec.Body.String()
	if !strings.Contains(output, `"type":"abort"`) {
		t.Error("output should contain abort chunk")
	}
	if !strings.Contains(output, "[DONE]") {
		t.Error("output should contain DONE marker")
	}
}

func TestStreamWriter_WriteChunk(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewStreamWriter(rec)

	if err := sw.Start("msg-123", nil); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Write a custom data chunk
	chunk := &DataChunk{
		Type: "data-weather",
		ID:   "w-1",
		Data: map[string]any{"temp": 72, "condition": "sunny"},
	}
	if err := sw.WriteChunk(chunk); err != nil {
		t.Fatalf("WriteChunk() error = %v", err)
	}

	if err := sw.Finish(FinishReasonStop, nil); err != nil {
		t.Fatalf("Finish() error = %v", err)
	}

	output := rec.Body.String()
	if !strings.Contains(output, `"type":"data-weather"`) {
		t.Error("output should contain data-weather chunk")
	}
}

func TestStreamWriter_Metadata(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := NewStreamWriter(rec)

	metadata := map[string]string{"model": "gpt-4", "temperature": "0.7"}

	if err := sw.Start("msg-123", metadata); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if err := sw.Finish(FinishReasonStop, map[string]int{"tokens": 100}); err != nil {
		t.Fatalf("Finish() error = %v", err)
	}

	output := rec.Body.String()
	if !strings.Contains(output, `"messageMetadata"`) {
		t.Error("output should contain messageMetadata")
	}
}
