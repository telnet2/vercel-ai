package uimessagestream

// Role represents the role of a message participant.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// UIMessage represents a message in the UI.
type UIMessage struct {
	ID       string          `json:"id"`
	Role     Role            `json:"role"`
	Metadata any             `json:"metadata,omitempty"`
	Parts    []UIMessagePart `json:"parts"`
}

// UIMessagePart is the interface implemented by all message part types.
type UIMessagePart interface {
	// PartType returns the type identifier for this part.
	PartType() string
}

// PartState represents the streaming state of a part.
type PartState string

const (
	PartStateStreaming PartState = "streaming"
	PartStateDone      PartState = "done"
)

// ToolState represents the state of a tool invocation.
type ToolState string

const (
	ToolStateInputStreaming    ToolState = "input-streaming"
	ToolStateInputAvailable    ToolState = "input-available"
	ToolStateApprovalRequested ToolState = "approval-requested"
	ToolStateApprovalResponded ToolState = "approval-responded"
	ToolStateOutputAvailable   ToolState = "output-available"
	ToolStateOutputError       ToolState = "output-error"
	ToolStateOutputDenied      ToolState = "output-denied"
)

// ToolApproval represents approval state for a tool invocation.
type ToolApproval struct {
	ID       string `json:"id"`
	Approved *bool  `json:"approved,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

// --- UI Part Types ---

// TextUIPart represents a text part in the message.
type TextUIPart struct {
	Type             string           `json:"type"` // Always "text"
	Text             string           `json:"text"`
	State            PartState        `json:"state,omitempty"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata,omitempty"`
}

func (p *TextUIPart) PartType() string { return "text" }

// ReasoningUIPart represents a reasoning/thinking part in the message.
type ReasoningUIPart struct {
	Type             string           `json:"type"` // Always "reasoning"
	Text             string           `json:"text"`
	State            PartState        `json:"state,omitempty"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata,omitempty"`
}

func (p *ReasoningUIPart) PartType() string { return "reasoning" }

// ToolUIPart represents a static tool invocation in the message.
// The Type field is "tool-{toolName}".
type ToolUIPart struct {
	Type                 string           `json:"type"` // "tool-{toolName}"
	ToolCallID           string           `json:"toolCallId"`
	Title                string           `json:"title,omitempty"`
	ProviderExecuted     *bool            `json:"providerExecuted,omitempty"`
	State                ToolState        `json:"state"`
	Input                any              `json:"input,omitempty"`
	Output               any              `json:"output,omitempty"`
	ErrorText            string           `json:"errorText,omitempty"`
	Preliminary          *bool            `json:"preliminary,omitempty"`
	CallProviderMetadata ProviderMetadata `json:"callProviderMetadata,omitempty"`
	Approval             *ToolApproval    `json:"approval,omitempty"`
}

func (p *ToolUIPart) PartType() string { return p.Type }

// DynamicToolUIPart represents a dynamic (runtime-registered) tool invocation.
type DynamicToolUIPart struct {
	Type                 string           `json:"type"` // Always "dynamic-tool"
	ToolName             string           `json:"toolName"`
	ToolCallID           string           `json:"toolCallId"`
	Title                string           `json:"title,omitempty"`
	ProviderExecuted     *bool            `json:"providerExecuted,omitempty"`
	State                ToolState        `json:"state"`
	Input                any              `json:"input,omitempty"`
	Output               any              `json:"output,omitempty"`
	ErrorText            string           `json:"errorText,omitempty"`
	Preliminary          *bool            `json:"preliminary,omitempty"`
	CallProviderMetadata ProviderMetadata `json:"callProviderMetadata,omitempty"`
	Approval             *ToolApproval    `json:"approval,omitempty"`
}

func (p *DynamicToolUIPart) PartType() string { return "dynamic-tool" }

// SourceURLUIPart represents a web URL source reference.
type SourceURLUIPart struct {
	Type             string           `json:"type"` // Always "source-url"
	SourceID         string           `json:"sourceId"`
	URL              string           `json:"url"`
	Title            string           `json:"title,omitempty"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata,omitempty"`
}

func (p *SourceURLUIPart) PartType() string { return "source-url" }

// SourceDocumentUIPart represents a document source reference.
type SourceDocumentUIPart struct {
	Type             string           `json:"type"` // Always "source-document"
	SourceID         string           `json:"sourceId"`
	MediaType        string           `json:"mediaType"`
	Title            string           `json:"title"`
	Filename         string           `json:"filename,omitempty"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata,omitempty"`
}

func (p *SourceDocumentUIPart) PartType() string { return "source-document" }

// FileUIPart represents a file attachment.
type FileUIPart struct {
	Type             string           `json:"type"` // Always "file"
	MediaType        string           `json:"mediaType"`
	Filename         string           `json:"filename,omitempty"`
	URL              string           `json:"url"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata,omitempty"`
}

func (p *FileUIPart) PartType() string { return "file" }

// DataUIPart represents custom data in the message.
// The Type field is "data-{name}".
type DataUIPart struct {
	Type string `json:"type"` // "data-{name}"
	ID   string `json:"id,omitempty"`
	Data any    `json:"data"`
}

func (p *DataUIPart) PartType() string { return p.Type }

// StepStartUIPart marks the beginning of a reasoning step.
type StepStartUIPart struct {
	Type string `json:"type"` // Always "step-start"
}

func (p *StepStartUIPart) PartType() string { return "step-start" }
