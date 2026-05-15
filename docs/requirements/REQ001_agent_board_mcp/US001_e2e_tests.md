# US001 E2E Tests — MCP server setup

## E2E Tests

| ID | Test Case | Steps | Expected Outcome |
|---|---|---|---|
| E2E-001 | Verify SSE connection | 1. Connect to `GET /sse`<br>2. Wait for `endpoint` event | Connection is successful; `endpoint` event contains the message URL. |
| E2E-002 | Verify tool discovery | 1. Establish session<br>2. Call `tools/list` | Returns a list of available tools including `create_project`, etc. |
