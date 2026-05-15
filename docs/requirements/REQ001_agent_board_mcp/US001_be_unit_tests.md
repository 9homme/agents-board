# US001 Backend Unit Tests — MCP server setup

## Backend Unit/Integration Tests

| ID | Component | Test Case | Expected Outcome |
|---|---|---|---|
| UT-001 | `internal/mcp` | Session creation and message queuing | Session is created with unique ID; messages can be queued and retrieved. |
| UT-002 | `internal/handler` | `GET /sse` endpoint | Responds with 200 OK and `text/event-stream`. Sends initial `endpoint` event. |
| UT-003 | `internal/handler` | `POST /message` with invalid JSON-RPC | Responds with 400 Bad Request or error JSON-RPC. |
| UT-004 | `internal/handler` | `POST /message` with valid tool call | Parses `tools/call` and routes to correct tool handler. |
| IT-001 | `internal/handler` | Full handshake: SSE + POST | Client connects via SSE, sends a message via POST, and receives response via SSE. |

## Mocking Strategy
- Mock the tool registry to isolate handler logic from specific tool implementations.
- Use `httptest` for Echo handler testing.
- Mock SSE stream for message delivery verification.
