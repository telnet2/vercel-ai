package uimessagestream

// FinishReason indicates why the stream finished.
type FinishReason string

const (
	// FinishReasonStop indicates normal completion.
	FinishReasonStop FinishReason = "stop"
	// FinishReasonLength indicates the response was truncated due to length limits.
	FinishReasonLength FinishReason = "length"
	// FinishReasonContentFilter indicates content was filtered.
	FinishReasonContentFilter FinishReason = "content-filter"
	// FinishReasonToolCalls indicates the model wants to call tools.
	FinishReasonToolCalls FinishReason = "tool-calls"
	// FinishReasonError indicates an error occurred.
	FinishReasonError FinishReason = "error"
	// FinishReasonOther indicates another reason.
	FinishReasonOther FinishReason = "other"
	// FinishReasonUnknown indicates an unknown reason.
	FinishReasonUnknown FinishReason = "unknown"
)
