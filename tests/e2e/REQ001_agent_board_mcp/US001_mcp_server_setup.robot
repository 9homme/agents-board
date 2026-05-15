*** Settings ***
Resource          mcp_keywords.resource
Test Tags         US001

*** Test Cases ***
Scenario: MCP Server SSE Connection Setup
    [Documentation]    Verify that the MCP server accepts SSE connections and returns a session ID.
    [Tags]    E2E-001
    ${session_id}=    Connect To MCP SSE
    Should Not Be Empty    ${session_id}
