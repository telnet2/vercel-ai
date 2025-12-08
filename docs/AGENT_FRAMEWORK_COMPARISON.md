# Agent Framework Comparison: Vercel AI SDK vs OpenCode vs Mastra

A comprehensive analysis of three AI agent frameworks, their architectures, capabilities, and use cases.

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Framework Overview](#framework-overview)
3. [Agent Architecture Comparison](#agent-architecture-comparison)
4. [Tool Loop Implementation](#tool-loop-implementation)
5. [Session & Memory Management](#session--memory-management)
6. [Subagent & Task Orchestration](#subagent--task-orchestration)
7. [Built-in Tools Comparison](#built-in-tools-comparison)
8. [Generative UI Support](#generative-ui-support)
9. [Storage Backends](#storage-backends)
10. [Permission & Safety Systems](#permission--safety-systems)
11. [When to Use Each Framework](#when-to-use-each-framework)
12. [Code Examples](#code-examples)

---

## Executive Summary

| Framework | Focus | Complexity | Best For |
|-----------|-------|------------|----------|
| **Vercel AI SDK** | Primitives & Composability | Low | Building custom agent features |
| **OpenCode** | Production CLI Agent | High | IDE integrations, developer tools |
| **Mastra** | Agent Orchestration | Medium | Multi-agent systems, workflows |

### Key Differentiators

- **Vercel AI SDK**: Lightweight primitives (`ToolLoopAgent`, `streamUI`), no built-in session management
- **OpenCode**: Production-grade with TaskTool subagents, context compaction, permission system
- **Mastra**: Agent Network routing, DAG workflows, 24+ storage backends, built-in memory

---

## Framework Overview

### Vercel AI SDK v6

**Package**: `ai` (core), `ai/rsc` (React Server Components)
**Version Analyzed**: Latest beta
**Repository**: https://github.com/vercel/ai

```typescript
import { generateText, streamText, ToolLoopAgent } from 'ai';
import { streamUI, createStreamableUI } from 'ai/rsc';
```

**Architecture**:
- Single `ToolLoopAgent` class (~111 lines)
- Stateless request/response model
- Tool approval via callbacks
- Generative UI with React Server Components

### OpenCode (SST)

**Package**: `@opencode/opencode`
**AI SDK Version**: `5.0.97`
**Repository**: https://github.com/sst/opencode

```typescript
import { Session, SessionPrompt } from '@opencode/opencode';
```

**Architecture**:
- Custom tool loop implementation (NOT using `ToolLoopAgent`)
- Persistent session storage (file-based)
- 4 built-in agent types with hierarchical spawning
- 15+ built-in tools for development tasks

### Mastra

**Package**: `@mastra/core`, `@mastra/memory`, `@mastra/rag`
**AI SDK Version**: `5.0.97` (aliased as `ai-v5`)
**Repository**: https://github.com/mastra-ai/mastra

```typescript
import { Agent } from '@mastra/core';
import { Memory } from '@mastra/memory';
import { createVectorQueryTool } from '@mastra/rag';
```

**Architecture**:
- Full `Agent` class with memory, workflows, processors
- Agent Network for multi-agent routing
- DAG workflow engine with branching/loops
- 24+ persistent storage backends

---

## Agent Architecture Comparison

### Vercel AI SDK: ToolLoopAgent

**Location**: `packages/ai/src/agent/tool-loop-agent.ts`

```typescript
// Simple do-while loop
do {
  response = await generateText({ messages, tools });

  for (const toolCall of response.toolCalls) {
    if (needsApproval) {
      await onInputAvailable(toolCall);
    }
    result = await tool.execute(toolCall.args);
    messages.push(result);
  }

  stepCount++;
} while (hasToolCalls && !stopWhen(stepCount));

// Default: stopWhen = stepCountIs(20)
```

**Characteristics**:
- Stateless per invocation
- Maximum 20 steps by default
- Simple retry with exponential backoff
- No context overflow handling

### OpenCode: SessionPrompt Loop

**Location**: `packages/opencode/src/session/prompt.ts`

```typescript
// Production while(true) loop
while (true) {
  // 1. Stream messages from storage
  const messages = await filterAndStreamMessages();

  // 2. Check for pending subtasks
  if (pendingSubtask) {
    await handleSubtaskViaTaskTool();
  }

  // 3. Handle context overflow
  if (SessionCompaction.isOverflow()) {
    await compactContext();
  }

  // 4. Resolve agent-specific system prompts
  const systemPrompt = await resolveSystemPrompt(agent);

  // 5. Resolve available tools with hooks
  const tools = await resolveTools(agent, plugins);

  // 6. Process LLM call
  const result = await processor.process({
    messages, tools, systemPrompt,
    // Handles: reasoning tokens, tool calls, dead loops
  });

  // 7. Update message with cost tracking
  await updateMessageWithCosts(result);

  // 8. Break on completion
  if (result.finishReason !== 'tool-calls') break;
}
```

**Characteristics**:
- Persistent session state
- Automatic context compaction
- Hierarchical task spawning
- Cost and token tracking
- Snapshot-based rollback

### Mastra: Agent with Memory

**Location**: `packages/core/src/agent/agent.ts`

```typescript
class Agent {
  constructor({
    name,
    instructions,
    model,
    tools,
    memory,           // Built-in memory integration
    workflows,        // Workflow triggers
    outputProcessors, // Response transformation
    scorers,          // Quality evaluation
  }) {}

  async generate(options) {
    // 1. Recall context from memory
    const context = await this.memory?.recall({
      threadId, resourceId, vectorSearchString
    });

    // 2. Inject working memory into system prompt
    const systemPrompt = this.buildSystemPrompt(context);

    // 3. Run tool loop with AI SDK
    const result = await generateText({
      model: this.model,
      system: systemPrompt,
      messages,
      tools: this.tools,
    });

    // 4. Save messages to memory
    await this.memory?.saveMessages({ messages: result.messages });

    // 5. Apply output processors
    return this.processOutput(result);
  }
}
```

**Characteristics**:
- Built-in memory integration
- Output processors for transformation
- Workflow triggers
- Agent Network routing

---

## Tool Loop Implementation

### Comparison Table

| Feature | Vercel AI SDK | OpenCode | Mastra |
|---------|--------------|----------|--------|
| **Loop Type** | `do-while` | `while(true)` | `do-while` (via AI SDK) |
| **Max Steps** | 20 (configurable) | Unlimited | Configurable |
| **State** | Stateless | Persistent | Persistent (via Memory) |
| **Context Management** | Manual | Auto-compaction | Manual + Memory |
| **Streaming** | Full support | Via processor | Full support |
| **Telemetry** | OpenTelemetry | Custom events | OpenTelemetry |

### Context Overflow Handling

**Vercel AI SDK**: No built-in handling

**OpenCode**: Automatic compaction
```typescript
// Two-tier pruning strategy
PRUNE_PROTECT = 40_000;  // Preserve recent context
PRUNE_MINIMUM = 20_000;  // Only prune if worth it

if (tokenCount > modelLimit) {
  // Create CompactionPart
  // Summarize old messages
  // Compress and continue
}
```

**Mastra**: Via Memory configuration
```typescript
const memory = new Memory({
  options: {
    lastMessages: 50,  // Keep last N messages
    semanticRecall: {
      topK: 4,         // Retrieve relevant past messages
    }
  }
});
```

---

## Session & Memory Management

### Vercel AI SDK

**No built-in session management**. Developers must implement:

```typescript
// Manual implementation required
class SessionManager {
  async create(): Promise<string> { /* ... */ }
  async get(id: string): Promise<Session> { /* ... */ }
  async save(session: Session): Promise<void> { /* ... */ }
  async delete(id: string): Promise<void> { /* ... */ }
  async fork(id: string): Promise<string> { /* ... */ }
}
```

### OpenCode

**File-based session storage**:

```typescript
// Session hierarchy
Session = {
  id: string,
  parentID?: string,      // For subagent tasks
  messages: MessageV2[],
  snapshots: Snapshot[],  // Filesystem state
  cost: { tokens, usd },

  // Operations
  revert(messageId),      // Rollback to point
  share(),                // Export session
}
```

### Mastra

**24+ persistent storage backends**:

```typescript
import { Memory } from '@mastra/memory';
import { PostgresStore, PgVector } from '@mastra/pg';

const memory = new Memory({
  storage: new PostgresStore({
    connectionString: 'postgresql://...',
  }),
  vector: new PgVector({ connectionString: '...' }),
  embedder: openai.embedding('text-embedding-3-small'),
  options: {
    lastMessages: 100,
    semanticRecall: {
      enabled: true,
      topK: 4,
      scope: 'resource',  // Search all user threads
    },
    workingMemory: {
      enabled: true,
      scope: 'resource',  // Persist across threads
      template: '# User Profile\n- Name:\n- Preferences:',
    }
  }
});
```

**Supported Backends**:

| Category | Backends |
|----------|----------|
| **SQL** | PostgreSQL, LibSQL, MySQL, MSSQL, ClickHouse, DuckDB |
| **NoSQL** | MongoDB, DynamoDB, Couchbase, Elasticsearch |
| **Redis** | Upstash (Redis-compatible serverless) |
| **Serverless** | Cloudflare D1, Convex |
| **Vector** | pgvector, Pinecone, Qdrant, Chroma, Astra |

---

## Subagent & Task Orchestration

### Vercel AI SDK

**No built-in subagent support**. Manual implementation:

```typescript
// Wrap agent as tool
const subagentTool = tool({
  description: 'Delegate task to specialized agent',
  parameters: z.object({
    task: z.string(),
    agentType: z.enum(['researcher', 'coder']),
  }),
  execute: async ({ task, agentType }) => {
    const agent = getAgent(agentType);
    return await agent.run(task);
  },
});
```

### OpenCode: TaskTool

**Built-in hierarchical task spawning**:

```typescript
// 4 Agent Types
const agents = {
  build: {    // Primary - full permissions
    permission: { edit: 'allow', bash: { '*': 'allow' } }
  },
  plan: {     // Primary - restricted
    permission: { edit: 'deny', bash: { 'git*': 'allow' } }
  },
  general: {  // Subagent - multi-step tasks
    permission: { edit: 'allow', bash: { '*': 'ask' } }
  },
  explore: {  // Subagent - fast exploration
    permission: { edit: 'deny', bash: { '*': 'deny' } }
  },
};

// TaskTool implementation
const TaskTool = Tool.define('task', async () => ({
  parameters: z.object({
    prompt: z.string(),
    subagent_type: z.enum(['general', 'explore']),
    description: z.string(),
  }),
  execute: async (params, ctx) => {
    // Create child session
    const session = await Session.create({
      parentID: ctx.sessionID,
    });

    // Subscribe to updates
    Bus.subscribe(MessageV2.Event.PartUpdated, (evt) => {
      ctx.metadata({ summary: parts, sessionId: session.id });
    });

    // Run subagent
    const result = await SessionPrompt.prompt({
      sessionID: session.id,
      agentType: params.subagent_type,
    });

    return { output: result.text };
  },
}));
```

### Mastra: Agent Network

**Automatic routing between specialized agents**:

```typescript
import { AgentNetwork } from '@mastra/core';

const network = new AgentNetwork({
  agents: [researchAgent, codingAgent, reviewAgent],
  workflows: [deployWorkflow],
  routingAgent: new Agent({
    instructions: `
      You are a router. Available agents:
      - researcher: For information gathering
      - coder: For implementation
      - reviewer: For code review

      Route requests to the appropriate agent.
    `,
  }),
});

// Automatic routing
const result = await network.generate({
  messages: [{ role: 'user', content: 'Research and implement auth' }],
});
```

---

## Built-in Tools Comparison

### Vercel AI SDK

**No built-in tools**. Framework only:

```typescript
import { tool } from 'ai';

const myTool = tool({
  description: 'My custom tool',
  parameters: z.object({ /* ... */ }),
  execute: async (args) => { /* ... */ },
});
```

### OpenCode

**15+ development tools**:

| Tool | Purpose |
|------|---------|
| `Read` | Read file contents |
| `Write` | Write file contents |
| `Edit` | Edit file with search/replace |
| `Glob` | Find files by pattern |
| `Grep` | Search file contents |
| `Bash` | Execute shell commands |
| `WebFetch` | Fetch web content |
| `WebSearch` | Search the web |
| `Task` | Spawn subagent |
| `TodoWrite` | Track task progress |
| `NotebookEdit` | Edit Jupyter notebooks |
| `AskUser` | Request user input |
| `ExitPlanMode` | Transition from planning |

### Mastra

**6 built-in tools** (memory & RAG focused):

| Tool | Package | Purpose |
|------|---------|---------|
| `updateWorkingMemoryTool` | `@mastra/memory` | Update persistent memory |
| `createVectorQueryTool` | `@mastra/rag` | Semantic search |
| `createGraphRAGTool` | `@mastra/rag` | Graph-based RAG |
| `createDocumentChunkerTool` | `@mastra/rag` | Document chunking |
| `createTool` | `@mastra/core` | Tool framework |

```typescript
// Mastra tool creation
import { createTool } from '@mastra/core/tools';

const weatherTool = createTool({
  id: 'get-weather',
  description: 'Get weather for a location',
  inputSchema: z.object({ location: z.string() }),
  outputSchema: z.object({ temp: z.number(), condition: z.string() }),
  requireApproval: true,
  execute: async ({ location }, ctx) => {
    const data = await fetchWeather(location);
    return { temp: data.temp, condition: data.condition };
  },
});
```

---

## Generative UI Support

### Vercel AI SDK

**Full generative UI support** via React Server Components:

```typescript
import { streamUI, createStreamableUI } from 'ai/rsc';

const result = await streamUI({
  model: openai('gpt-4-turbo'),
  initial: <LoadingSpinner />,

  tools: {
    get_weather: {
      description: 'Get weather for a location',
      inputSchema: z.object({ location: z.string() }),

      // Generator yields React components
      generate: async function* ({ location }) {
        yield <WeatherSkeleton location={location} />;

        const data = await fetchWeather(location);

        return (
          <WeatherWidget
            location={location}
            temperature={data.temp}
            condition={data.condition}
          />
        );
      },
    },
  },
});
```

**Key APIs**:
- `streamUI()` - Stream React components from tools
- `createStreamableUI()` - Manage UI state with update/append/done
- `createStreamableValue()` - Stream generic values

### OpenCode

**No generative UI** - CLI/TUI focused.

### Mastra

**No generative UI**. Data-driven approach:

```typescript
// Tools return data, not components
const tool = createTool({
  execute: async ({ location }) => {
    return { temp: 72, condition: 'sunny' };  // Plain data
  },
});

// Frontend maps data to components
function ToolResult({ part }) {
  if (part.toolName === 'get_weather') {
    return <WeatherWidget {...part.output} />;
  }
  return <JsonViewer data={part.output} />;
}
```

---

## Storage Backends

### Feature Matrix

| Feature | PostgreSQL | Upstash | MongoDB | DynamoDB | LibSQL |
|---------|-----------|---------|---------|----------|--------|
| Thread Persistence | ✅ | ✅ | ✅ | ✅ | ✅ |
| Message History | ✅ | ✅ | ✅ | ✅ | ✅ |
| Working Memory | ✅ | ✅ | ✅ | ✅ | ✅ |
| Resource Scope | ✅ | ✅ | ✅ | ❌ | ✅ |
| Semantic Recall | ✅ | ✅ | ✅ | ❓ | ✅ |
| Vector Index | ✅ | ❌ | ✅ | ❌ | ❌ |
| Observability | ✅ | ❌ | ✅ | ❌ | ✅ |

### Configuration Examples

**PostgreSQL (Recommended for Production)**:
```typescript
import { PostgresStore, PgVector } from '@mastra/pg';

const storage = new PostgresStore({
  connectionString: 'postgresql://user:pass@localhost:5432/db',
  schemaName: 'public',
  max: 20,
});

const memory = new Memory({
  storage,
  vector: new PgVector({ connectionString: '...' }),
  embedder: openai.embedding('text-embedding-3-small'),
});
```

**Upstash (Redis-compatible Serverless)**:
```typescript
import { UpstashStore } from '@mastra/upstash';

const storage = new UpstashStore({
  url: 'https://your-instance.upstash.io',
  token: 'your-token',
});
```

**LibSQL (Local Development)**:
```typescript
import { LibSQLStore } from '@mastra/libsql';

const storage = new LibSQLStore({
  url: 'file:./agent-memory.db',
});
```

---

## Permission & Safety Systems

### Vercel AI SDK

**Basic callback-based approval**:

```typescript
const result = await generateText({
  tools: {
    dangerousTool: tool({
      execute: async (args) => { /* ... */ },
    }),
  },
  experimental_toToolCallApprovalRequest: (toolCall) => {
    if (toolCall.toolName === 'dangerousTool') {
      return { type: 'approval-required' };
    }
    return { type: 'approved' };
  },
});
```

### OpenCode

**Comprehensive permission system**:

```typescript
// Agent-level permissions
Agent.Info.permission = {
  edit: 'allow' | 'deny' | 'ask',
  bash: {
    '*': 'allow',
    'git*': 'allow',
    'find * -delete*': 'ask',
    'rm -rf*': 'deny',
  },
  webfetch: 'allow' | 'deny',
  doom_loop: 'ask',
  external_directory: 'ask',
};

// Runtime permission check
const allowed = await Permission.ask({
  type: 'bash',
  pattern: 'rm -rf /tmp/*',
  sessionID,
  messageID,
  callID,
  title: 'Delete temporary files',
});

// Returns: 'once' | 'always' | 'reject'
```

**Doom Loop Detection**:
```typescript
// Detects repeated identical tool calls
if (identicalToolCallCount >= 3) {
  const allowed = await Permission.ask({
    type: 'doom_loop',
    title: 'Agent appears stuck',
  });
  if (!allowed) break;  // User breaks the loop
}
```

### Mastra

**Tool-level approval**:

```typescript
const tool = createTool({
  id: 'dangerous-operation',
  requireApproval: true,  // Triggers approval flow
  execute: async (args, ctx) => {
    // Only runs after approval
  },
});
```

---

## When to Use Each Framework

### Use Vercel AI SDK When:

- Building simple agent features
- Need minimal dependencies
- Want stateless request/response patterns
- Need generative UI (streamUI)
- Fast iteration on core logic
- Building Next.js applications

### Use OpenCode Architecture When:

- Building production CLI/IDE tools
- Need persistent conversation history
- Require user approval workflows
- Want hierarchical task decomposition
- Need detailed execution audit trails
- Building developer tools

### Use Mastra When:

- Building multi-agent systems
- Need complex workflow orchestration
- Want built-in memory with multiple backends
- Need Agent Network routing
- Building enterprise applications
- Require RAG/vector search integration

---

## Code Examples

### Basic Agent with Vercel AI SDK

```typescript
import { generateText, tool } from 'ai';
import { openai } from '@ai-sdk/openai';

const result = await generateText({
  model: openai('gpt-4-turbo'),
  system: 'You are a helpful assistant.',
  messages: [{ role: 'user', content: 'Hello!' }],
  tools: {
    search: tool({
      description: 'Search the web',
      parameters: z.object({ query: z.string() }),
      execute: async ({ query }) => {
        return await searchWeb(query);
      },
    }),
  },
  maxSteps: 10,
});
```

### Session-Based Agent with OpenCode Pattern

```typescript
// Simplified OpenCode-style implementation
class AgentSession {
  private messages: Message[] = [];
  private snapshots: Map<string, FileSnapshot> = new Map();

  async run(prompt: string) {
    this.messages.push({ role: 'user', content: prompt });

    while (true) {
      // Check context overflow
      if (this.getTokenCount() > MODEL_LIMIT) {
        await this.compact();
      }

      const result = await generateText({
        model: this.model,
        messages: this.messages,
        tools: this.tools,
      });

      this.messages.push(...result.messages);
      this.trackCost(result.usage);

      if (result.finishReason !== 'tool-calls') {
        return result.text;
      }
    }
  }

  async spawnSubagent(task: string, type: AgentType) {
    const child = new AgentSession({ parentId: this.id, type });
    return await child.run(task);
  }

  async revert(messageId: string) {
    const snapshot = this.snapshots.get(messageId);
    await this.restoreFiles(snapshot);
    this.messages = this.messages.slice(0, this.getMessageIndex(messageId));
  }
}
```

### Multi-Agent System with Mastra

```typescript
import { Agent, AgentNetwork } from '@mastra/core';
import { Memory } from '@mastra/memory';
import { PostgresStore } from '@mastra/pg';

// Shared memory
const memory = new Memory({
  storage: new PostgresStore({ connectionString: '...' }),
  options: {
    lastMessages: 50,
    workingMemory: { enabled: true, scope: 'resource' },
  },
});

// Specialized agents
const researcher = new Agent({
  name: 'researcher',
  instructions: 'You research topics thoroughly.',
  model: openai('gpt-4-turbo'),
  memory,
  tools: [webSearchTool, vectorQueryTool],
});

const coder = new Agent({
  name: 'coder',
  instructions: 'You write clean, efficient code.',
  model: openai('gpt-4-turbo'),
  memory,
  tools: [fileReadTool, fileWriteTool],
});

// Agent network with routing
const network = new AgentNetwork({
  agents: [researcher, coder],
  routingAgent: new Agent({
    instructions: `
      Route to 'researcher' for information gathering.
      Route to 'coder' for implementation tasks.
    `,
  }),
});

// Automatic routing
const result = await network.generate({
  threadId: 'thread-123',
  resourceId: 'user-456',
  messages: [{
    role: 'user',
    content: 'Research best practices for auth and implement JWT'
  }],
});
```

### TUI Client for Mastra

```typescript
import React, { useState } from 'react';
import { render, Box, Text, useInput } from 'ink';
import TextInput from 'ink-text-input';
import { MastraClient } from '@mastra/client-js';

const client = new MastraClient({ baseUrl: 'http://localhost:4111' });

function Chat() {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const [streaming, setStreaming] = useState(false);

  const handleSubmit = async (value: string) => {
    setMessages(prev => [...prev, { role: 'user', content: value }]);
    setInput('');
    setStreaming(true);

    const agent = client.getAgent('my-agent');
    let response = '';

    for await (const chunk of await agent.stream({ messages })) {
      if (chunk.type === 'text-delta') {
        response += chunk.textDelta;
      }
    }

    setMessages(prev => [...prev, { role: 'assistant', content: response }]);
    setStreaming(false);
  };

  return (
    <Box flexDirection="column">
      {messages.map((msg, i) => (
        <Text key={i} color={msg.role === 'user' ? 'blue' : 'green'}>
          {msg.role}: {msg.content}
        </Text>
      ))}
      <TextInput value={input} onChange={setInput} onSubmit={handleSubmit} />
    </Box>
  );
}

render(<Chat />);
```

---

## Conclusion

Each framework serves different needs:

| Need | Recommendation |
|------|----------------|
| Simple agent features | Vercel AI SDK |
| Production CLI tools | OpenCode patterns |
| Multi-agent orchestration | Mastra |
| Generative UI | Vercel AI SDK |
| Enterprise memory | Mastra |
| Developer tools | OpenCode |

For most production applications, consider:
1. **Vercel AI SDK** as the foundation (primitives)
2. **Mastra** for agent orchestration and memory
3. **OpenCode patterns** for CLI/TUI implementations

---

## References

- Vercel AI SDK: https://sdk.vercel.ai/docs
- OpenCode: https://github.com/sst/opencode
- Mastra: https://mastra.ai/docs
- AI SDK v6 Agent RFC: https://github.com/vercel/ai/discussions

---

*Document generated: December 2024*
*Analysis based on: Vercel AI SDK (beta), OpenCode (latest), Mastra (latest)*
