package uimessagestream

import (
	"encoding/json"
	"testing"
)

func TestParseChunk_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string
	}{
		{
			name:     "start",
			input:    `{"type":"start","messageId":"msg-123"}`,
			wantType: ChunkTypeStart,
		},
		{
			name:     "finish",
			input:    `{"type":"finish","finishReason":"stop"}`,
			wantType: ChunkTypeFinish,
		},
		{
			name:     "abort",
			input:    `{"type":"abort"}`,
			wantType: ChunkTypeAbort,
		},
		{
			name:     "message-metadata",
			input:    `{"type":"message-metadata","messageMetadata":{"key":"value"}}`,
			wantType: ChunkTypeMessageMetadata,
		},
		{
			name:     "text-start",
			input:    `{"type":"text-start","id":"text-1"}`,
			wantType: ChunkTypeTextStart,
		},
		{
			name:     "text-delta",
			input:    `{"type":"text-delta","id":"text-1","delta":"Hello"}`,
			wantType: ChunkTypeTextDelta,
		},
		{
			name:     "text-end",
			input:    `{"type":"text-end","id":"text-1"}`,
			wantType: ChunkTypeTextEnd,
		},
		{
			name:     "reasoning-start",
			input:    `{"type":"reasoning-start","id":"reason-1"}`,
			wantType: ChunkTypeReasoningStart,
		},
		{
			name:     "reasoning-delta",
			input:    `{"type":"reasoning-delta","id":"reason-1","delta":"Thinking..."}`,
			wantType: ChunkTypeReasoningDelta,
		},
		{
			name:     "reasoning-end",
			input:    `{"type":"reasoning-end","id":"reason-1"}`,
			wantType: ChunkTypeReasoningEnd,
		},
		{
			name:     "tool-input-start",
			input:    `{"type":"tool-input-start","toolCallId":"call-1","toolName":"getWeather"}`,
			wantType: ChunkTypeToolInputStart,
		},
		{
			name:     "tool-input-delta",
			input:    `{"type":"tool-input-delta","toolCallId":"call-1","inputTextDelta":"{\"location\":"}`,
			wantType: ChunkTypeToolInputDelta,
		},
		{
			name:     "tool-input-available",
			input:    `{"type":"tool-input-available","toolCallId":"call-1","toolName":"getWeather","input":{"location":"NYC"}}`,
			wantType: ChunkTypeToolInputAvail,
		},
		{
			name:     "tool-input-error",
			input:    `{"type":"tool-input-error","toolCallId":"call-1","toolName":"getWeather","input":null,"errorText":"invalid input"}`,
			wantType: ChunkTypeToolInputError,
		},
		{
			name:     "tool-approval-request",
			input:    `{"type":"tool-approval-request","approvalId":"apr-1","toolCallId":"call-1"}`,
			wantType: ChunkTypeToolApprovalReq,
		},
		{
			name:     "tool-output-available",
			input:    `{"type":"tool-output-available","toolCallId":"call-1","output":{"temp":72}}`,
			wantType: ChunkTypeToolOutputAvail,
		},
		{
			name:     "tool-output-error",
			input:    `{"type":"tool-output-error","toolCallId":"call-1","errorText":"API error"}`,
			wantType: ChunkTypeToolOutputError,
		},
		{
			name:     "tool-output-denied",
			input:    `{"type":"tool-output-denied","toolCallId":"call-1"}`,
			wantType: ChunkTypeToolOutputDenied,
		},
		{
			name:     "source-url",
			input:    `{"type":"source-url","sourceId":"src-1","url":"https://example.com","title":"Example"}`,
			wantType: ChunkTypeSourceURL,
		},
		{
			name:     "source-document",
			input:    `{"type":"source-document","sourceId":"src-1","mediaType":"application/pdf","title":"Doc"}`,
			wantType: ChunkTypeSourceDocument,
		},
		{
			name:     "file",
			input:    `{"type":"file","url":"https://example.com/file.png","mediaType":"image/png"}`,
			wantType: ChunkTypeFile,
		},
		{
			name:     "start-step",
			input:    `{"type":"start-step"}`,
			wantType: ChunkTypeStartStep,
		},
		{
			name:     "finish-step",
			input:    `{"type":"finish-step"}`,
			wantType: ChunkTypeFinishStep,
		},
		{
			name:     "error",
			input:    `{"type":"error","errorText":"Something went wrong"}`,
			wantType: ChunkTypeError,
		},
		{
			name:     "data-custom",
			input:    `{"type":"data-weather","id":"w-1","data":{"temp":72,"condition":"sunny"}}`,
			wantType: "data-weather",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk, err := ParseChunk([]byte(tt.input))
			if err != nil {
				t.Fatalf("ParseChunk() error = %v", err)
			}
			if got := chunk.ChunkType(); got != tt.wantType {
				t.Errorf("ChunkType() = %v, want %v", got, tt.wantType)
			}
		})
	}
}

func TestMarshalChunk_RoundTrip(t *testing.T) {
	chunks := []Chunk{
		&StartChunk{MessageID: "msg-123", MessageMetadata: map[string]string{"key": "value"}},
		&FinishChunk{FinishReason: FinishReasonStop},
		&AbortChunk{},
		&TextStartChunk{ID: "text-1"},
		&TextDeltaChunk{ID: "text-1", Delta: "Hello, world!"},
		&TextEndChunk{ID: "text-1"},
		&ReasoningStartChunk{ID: "reason-1"},
		&ReasoningDeltaChunk{ID: "reason-1", Delta: "Let me think..."},
		&ReasoningEndChunk{ID: "reason-1"},
		&ToolInputStartChunk{ToolCallID: "call-1", ToolName: "getWeather"},
		&ToolInputDeltaChunk{ToolCallID: "call-1", InputTextDelta: `{"loc`},
		&ToolInputAvailableChunk{ToolCallID: "call-1", ToolName: "getWeather", Input: map[string]string{"location": "NYC"}},
		&ToolOutputAvailableChunk{ToolCallID: "call-1", Output: map[string]int{"temp": 72}},
		&SourceURLChunk{SourceID: "src-1", URL: "https://example.com", Title: "Example"},
		&FileChunk{URL: "https://example.com/file.png", MediaType: "image/png"},
		&StartStepChunk{},
		&FinishStepChunk{},
		&ErrorChunk{ErrorText: "Something went wrong"},
		&DataChunk{Type: "data-weather", ID: "w-1", Data: map[string]any{"temp": 72}},
	}

	for _, original := range chunks {
		t.Run(original.ChunkType(), func(t *testing.T) {
			// Marshal
			data, err := MarshalChunk(original)
			if err != nil {
				t.Fatalf("MarshalChunk() error = %v", err)
			}

			// Parse back
			parsed, err := ParseChunk(data)
			if err != nil {
				t.Fatalf("ParseChunk() error = %v", err)
			}

			// Verify type matches
			if parsed.ChunkType() != original.ChunkType() {
				t.Errorf("ChunkType() = %v, want %v", parsed.ChunkType(), original.ChunkType())
			}

			// Marshal again and compare JSON
			data2, err := MarshalChunk(parsed)
			if err != nil {
				t.Fatalf("MarshalChunk() round-trip error = %v", err)
			}

			// Compare as JSON objects to ignore key ordering
			var obj1, obj2 map[string]any
			if err := json.Unmarshal(data, &obj1); err != nil {
				t.Fatalf("json.Unmarshal(data) error = %v", err)
			}
			if err := json.Unmarshal(data2, &obj2); err != nil {
				t.Fatalf("json.Unmarshal(data2) error = %v", err)
			}

			// Check type field is present
			if obj1["type"] != obj2["type"] {
				t.Errorf("type mismatch: %v != %v", obj1["type"], obj2["type"])
			}
		})
	}
}

func TestParseChunk_UnknownType(t *testing.T) {
	input := `{"type":"unknown-type","data":"test"}`
	_, err := ParseChunk([]byte(input))
	if err == nil {
		t.Error("ParseChunk() expected error for unknown type")
	}
}

func TestParseChunk_InvalidJSON(t *testing.T) {
	input := `{invalid json}`
	_, err := ParseChunk([]byte(input))
	if err == nil {
		t.Error("ParseChunk() expected error for invalid JSON")
	}
}

func TestDataChunk_DataName(t *testing.T) {
	chunk := &DataChunk{Type: "data-weather"}
	if got := chunk.DataName(); got != "weather" {
		t.Errorf("DataName() = %v, want %v", got, "weather")
	}
}

func TestTextDeltaChunk_Fields(t *testing.T) {
	input := `{"type":"text-delta","id":"text-1","delta":"Hello","providerMetadata":{"anthropic":{"cacheControl":"ephemeral"}}}`
	chunk, err := ParseChunk([]byte(input))
	if err != nil {
		t.Fatalf("ParseChunk() error = %v", err)
	}

	textDelta, ok := chunk.(*TextDeltaChunk)
	if !ok {
		t.Fatalf("expected *TextDeltaChunk, got %T", chunk)
	}

	if textDelta.ID != "text-1" {
		t.Errorf("ID = %v, want text-1", textDelta.ID)
	}
	if textDelta.Delta != "Hello" {
		t.Errorf("Delta = %v, want Hello", textDelta.Delta)
	}
	if textDelta.ProviderMetadata == nil {
		t.Error("ProviderMetadata is nil")
	}
	if textDelta.ProviderMetadata["anthropic"]["cacheControl"] != "ephemeral" {
		t.Errorf("ProviderMetadata[anthropic][cacheControl] = %v, want ephemeral", textDelta.ProviderMetadata["anthropic"]["cacheControl"])
	}
}

func TestFinishChunk_Reasons(t *testing.T) {
	reasons := []FinishReason{
		FinishReasonStop,
		FinishReasonLength,
		FinishReasonContentFilter,
		FinishReasonToolCalls,
		FinishReasonError,
		FinishReasonOther,
		FinishReasonUnknown,
	}

	for _, reason := range reasons {
		t.Run(string(reason), func(t *testing.T) {
			chunk := &FinishChunk{FinishReason: reason}
			data, err := MarshalChunk(chunk)
			if err != nil {
				t.Fatalf("MarshalChunk() error = %v", err)
			}

			parsed, err := ParseChunk(data)
			if err != nil {
				t.Fatalf("ParseChunk() error = %v", err)
			}

			finish, ok := parsed.(*FinishChunk)
			if !ok {
				t.Fatalf("expected *FinishChunk, got %T", parsed)
			}

			if finish.FinishReason != reason {
				t.Errorf("FinishReason = %v, want %v", finish.FinishReason, reason)
			}
		})
	}
}
