package uimessagestream

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ProviderMetadata holds provider-specific metadata.
// Example: {"anthropic": {"cacheControl": "ephemeral"}}
type ProviderMetadata map[string]map[string]any

// Chunk is the interface implemented by all chunk types.
type Chunk interface {
	// ChunkType returns the type identifier for this chunk.
	ChunkType() string
}

// ChunkType constants for all supported chunk types.
const (
	ChunkTypeStart            = "start"
	ChunkTypeFinish           = "finish"
	ChunkTypeAbort            = "abort"
	ChunkTypeMessageMetadata  = "message-metadata"
	ChunkTypeTextStart        = "text-start"
	ChunkTypeTextDelta        = "text-delta"
	ChunkTypeTextEnd          = "text-end"
	ChunkTypeReasoningStart   = "reasoning-start"
	ChunkTypeReasoningDelta   = "reasoning-delta"
	ChunkTypeReasoningEnd     = "reasoning-end"
	ChunkTypeToolInputStart   = "tool-input-start"
	ChunkTypeToolInputDelta   = "tool-input-delta"
	ChunkTypeToolInputAvail   = "tool-input-available"
	ChunkTypeToolInputError   = "tool-input-error"
	ChunkTypeToolApprovalReq  = "tool-approval-request"
	ChunkTypeToolOutputAvail  = "tool-output-available"
	ChunkTypeToolOutputError  = "tool-output-error"
	ChunkTypeToolOutputDenied = "tool-output-denied"
	ChunkTypeSourceURL        = "source-url"
	ChunkTypeSourceDocument   = "source-document"
	ChunkTypeFile             = "file"
	ChunkTypeStartStep        = "start-step"
	ChunkTypeFinishStep       = "finish-step"
	ChunkTypeError            = "error"
)

// DataChunkTypePrefix is the prefix for custom data chunk types.
const DataChunkTypePrefix = "data-"

// --- Flow Control Chunks ---

// StartChunk initializes the stream and optionally sets message ID and metadata.
type StartChunk struct {
	Type            string `json:"type"` // Always "start"
	MessageID       string `json:"messageId,omitempty"`
	MessageMetadata any    `json:"messageMetadata,omitempty"`
}

func (c *StartChunk) ChunkType() string { return ChunkTypeStart }

// FinishChunk signals stream completion with finish reason.
type FinishChunk struct {
	Type            string       `json:"type"` // Always "finish"
	FinishReason    FinishReason `json:"finishReason,omitempty"`
	MessageMetadata any          `json:"messageMetadata,omitempty"`
}

func (c *FinishChunk) ChunkType() string { return ChunkTypeFinish }

// AbortChunk signals stream was aborted.
type AbortChunk struct {
	Type string `json:"type"` // Always "abort"
}

func (c *AbortChunk) ChunkType() string { return ChunkTypeAbort }

// MessageMetadataChunk updates message metadata during streaming.
type MessageMetadataChunk struct {
	Type            string `json:"type"` // Always "message-metadata"
	MessageMetadata any    `json:"messageMetadata"`
}

func (c *MessageMetadataChunk) ChunkType() string { return ChunkTypeMessageMetadata }

// --- Text Streaming Chunks ---

// TextStartChunk begins a new text part.
type TextStartChunk struct {
	Type             string           `json:"type"` // Always "text-start"
	ID               string           `json:"id"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata,omitempty"`
}

func (c *TextStartChunk) ChunkType() string { return ChunkTypeTextStart }

// TextDeltaChunk appends text content to an active text part.
type TextDeltaChunk struct {
	Type             string           `json:"type"` // Always "text-delta"
	ID               string           `json:"id"`
	Delta            string           `json:"delta"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata,omitempty"`
}

func (c *TextDeltaChunk) ChunkType() string { return ChunkTypeTextDelta }

// TextEndChunk finalizes a text part.
type TextEndChunk struct {
	Type             string           `json:"type"` // Always "text-end"
	ID               string           `json:"id"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata,omitempty"`
}

func (c *TextEndChunk) ChunkType() string { return ChunkTypeTextEnd }

// --- Reasoning Chunks ---

// ReasoningStartChunk begins a reasoning block.
type ReasoningStartChunk struct {
	Type             string           `json:"type"` // Always "reasoning-start"
	ID               string           `json:"id"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata,omitempty"`
}

func (c *ReasoningStartChunk) ChunkType() string { return ChunkTypeReasoningStart }

// ReasoningDeltaChunk appends reasoning content.
type ReasoningDeltaChunk struct {
	Type             string           `json:"type"` // Always "reasoning-delta"
	ID               string           `json:"id"`
	Delta            string           `json:"delta"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata,omitempty"`
}

func (c *ReasoningDeltaChunk) ChunkType() string { return ChunkTypeReasoningDelta }

// ReasoningEndChunk finalizes a reasoning block.
type ReasoningEndChunk struct {
	Type             string           `json:"type"` // Always "reasoning-end"
	ID               string           `json:"id"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata,omitempty"`
}

func (c *ReasoningEndChunk) ChunkType() string { return ChunkTypeReasoningEnd }

// --- Tool Invocation Chunks ---

// ToolInputStartChunk begins a tool invocation.
type ToolInputStartChunk struct {
	Type             string `json:"type"` // Always "tool-input-start"
	ToolCallID       string `json:"toolCallId"`
	ToolName         string `json:"toolName"`
	ProviderExecuted *bool  `json:"providerExecuted,omitempty"`
	Dynamic          *bool  `json:"dynamic,omitempty"`
	Title            string `json:"title,omitempty"`
}

func (c *ToolInputStartChunk) ChunkType() string { return ChunkTypeToolInputStart }

// ToolInputDeltaChunk streams partial JSON input arguments.
type ToolInputDeltaChunk struct {
	Type           string `json:"type"` // Always "tool-input-delta"
	ToolCallID     string `json:"toolCallId"`
	InputTextDelta string `json:"inputTextDelta"`
}

func (c *ToolInputDeltaChunk) ChunkType() string { return ChunkTypeToolInputDelta }

// ToolInputAvailableChunk signals complete, validated tool input is ready.
type ToolInputAvailableChunk struct {
	Type             string           `json:"type"` // Always "tool-input-available"
	ToolCallID       string           `json:"toolCallId"`
	ToolName         string           `json:"toolName"`
	Input            any              `json:"input"`
	ProviderExecuted *bool            `json:"providerExecuted,omitempty"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata,omitempty"`
	Dynamic          *bool            `json:"dynamic,omitempty"`
	Title            string           `json:"title,omitempty"`
}

func (c *ToolInputAvailableChunk) ChunkType() string { return ChunkTypeToolInputAvail }

// ToolInputErrorChunk signals an error occurred during input parsing.
type ToolInputErrorChunk struct {
	Type             string           `json:"type"` // Always "tool-input-error"
	ToolCallID       string           `json:"toolCallId"`
	ToolName         string           `json:"toolName"`
	Input            any              `json:"input"`
	ErrorText        string           `json:"errorText"`
	ProviderExecuted *bool            `json:"providerExecuted,omitempty"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata,omitempty"`
	Dynamic          *bool            `json:"dynamic,omitempty"`
	Title            string           `json:"title,omitempty"`
}

func (c *ToolInputErrorChunk) ChunkType() string { return ChunkTypeToolInputError }

// ToolApprovalRequestChunk requests user approval before tool execution.
type ToolApprovalRequestChunk struct {
	Type       string `json:"type"` // Always "tool-approval-request"
	ApprovalID string `json:"approvalId"`
	ToolCallID string `json:"toolCallId"`
}

func (c *ToolApprovalRequestChunk) ChunkType() string { return ChunkTypeToolApprovalReq }

// ToolOutputAvailableChunk provides tool execution result.
type ToolOutputAvailableChunk struct {
	Type             string `json:"type"` // Always "tool-output-available"
	ToolCallID       string `json:"toolCallId"`
	Output           any    `json:"output"`
	ProviderExecuted *bool  `json:"providerExecuted,omitempty"`
	Dynamic          *bool  `json:"dynamic,omitempty"`
	Preliminary      *bool  `json:"preliminary,omitempty"`
}

func (c *ToolOutputAvailableChunk) ChunkType() string { return ChunkTypeToolOutputAvail }

// ToolOutputErrorChunk signals tool execution failed.
type ToolOutputErrorChunk struct {
	Type             string `json:"type"` // Always "tool-output-error"
	ToolCallID       string `json:"toolCallId"`
	ErrorText        string `json:"errorText"`
	ProviderExecuted *bool  `json:"providerExecuted,omitempty"`
	Dynamic          *bool  `json:"dynamic,omitempty"`
}

func (c *ToolOutputErrorChunk) ChunkType() string { return ChunkTypeToolOutputError }

// ToolOutputDeniedChunk signals user denied tool execution approval.
type ToolOutputDeniedChunk struct {
	Type       string `json:"type"` // Always "tool-output-denied"
	ToolCallID string `json:"toolCallId"`
}

func (c *ToolOutputDeniedChunk) ChunkType() string { return ChunkTypeToolOutputDenied }

// --- Source Chunks ---

// SourceURLChunk represents a web URL reference.
type SourceURLChunk struct {
	Type             string           `json:"type"` // Always "source-url"
	SourceID         string           `json:"sourceId"`
	URL              string           `json:"url"`
	Title            string           `json:"title,omitempty"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata,omitempty"`
}

func (c *SourceURLChunk) ChunkType() string { return ChunkTypeSourceURL }

// SourceDocumentChunk represents a document reference.
type SourceDocumentChunk struct {
	Type             string           `json:"type"` // Always "source-document"
	SourceID         string           `json:"sourceId"`
	MediaType        string           `json:"mediaType"`
	Title            string           `json:"title"`
	Filename         string           `json:"filename,omitempty"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata,omitempty"`
}

func (c *SourceDocumentChunk) ChunkType() string { return ChunkTypeSourceDocument }

// --- File Chunk ---

// FileChunk represents a file attachment.
type FileChunk struct {
	Type             string           `json:"type"` // Always "file"
	URL              string           `json:"url"`
	MediaType        string           `json:"mediaType"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata,omitempty"`
}

func (c *FileChunk) ChunkType() string { return ChunkTypeFile }

// --- Step Boundary Chunks ---

// StartStepChunk marks the beginning of a new reasoning step.
type StartStepChunk struct {
	Type string `json:"type"` // Always "start-step"
}

func (c *StartStepChunk) ChunkType() string { return ChunkTypeStartStep }

// FinishStepChunk marks the end of a reasoning step.
type FinishStepChunk struct {
	Type string `json:"type"` // Always "finish-step"
}

func (c *FinishStepChunk) ChunkType() string { return ChunkTypeFinishStep }

// --- Custom Data Chunk ---

// DataChunk represents custom data parts with dynamic type names.
// The Type field should be "data-{name}" (e.g., "data-weather", "data-chart").
type DataChunk struct {
	Type      string `json:"type"` // "data-{name}"
	ID        string `json:"id,omitempty"`
	Data      any    `json:"data"`
	Transient *bool  `json:"transient,omitempty"`
}

func (c *DataChunk) ChunkType() string { return c.Type }

// DataName returns the name portion of the data chunk type (after "data-").
func (c *DataChunk) DataName() string {
	return strings.TrimPrefix(c.Type, DataChunkTypePrefix)
}

// --- Error Chunk ---

// ErrorChunk represents an error notification.
type ErrorChunk struct {
	Type      string `json:"type"` // Always "error"
	ErrorText string `json:"errorText"`
}

func (c *ErrorChunk) ChunkType() string { return ChunkTypeError }

// --- Chunk Parsing ---

// rawChunk is used for initial JSON parsing to extract the type field.
type rawChunk struct {
	Type string `json:"type"`
}

// ParseChunk parses a JSON-encoded chunk into the appropriate typed struct.
func ParseChunk(data []byte) (Chunk, error) {
	var raw rawChunk
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse chunk type: %w", err)
	}

	var chunk Chunk

	switch raw.Type {
	case ChunkTypeStart:
		chunk = &StartChunk{}
	case ChunkTypeFinish:
		chunk = &FinishChunk{}
	case ChunkTypeAbort:
		chunk = &AbortChunk{}
	case ChunkTypeMessageMetadata:
		chunk = &MessageMetadataChunk{}
	case ChunkTypeTextStart:
		chunk = &TextStartChunk{}
	case ChunkTypeTextDelta:
		chunk = &TextDeltaChunk{}
	case ChunkTypeTextEnd:
		chunk = &TextEndChunk{}
	case ChunkTypeReasoningStart:
		chunk = &ReasoningStartChunk{}
	case ChunkTypeReasoningDelta:
		chunk = &ReasoningDeltaChunk{}
	case ChunkTypeReasoningEnd:
		chunk = &ReasoningEndChunk{}
	case ChunkTypeToolInputStart:
		chunk = &ToolInputStartChunk{}
	case ChunkTypeToolInputDelta:
		chunk = &ToolInputDeltaChunk{}
	case ChunkTypeToolInputAvail:
		chunk = &ToolInputAvailableChunk{}
	case ChunkTypeToolInputError:
		chunk = &ToolInputErrorChunk{}
	case ChunkTypeToolApprovalReq:
		chunk = &ToolApprovalRequestChunk{}
	case ChunkTypeToolOutputAvail:
		chunk = &ToolOutputAvailableChunk{}
	case ChunkTypeToolOutputError:
		chunk = &ToolOutputErrorChunk{}
	case ChunkTypeToolOutputDenied:
		chunk = &ToolOutputDeniedChunk{}
	case ChunkTypeSourceURL:
		chunk = &SourceURLChunk{}
	case ChunkTypeSourceDocument:
		chunk = &SourceDocumentChunk{}
	case ChunkTypeFile:
		chunk = &FileChunk{}
	case ChunkTypeStartStep:
		chunk = &StartStepChunk{}
	case ChunkTypeFinishStep:
		chunk = &FinishStepChunk{}
	case ChunkTypeError:
		chunk = &ErrorChunk{}
	default:
		// Check for data-* chunks
		if strings.HasPrefix(raw.Type, DataChunkTypePrefix) {
			chunk = &DataChunk{}
		} else {
			return nil, fmt.Errorf("unknown chunk type: %s", raw.Type)
		}
	}

	if err := json.Unmarshal(data, chunk); err != nil {
		return nil, fmt.Errorf("failed to parse %s chunk: %w", raw.Type, err)
	}

	return chunk, nil
}

// MarshalChunk marshals a chunk to JSON, ensuring the type field is set correctly.
func MarshalChunk(chunk Chunk) ([]byte, error) {
	// Set the type field based on the chunk type
	switch c := chunk.(type) {
	case *StartChunk:
		c.Type = ChunkTypeStart
	case *FinishChunk:
		c.Type = ChunkTypeFinish
	case *AbortChunk:
		c.Type = ChunkTypeAbort
	case *MessageMetadataChunk:
		c.Type = ChunkTypeMessageMetadata
	case *TextStartChunk:
		c.Type = ChunkTypeTextStart
	case *TextDeltaChunk:
		c.Type = ChunkTypeTextDelta
	case *TextEndChunk:
		c.Type = ChunkTypeTextEnd
	case *ReasoningStartChunk:
		c.Type = ChunkTypeReasoningStart
	case *ReasoningDeltaChunk:
		c.Type = ChunkTypeReasoningDelta
	case *ReasoningEndChunk:
		c.Type = ChunkTypeReasoningEnd
	case *ToolInputStartChunk:
		c.Type = ChunkTypeToolInputStart
	case *ToolInputDeltaChunk:
		c.Type = ChunkTypeToolInputDelta
	case *ToolInputAvailableChunk:
		c.Type = ChunkTypeToolInputAvail
	case *ToolInputErrorChunk:
		c.Type = ChunkTypeToolInputError
	case *ToolApprovalRequestChunk:
		c.Type = ChunkTypeToolApprovalReq
	case *ToolOutputAvailableChunk:
		c.Type = ChunkTypeToolOutputAvail
	case *ToolOutputErrorChunk:
		c.Type = ChunkTypeToolOutputError
	case *ToolOutputDeniedChunk:
		c.Type = ChunkTypeToolOutputDenied
	case *SourceURLChunk:
		c.Type = ChunkTypeSourceURL
	case *SourceDocumentChunk:
		c.Type = ChunkTypeSourceDocument
	case *FileChunk:
		c.Type = ChunkTypeFile
	case *StartStepChunk:
		c.Type = ChunkTypeStartStep
	case *FinishStepChunk:
		c.Type = ChunkTypeFinishStep
	case *ErrorChunk:
		c.Type = ChunkTypeError
	case *DataChunk:
		// DataChunk already has its type set
		if c.Type == "" || !strings.HasPrefix(c.Type, DataChunkTypePrefix) {
			return nil, fmt.Errorf("DataChunk must have a type starting with %q", DataChunkTypePrefix)
		}
	}

	return json.Marshal(chunk)
}
