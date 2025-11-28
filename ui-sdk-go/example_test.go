package uimessagestream_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	uimessagestream "github.com/vercel/ai-sdk-go/uimessagestream"
)

// Example_serverHandler demonstrates how to create an SSE endpoint
// that streams AI responses using the UI Message Stream Protocol.
func Example_serverHandler() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := uimessagestream.NewStreamWriter(w)

		// Start the stream
		sw.Start("msg-123", map[string]string{"model": "gpt-4"})

		// Stream some text
		sw.WriteChunk(&uimessagestream.TextStartChunk{ID: "text-1"})
		sw.WriteChunk(&uimessagestream.TextDeltaChunk{ID: "text-1", Delta: "Hello"})
		sw.WriteChunk(&uimessagestream.TextDeltaChunk{ID: "text-1", Delta: ", "})
		sw.WriteChunk(&uimessagestream.TextDeltaChunk{ID: "text-1", Delta: "world!"})
		sw.WriteChunk(&uimessagestream.TextEndChunk{ID: "text-1"})

		// Finish the stream
		sw.Finish(uimessagestream.FinishReasonStop, nil)
	})

	// Test the handler
	req := httptest.NewRequest("POST", "/api/chat", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Print the response
	fmt.Println("Content-Type:", rec.Header().Get("Content-Type"))
	fmt.Println("Protocol Header:", rec.Header().Get(uimessagestream.ProtocolHeader))
	fmt.Println("Contains DONE:", strings.Contains(rec.Body.String(), "[DONE]"))

	// Output:
	// Content-Type: text/event-stream
	// Protocol Header: v1
	// Contains DONE: true
}

// Example_clientDecoder demonstrates how to decode an SSE stream
// from a server response.
func Example_clientDecoder() {
	// Simulated SSE response
	sseData := `data: {"type":"start","messageId":"msg-123"}

data: {"type":"text-start","id":"text-1"}

data: {"type":"text-delta","id":"text-1","delta":"Hello, world!"}

data: {"type":"text-end","id":"text-1"}

data: {"type":"finish","finishReason":"stop"}

data: [DONE]

`

	// Decode the stream
	reader := uimessagestream.NewStreamReader(strings.NewReader(sseData))

	var textContent string
	for {
		chunk, err := reader.Next()
		if err != nil {
			break
		}

		switch c := chunk.(type) {
		case *uimessagestream.StartChunk:
			fmt.Println("Started message:", c.MessageID)
		case *uimessagestream.TextDeltaChunk:
			textContent += c.Delta
		case *uimessagestream.FinishChunk:
			fmt.Println("Finished with reason:", c.FinishReason)
		}
	}

	fmt.Println("Text content:", textContent)

	// Output:
	// Started message: msg-123
	// Finished with reason: stop
	// Text content: Hello, world!
}

// Example_toolCall demonstrates streaming a tool invocation.
func Example_toolCall() {
	var buf bytes.Buffer
	enc := uimessagestream.NewEncoder(&buf)

	// Start
	enc.Encode(&uimessagestream.StartChunk{MessageID: "msg-456"})

	// Text before tool call
	enc.Encode(&uimessagestream.TextStartChunk{ID: "text-1"})
	enc.Encode(&uimessagestream.TextDeltaChunk{ID: "text-1", Delta: "Let me check the weather."})
	enc.Encode(&uimessagestream.TextEndChunk{ID: "text-1"})

	// Tool call
	enc.Encode(&uimessagestream.ToolInputStartChunk{
		ToolCallID: "call-1",
		ToolName:   "getWeather",
	})
	enc.Encode(&uimessagestream.ToolInputDeltaChunk{
		ToolCallID:     "call-1",
		InputTextDelta: `{"location":`,
	})
	enc.Encode(&uimessagestream.ToolInputDeltaChunk{
		ToolCallID:     "call-1",
		InputTextDelta: `"San Francisco"}`,
	})
	enc.Encode(&uimessagestream.ToolInputAvailableChunk{
		ToolCallID: "call-1",
		ToolName:   "getWeather",
		Input:      map[string]string{"location": "San Francisco"},
	})

	// Tool output
	enc.Encode(&uimessagestream.ToolOutputAvailableChunk{
		ToolCallID: "call-1",
		Output:     map[string]any{"temperature": 18, "condition": "sunny"},
	})

	// Finish
	enc.Encode(&uimessagestream.FinishChunk{FinishReason: uimessagestream.FinishReasonStop})
	enc.Done()

	// Decode and process
	chunks, _ := uimessagestream.DecodeFromString(buf.String())
	fmt.Println("Chunk count:", len(chunks))

	for _, chunk := range chunks {
		if toolOutput, ok := chunk.(*uimessagestream.ToolOutputAvailableChunk); ok {
			output := toolOutput.Output.(map[string]any)
			fmt.Println("Weather temperature:", output["temperature"])
		}
	}

	// Output:
	// Chunk count: 10
	// Weather temperature: 18
}

// Example_chunkIterator demonstrates using the iterator pattern.
func Example_chunkIterator() {
	sseData := `data: {"type":"start","messageId":"msg-123"}

data: {"type":"text-delta","id":"text-1","delta":"Hello"}

data: {"type":"text-delta","id":"text-1","delta":" world"}

data: {"type":"finish","finishReason":"stop"}

data: [DONE]

`

	iter := uimessagestream.NewChunkIterator(strings.NewReader(sseData))

	textDeltas := 0
	for iter.Next() {
		if _, ok := iter.Chunk().(*uimessagestream.TextDeltaChunk); ok {
			textDeltas++
		}
	}

	if err := iter.Err(); err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println("Text deltas:", textDeltas)

	// Output:
	// Text deltas: 2
}

// Example_customDataChunk demonstrates using custom data chunks.
func Example_customDataChunk() {
	var buf bytes.Buffer
	enc := uimessagestream.NewEncoder(&buf)

	enc.Encode(&uimessagestream.StartChunk{MessageID: "msg-789"})

	// Custom weather data
	enc.Encode(&uimessagestream.DataChunk{
		Type: "data-weather",
		ID:   "weather-1",
		Data: map[string]any{
			"location":    "San Francisco",
			"temperature": 18,
			"condition":   "sunny",
		},
	})

	// Custom chart data
	transient := true
	enc.Encode(&uimessagestream.DataChunk{
		Type:      "data-chart",
		ID:        "chart-1",
		Data:      []int{10, 20, 30, 40, 50},
		Transient: &transient,
	})

	enc.Encode(&uimessagestream.FinishChunk{FinishReason: uimessagestream.FinishReasonStop})
	enc.Done()

	// Decode
	chunks, _ := uimessagestream.DecodeFromString(buf.String())

	for _, chunk := range chunks {
		if dataChunk, ok := chunk.(*uimessagestream.DataChunk); ok {
			fmt.Println("Data type:", dataChunk.DataName())
		}
	}

	// Output:
	// Data type: weather
	// Data type: chart
}

// Example_reasoning demonstrates streaming reasoning content.
func Example_reasoning() {
	var buf bytes.Buffer
	enc := uimessagestream.NewEncoder(&buf)

	enc.Encode(&uimessagestream.StartChunk{MessageID: "msg-reason"})

	// Reasoning block
	enc.Encode(&uimessagestream.ReasoningStartChunk{ID: "reason-1"})
	enc.Encode(&uimessagestream.ReasoningDeltaChunk{ID: "reason-1", Delta: "Let me think about this problem..."})
	enc.Encode(&uimessagestream.ReasoningDeltaChunk{ID: "reason-1", Delta: " I need to consider multiple factors."})
	enc.Encode(&uimessagestream.ReasoningEndChunk{ID: "reason-1"})

	// Response text
	enc.Encode(&uimessagestream.TextStartChunk{ID: "text-1"})
	enc.Encode(&uimessagestream.TextDeltaChunk{ID: "text-1", Delta: "Based on my analysis, the answer is 42."})
	enc.Encode(&uimessagestream.TextEndChunk{ID: "text-1"})

	enc.Encode(&uimessagestream.FinishChunk{FinishReason: uimessagestream.FinishReasonStop})
	enc.Done()

	// Decode and collect content
	chunks, _ := uimessagestream.DecodeFromString(buf.String())

	var reasoning, response string
	for _, chunk := range chunks {
		switch c := chunk.(type) {
		case *uimessagestream.ReasoningDeltaChunk:
			reasoning += c.Delta
		case *uimessagestream.TextDeltaChunk:
			response += c.Delta
		}
	}

	fmt.Println("Has reasoning:", len(reasoning) > 0)
	fmt.Println("Has response:", len(response) > 0)

	// Output:
	// Has reasoning: true
	// Has response: true
}
