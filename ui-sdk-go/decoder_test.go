package uimessagestream

import (
	"io"
	"strings"
	"testing"
)

func TestDecoder_Decode(t *testing.T) {
	input := `data: {"type":"text-delta","id":"text-1","delta":"Hello"}

data: {"type":"text-delta","id":"text-1","delta":" world"}

data: [DONE]

`

	dec := NewDecoder(strings.NewReader(input))

	// First chunk
	chunk1, err := dec.Decode()
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if chunk1.ChunkType() != ChunkTypeTextDelta {
		t.Errorf("ChunkType() = %v, want %v", chunk1.ChunkType(), ChunkTypeTextDelta)
	}
	delta1 := chunk1.(*TextDeltaChunk)
	if delta1.Delta != "Hello" {
		t.Errorf("Delta = %v, want Hello", delta1.Delta)
	}

	// Second chunk
	chunk2, err := dec.Decode()
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	delta2 := chunk2.(*TextDeltaChunk)
	if delta2.Delta != " world" {
		t.Errorf("Delta = %v, want ' world'", delta2.Delta)
	}

	// Done
	_, err = dec.Decode()
	if err != ErrStreamDone {
		t.Errorf("Decode() error = %v, want ErrStreamDone", err)
	}

	if !dec.IsDone() {
		t.Error("IsDone() should return true")
	}
}

func TestDecoder_DecodeAll(t *testing.T) {
	input := `data: {"type":"start","messageId":"msg-123"}

data: {"type":"text-start","id":"text-1"}

data: {"type":"text-delta","id":"text-1","delta":"Hello"}

data: {"type":"text-end","id":"text-1"}

data: {"type":"finish","finishReason":"stop"}

data: [DONE]

`

	dec := NewDecoder(strings.NewReader(input))
	chunks, err := dec.DecodeAll()
	if err != nil {
		t.Fatalf("DecodeAll() error = %v", err)
	}

	if len(chunks) != 5 {
		t.Fatalf("DecodeAll() returned %d chunks, want 5", len(chunks))
	}

	// Verify chunk types
	expectedTypes := []string{
		ChunkTypeStart,
		ChunkTypeTextStart,
		ChunkTypeTextDelta,
		ChunkTypeTextEnd,
		ChunkTypeFinish,
	}

	for i, expected := range expectedTypes {
		if chunks[i].ChunkType() != expected {
			t.Errorf("chunks[%d].ChunkType() = %v, want %v", i, chunks[i].ChunkType(), expected)
		}
	}
}

func TestDecoder_NoData(t *testing.T) {
	input := ``
	dec := NewDecoder(strings.NewReader(input))

	_, err := dec.Decode()
	if err != io.EOF {
		t.Errorf("Decode() error = %v, want io.EOF", err)
	}
}

func TestDecoder_SkipsEmptyLines(t *testing.T) {
	input := `

data: {"type":"text-delta","id":"text-1","delta":"Hello"}


data: [DONE]

`

	dec := NewDecoder(strings.NewReader(input))
	chunk, err := dec.Decode()
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if chunk.ChunkType() != ChunkTypeTextDelta {
		t.Errorf("ChunkType() = %v, want %v", chunk.ChunkType(), ChunkTypeTextDelta)
	}
}

func TestDecoder_SkipsComments(t *testing.T) {
	input := `: this is a comment
data: {"type":"text-delta","id":"text-1","delta":"Hello"}

data: [DONE]

`

	dec := NewDecoder(strings.NewReader(input))
	chunk, err := dec.Decode()
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if chunk.ChunkType() != ChunkTypeTextDelta {
		t.Errorf("ChunkType() = %v, want %v", chunk.ChunkType(), ChunkTypeTextDelta)
	}
}

func TestDecoder_InvalidJSON(t *testing.T) {
	input := `data: {invalid json}

`

	dec := NewDecoder(strings.NewReader(input))
	_, err := dec.Decode()
	if err == nil {
		t.Error("Decode() expected error for invalid JSON")
	}
}

func TestDecodeFromString(t *testing.T) {
	input := `data: {"type":"start","messageId":"msg-123"}

data: {"type":"finish","finishReason":"stop"}

data: [DONE]

`

	chunks, err := DecodeFromString(input)
	if err != nil {
		t.Fatalf("DecodeFromString() error = %v", err)
	}

	if len(chunks) != 2 {
		t.Fatalf("DecodeFromString() returned %d chunks, want 2", len(chunks))
	}
}

func TestStreamReader_Next(t *testing.T) {
	input := `data: {"type":"start","messageId":"msg-123","messageMetadata":{"key":"value"}}

data: {"type":"text-delta","id":"text-1","delta":"Hello"}

data: {"type":"finish","finishReason":"stop","messageMetadata":{"tokens":10}}

data: [DONE]

`

	sr := NewStreamReader(strings.NewReader(input))

	// Start chunk
	chunk1, err := sr.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if chunk1.ChunkType() != ChunkTypeStart {
		t.Errorf("ChunkType() = %v, want %v", chunk1.ChunkType(), ChunkTypeStart)
	}
	if sr.MessageID() != "msg-123" {
		t.Errorf("MessageID() = %v, want msg-123", sr.MessageID())
	}

	// Text delta
	chunk2, err := sr.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if chunk2.ChunkType() != ChunkTypeTextDelta {
		t.Errorf("ChunkType() = %v, want %v", chunk2.ChunkType(), ChunkTypeTextDelta)
	}

	// Finish chunk - should update metadata
	chunk3, err := sr.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if chunk3.ChunkType() != ChunkTypeFinish {
		t.Errorf("ChunkType() = %v, want %v", chunk3.ChunkType(), ChunkTypeFinish)
	}

	// Metadata should be updated from finish chunk
	meta, ok := sr.Metadata().(map[string]any)
	if !ok {
		t.Fatalf("Metadata() type = %T, want map[string]any", sr.Metadata())
	}
	if meta["tokens"] != float64(10) {
		t.Errorf("Metadata()[tokens] = %v, want 10", meta["tokens"])
	}

	// Done
	_, err = sr.Next()
	if err != ErrStreamDone {
		t.Errorf("Next() error = %v, want ErrStreamDone", err)
	}
}

func TestChunkIterator(t *testing.T) {
	input := `data: {"type":"start","messageId":"msg-123"}

data: {"type":"text-delta","id":"text-1","delta":"Hello"}

data: {"type":"finish","finishReason":"stop"}

data: [DONE]

`

	iter := NewChunkIterator(strings.NewReader(input))

	var chunks []Chunk
	for iter.Next() {
		chunks = append(chunks, iter.Chunk())
	}

	if err := iter.Err(); err != nil {
		t.Fatalf("Err() = %v", err)
	}

	if len(chunks) != 3 {
		t.Fatalf("got %d chunks, want 3", len(chunks))
	}
}

func TestDecoder_ToolCallSequence(t *testing.T) {
	input := `data: {"type":"start","messageId":"msg-456"}

data: {"type":"tool-input-start","toolCallId":"call-1","toolName":"getWeather"}

data: {"type":"tool-input-delta","toolCallId":"call-1","inputTextDelta":"{\"location\":"}

data: {"type":"tool-input-delta","toolCallId":"call-1","inputTextDelta":"\"San Francisco\"}"}

data: {"type":"tool-input-available","toolCallId":"call-1","toolName":"getWeather","input":{"location":"San Francisco"}}

data: {"type":"tool-output-available","toolCallId":"call-1","output":{"temperature":18,"condition":"sunny"}}

data: {"type":"finish","finishReason":"stop"}

data: [DONE]

`

	chunks, err := DecodeFromString(input)
	if err != nil {
		t.Fatalf("DecodeFromString() error = %v", err)
	}

	if len(chunks) != 7 {
		t.Fatalf("got %d chunks, want 7", len(chunks))
	}

	// Verify tool input start
	toolStart, ok := chunks[1].(*ToolInputStartChunk)
	if !ok {
		t.Fatalf("chunks[1] type = %T, want *ToolInputStartChunk", chunks[1])
	}
	if toolStart.ToolName != "getWeather" {
		t.Errorf("ToolName = %v, want getWeather", toolStart.ToolName)
	}

	// Verify tool input available
	toolAvail, ok := chunks[4].(*ToolInputAvailableChunk)
	if !ok {
		t.Fatalf("chunks[4] type = %T, want *ToolInputAvailableChunk", chunks[4])
	}
	input_map, ok := toolAvail.Input.(map[string]any)
	if !ok {
		t.Fatalf("Input type = %T, want map[string]any", toolAvail.Input)
	}
	if input_map["location"] != "San Francisco" {
		t.Errorf("Input[location] = %v, want San Francisco", input_map["location"])
	}

	// Verify tool output
	toolOutput, ok := chunks[5].(*ToolOutputAvailableChunk)
	if !ok {
		t.Fatalf("chunks[5] type = %T, want *ToolOutputAvailableChunk", chunks[5])
	}
	output_map, ok := toolOutput.Output.(map[string]any)
	if !ok {
		t.Fatalf("Output type = %T, want map[string]any", toolOutput.Output)
	}
	if output_map["temperature"] != float64(18) {
		t.Errorf("Output[temperature] = %v, want 18", output_map["temperature"])
	}
}

func TestDecoder_DataChunk(t *testing.T) {
	input := `data: {"type":"data-weather","id":"w-1","data":{"temp":72,"condition":"sunny"},"transient":true}

data: [DONE]

`

	chunks, err := DecodeFromString(input)
	if err != nil {
		t.Fatalf("DecodeFromString() error = %v", err)
	}

	if len(chunks) != 1 {
		t.Fatalf("got %d chunks, want 1", len(chunks))
	}

	dataChunk, ok := chunks[0].(*DataChunk)
	if !ok {
		t.Fatalf("chunks[0] type = %T, want *DataChunk", chunks[0])
	}

	if dataChunk.Type != "data-weather" {
		t.Errorf("Type = %v, want data-weather", dataChunk.Type)
	}
	if dataChunk.DataName() != "weather" {
		t.Errorf("DataName() = %v, want weather", dataChunk.DataName())
	}
	if dataChunk.ID != "w-1" {
		t.Errorf("ID = %v, want w-1", dataChunk.ID)
	}
	if dataChunk.Transient == nil || *dataChunk.Transient != true {
		t.Error("Transient should be true")
	}
}
