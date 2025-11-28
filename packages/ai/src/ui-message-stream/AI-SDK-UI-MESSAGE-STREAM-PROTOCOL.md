# AI SDK UI Message Stream Protocol Specification

**Version**: 1.0
**Protocol Header**: `x-vercel-ai-ui-message-stream: v1`

## Overview

The AI SDK UI Message Stream Protocol defines how AI responses are streamed from server to browser using Server-Sent Events (SSE). It provides real-time streaming of text, reasoning, tool invocations, files, sources, and custom data parts.

## Transport Layer

### SSE Format

The protocol uses **Server-Sent Events (SSE)** as the underlying transport mechanism.

#### HTTP Headers

```http
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
X-Vercel-AI-UI-Message-Stream: v1
X-Accel-Buffering: no
```

The `X-Accel-Buffering: no` header disables nginx buffering for proper streaming.

#### Wire Format

Each message chunk is encoded as a JSON object prefixed with `data: ` and terminated with double newlines:

```
data: {"type":"text-start","id":"text-1"}\n\n
data: {"type":"text-delta","id":"text-1","delta":"Hello"}\n\n
data: {"type":"text-delta","id":"text-1","delta":" world"}\n\n
data: {"type":"text-end","id":"text-1"}\n\n
data: {"type":"finish","finishReason":"stop"}\n\n
data: [DONE]\n\n
```

The stream terminates with `[DONE]` marker.

### Encoding Pipeline

**Server-Side (Encoding)**:
```
UIMessageChunk Objects
    ↓ JsonToSseTransformStream
SSE-formatted strings ("data: {...}\n\n")
    ↓ TextEncoderStream
UTF-8 bytes
    ↓
HTTP Response body
```

**Client-Side (Decoding)**:
```
HTTP Response body (Uint8Array)
    ↓ TextDecoderStream
UTF-8 strings
    ↓ EventSourceParserStream
Parsed SSE events
    ↓ JSON.parse + Schema Validation
UIMessageChunk Objects
```

---

## Message Chunk Types

All chunks have a `type` field that determines their structure. The protocol defines 24 chunk types organized into categories.

### Flow Control Chunks

#### `start`
Initializes the stream and optionally sets message ID and metadata.

```typescript
{
  type: "start";
  messageId?: string;          // Unique message identifier
  messageMetadata?: unknown;   // Custom metadata for the message
}
```

#### `finish`
Signals stream completion with finish reason.

```typescript
{
  type: "finish";
  finishReason?: "stop" | "length" | "content-filter" | "tool-calls" | "error" | "other" | "unknown";
  messageMetadata?: unknown;
}
```

#### `abort`
Signals stream was aborted.

```typescript
{
  type: "abort";
}
```

#### `message-metadata`
Updates message metadata during streaming.

```typescript
{
  type: "message-metadata";
  messageMetadata: unknown;
}
```

---

### Text Streaming Chunks

Text content is streamed using a start/delta/end pattern with unique IDs to support multiple concurrent text parts.

#### `text-start`
Begins a new text part.

```typescript
{
  type: "text-start";
  id: string;                           // Unique identifier for this text part
  providerMetadata?: ProviderMetadata;  // Provider-specific metadata
}
```

#### `text-delta`
Appends text content to an active text part.

```typescript
{
  type: "text-delta";
  id: string;                           // Must match a text-start id
  delta: string;                        // Text content to append
  providerMetadata?: ProviderMetadata;
}
```

#### `text-end`
Finalizes a text part.

```typescript
{
  type: "text-end";
  id: string;                           // Must match a text-start id
  providerMetadata?: ProviderMetadata;
}
```

**Example Sequence**:
```json
{"type":"text-start","id":"text-1"}
{"type":"text-delta","id":"text-1","delta":"Hello"}
{"type":"text-delta","id":"text-1","delta":" world!"}
{"type":"text-end","id":"text-1"}
```

---

### Reasoning Chunks

Extended thinking/reasoning content follows the same pattern as text.

#### `reasoning-start`
Begins a reasoning block.

```typescript
{
  type: "reasoning-start";
  id: string;
  providerMetadata?: ProviderMetadata;
}
```

#### `reasoning-delta`
Appends reasoning content.

```typescript
{
  type: "reasoning-delta";
  id: string;
  delta: string;
  providerMetadata?: ProviderMetadata;
}
```

#### `reasoning-end`
Finalizes a reasoning block.

```typescript
{
  type: "reasoning-end";
  id: string;
  providerMetadata?: ProviderMetadata;
}
```

---

### Tool Invocation Chunks

Tool calls support streaming input arguments, approval workflows, and output results.

#### `tool-input-start`
Begins a tool invocation.

```typescript
{
  type: "tool-input-start";
  toolCallId: string;         // Unique identifier for this tool call
  toolName: string;           // Name of the tool being invoked
  providerExecuted?: boolean; // True if provider will execute the tool
  dynamic?: boolean;          // True for dynamic (runtime-registered) tools
  title?: string;             // Human-readable title for the tool call
}
```

#### `tool-input-delta`
Streams partial JSON input arguments.

```typescript
{
  type: "tool-input-delta";
  toolCallId: string;
  inputTextDelta: string;     // Partial JSON string to append
}
```

The client uses partial JSON parsing to extract incomplete objects for UI updates.

#### `tool-input-available`
Signals complete, validated tool input is ready.

```typescript
{
  type: "tool-input-available";
  toolCallId: string;
  toolName: string;
  input: unknown;                       // Complete, parsed input object
  providerExecuted?: boolean;
  providerMetadata?: ProviderMetadata;
  dynamic?: boolean;
  title?: string;
}
```

#### `tool-input-error`
Signals an error occurred during input parsing.

```typescript
{
  type: "tool-input-error";
  toolCallId: string;
  toolName: string;
  input: unknown;                       // Partial or invalid input
  errorText: string;                    // Error description
  providerExecuted?: boolean;
  providerMetadata?: ProviderMetadata;
  dynamic?: boolean;
  title?: string;
}
```

#### `tool-approval-request`
Requests user approval before tool execution.

```typescript
{
  type: "tool-approval-request";
  approvalId: string;         // Unique identifier for this approval request
  toolCallId: string;
}
```

#### `tool-output-available`
Provides tool execution result.

```typescript
{
  type: "tool-output-available";
  toolCallId: string;
  output: unknown;            // Tool execution result
  providerExecuted?: boolean;
  dynamic?: boolean;
  preliminary?: boolean;      // True if this is a partial/streaming result
}
```

#### `tool-output-error`
Signals tool execution failed.

```typescript
{
  type: "tool-output-error";
  toolCallId: string;
  errorText: string;          // Error description
  providerExecuted?: boolean;
  dynamic?: boolean;
}
```

#### `tool-output-denied`
Signals user denied tool execution approval.

```typescript
{
  type: "tool-output-denied";
  toolCallId: string;
}
```

**Tool Call State Machine**:
```
                    ┌─────────────────┐
                    │  tool-input-    │
                    │     start       │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
┌───────────────┐   ┌────────────────┐   ┌────────────────┐
│ tool-input-   │   │ tool-input-    │   │ tool-input-    │
│    delta      │──▶│   available    │   │    error       │
│ (streaming)   │   └───────┬────────┘   └────────────────┘
└───────────────┘           │
                            │
         ┌──────────────────┼──────────────────┐
         │                  │                  │
         ▼                  ▼                  ▼
┌────────────────┐  ┌───────────────┐  ┌───────────────┐
│ tool-approval- │  │ tool-output-  │  │ tool-output-  │
│   request      │  │  available    │  │    error      │
└───────┬────────┘  └───────────────┘  └───────────────┘
        │
   ┌────┴─────┐
   │          │
   ▼          ▼
┌─────────┐  ┌─────────────────┐
│approved │  │ tool-output-    │
│         │  │    denied       │
└────┬────┘  └─────────────────┘
     │
     ▼
┌────────────────┐
│ tool-output-   │
│  available     │
└────────────────┘
```

---

### Source Chunks

References to source materials.

#### `source-url`
Web URL reference.

```typescript
{
  type: "source-url";
  sourceId: string;
  url: string;
  title?: string;
  providerMetadata?: ProviderMetadata;
}
```

#### `source-document`
Document reference.

```typescript
{
  type: "source-document";
  sourceId: string;
  mediaType: string;          // IANA media type
  title: string;
  filename?: string;
  providerMetadata?: ProviderMetadata;
}
```

---

### File Chunk

#### `file`
File attachment.

```typescript
{
  type: "file";
  url: string;                // URL or Data URL
  mediaType: string;          // IANA media type
  providerMetadata?: ProviderMetadata;
}
```

---

### Step Boundary Chunks

For multi-step reasoning workflows.

#### `start-step`
Marks the beginning of a new reasoning step.

```typescript
{
  type: "start-step";
}
```

#### `finish-step`
Marks the end of a reasoning step. Resets active text and reasoning parts.

```typescript
{
  type: "finish-step";
}
```

---

### Custom Data Chunks

#### `data-{name}`
Custom data parts with dynamic type names.

```typescript
{
  type: `data-${string}`;     // e.g., "data-weather", "data-chart"
  id?: string;                // Optional ID for updates
  data: unknown;              // Custom payload
  transient?: boolean;        // If true, not persisted to message state
}
```

Data parts with the same `type` and `id` will update existing parts instead of creating new ones.

---

### Error Chunk

#### `error`
Error notification.

```typescript
{
  type: "error";
  errorText: string;
}
```

---

## UI Message Structure

Chunks are processed into `UIMessage` objects for rendering:

```typescript
interface UIMessage {
  id: string;
  role: "system" | "user" | "assistant";
  metadata?: unknown;
  parts: UIMessagePart[];
}

type UIMessagePart =
  | TextUIPart
  | ReasoningUIPart
  | ToolUIPart
  | DynamicToolUIPart
  | SourceUrlUIPart
  | SourceDocumentUIPart
  | FileUIPart
  | DataUIPart
  | StepStartUIPart;
```

### Text UI Part

```typescript
interface TextUIPart {
  type: "text";
  text: string;
  state?: "streaming" | "done";
  providerMetadata?: ProviderMetadata;
}
```

### Reasoning UI Part

```typescript
interface ReasoningUIPart {
  type: "reasoning";
  text: string;
  state?: "streaming" | "done";
  providerMetadata?: ProviderMetadata;
}
```

### Tool UI Part

Static tools (registered at compile time):

```typescript
interface ToolUIPart {
  type: `tool-${toolName}`;
  toolCallId: string;
  title?: string;
  providerExecuted?: boolean;
  state: "input-streaming" | "input-available" | "approval-requested" |
         "approval-responded" | "output-available" | "output-error" | "output-denied";
  input?: unknown;
  output?: unknown;
  errorText?: string;
  preliminary?: boolean;
  callProviderMetadata?: ProviderMetadata;
  approval?: {
    id: string;
    approved?: boolean;
    reason?: string;
  };
}
```

### Dynamic Tool UI Part

Dynamic tools (registered at runtime):

```typescript
interface DynamicToolUIPart {
  type: "dynamic-tool";
  toolName: string;
  toolCallId: string;
  title?: string;
  providerExecuted?: boolean;
  state: /* same states as ToolUIPart */;
  input?: unknown;
  output?: unknown;
  errorText?: string;
  preliminary?: boolean;
  callProviderMetadata?: ProviderMetadata;
  approval?: /* same as ToolUIPart */;
}
```

### Source URL UI Part

```typescript
interface SourceUrlUIPart {
  type: "source-url";
  sourceId: string;
  url: string;
  title?: string;
  providerMetadata?: ProviderMetadata;
}
```

### Source Document UI Part

```typescript
interface SourceDocumentUIPart {
  type: "source-document";
  sourceId: string;
  mediaType: string;
  title: string;
  filename?: string;
  providerMetadata?: ProviderMetadata;
}
```

### File UI Part

```typescript
interface FileUIPart {
  type: "file";
  mediaType: string;
  filename?: string;
  url: string;
  providerMetadata?: ProviderMetadata;
}
```

### Data UI Part

```typescript
interface DataUIPart {
  type: `data-${name}`;
  id?: string;
  data: unknown;
}
```

### Step Start UI Part

```typescript
interface StepStartUIPart {
  type: "step-start";
}
```

---

## Provider Metadata

Optional provider-specific metadata can be attached to chunks:

```typescript
type ProviderMetadata = Record<string, Record<string, JSONValue | undefined>>;
```

Example:
```json
{
  "anthropic": {
    "cacheControl": "ephemeral",
    "stopSequence": "\n\nHuman:"
  }
}
```

---

## HTTP Transport

### Request Format

**Endpoint**: `POST /api/chat` (configurable)

**Request Headers**:
```http
Content-Type: application/json
```

**Request Body**:
```typescript
{
  id: string;                           // Chat session ID
  messages: UIMessage[];                // Conversation history
  trigger: "submit-message" | "regenerate-message";
  messageId?: string;                   // For regeneration
  // Additional custom fields
}
```

### Stream Reconnection

**Endpoint**: `GET /api/chat/{chatId}/stream`

Returns `204 No Content` if no active stream exists.

---

## Complete Example

### Simple Text Response

```
data: {"type":"start","messageId":"msg-123"}

data: {"type":"text-start","id":"text-1"}

data: {"type":"text-delta","id":"text-1","delta":"Hello"}

data: {"type":"text-delta","id":"text-1","delta":"! How can I help?"}

data: {"type":"text-end","id":"text-1"}

data: {"type":"finish","finishReason":"stop"}

data: [DONE]

```

### Tool Call with Streaming Input

```
data: {"type":"start","messageId":"msg-456"}

data: {"type":"text-start","id":"text-1"}

data: {"type":"text-delta","id":"text-1","delta":"Let me check the weather."}

data: {"type":"text-end","id":"text-1"}

data: {"type":"tool-input-start","toolCallId":"call-1","toolName":"getWeather"}

data: {"type":"tool-input-delta","toolCallId":"call-1","inputTextDelta":"{\"location\":"}

data: {"type":"tool-input-delta","toolCallId":"call-1","inputTextDelta":"\"San Francisco\",\"unit\":\"celsius\"}"}

data: {"type":"tool-input-available","toolCallId":"call-1","toolName":"getWeather","input":{"location":"San Francisco","unit":"celsius"}}

data: {"type":"tool-output-available","toolCallId":"call-1","output":{"temperature":18,"condition":"sunny"}}

data: {"type":"text-start","id":"text-2"}

data: {"type":"text-delta","id":"text-2","delta":"The weather in San Francisco is 18°C and sunny."}

data: {"type":"text-end","id":"text-2"}

data: {"type":"finish","finishReason":"stop"}

data: [DONE]

```

### Multi-Step Reasoning

```
data: {"type":"start","messageId":"msg-789"}

data: {"type":"reasoning-start","id":"reason-1"}

data: {"type":"reasoning-delta","id":"reason-1","delta":"Let me think through this problem step by step..."}

data: {"type":"reasoning-end","id":"reason-1"}

data: {"type":"text-start","id":"text-1"}

data: {"type":"text-delta","id":"text-1","delta":"Based on my analysis..."}

data: {"type":"text-end","id":"text-1"}

data: {"type":"finish-step"}

data: {"type":"start-step"}

data: {"type":"text-start","id":"text-2"}

data: {"type":"text-delta","id":"text-2","delta":"Here is the next step..."}

data: {"type":"text-end","id":"text-2"}

data: {"type":"finish","finishReason":"stop"}

data: [DONE]

```

---

## Schema Validation

All chunks are validated against Zod schemas on the client side. Invalid chunks will throw parsing errors.

---

## Error Handling

Errors can occur at multiple levels:

1. **HTTP Level**: Non-2xx status codes throw errors before streaming begins
2. **Stream Level**: `error` chunks notify the client of errors during streaming
3. **Parse Level**: Invalid JSON or schema validation failures throw errors

---

## Compatibility Notes

- The protocol is designed for browser `fetch` with `ReadableStream` support
- Uses the `eventsource-parser` library for SSE parsing
- Compatible with standard `Response` and Node.js `ServerResponse`
- Supports stream tee-ing for server-side consumption alongside client delivery
