*** Settings ***
Documentation    Shared keywords for REQ005 quality hardening retrospective e2e tests.
...              Write operations go through MCP tools (the only write API in this system).
...              Read operations use the api-server REST surface (port 8080).
...              UI assertions use the Browser library against web (port 3000).
Library          RequestsLibrary
Library          Browser
Library          String
Library          Collections
Resource         ../../REQ001_agent_board_mcp/mcp_keywords.resource

*** Variables ***
${API_BASE_URL}    http://localhost:8080
${WEB_BASE_URL}    http://localhost:3000

*** Keywords ***
Create API Session
    [Documentation]    Creates a persistent RequestsLibrary session named 'api'.
    Create Session    api    ${API_BASE_URL}    verify=False

Create Project Via API
    [Documentation]    Creates a project via the MCP create_project tool and returns the project dict.
    ...                The api-server has no REST POST endpoint — MCP is the only write API
    ...                in this codebase (arch §6.2 + §8). This keyword wraps that fact.
    [Arguments]    ${name}    ${description}=${EMPTY}    ${path}=${EMPTY}
    ${session_id}=    Connect To MCP SSE
    ${resp}=    Create Project Tool    ${session_id}    ${name}    ${description}    ${path}
    ${text}=    Set Variable    ${resp.json()['result']['content'][0]['text']}
    ${project}=    Evaluate    json.loads('''${text}''')    json
    RETURN    ${project}

Create Document Via API
    [Documentation]    Creates a document via the MCP create_document tool and returns the document dict.
    [Arguments]    ${project_id}    ${requirement_id}    ${title}    ${content}
    ${session_id}=    Connect To MCP SSE
    ${resp}=    Create Document Tool    ${session_id}    ${project_id}    ${requirement_id}    ${title}    ${content}
    ${text}=    Set Variable    ${resp.json()['result']['content'][0]['text']}
    ${document}=    Evaluate    json.loads('''${text}''')    json
    RETURN    ${document}

Wait For API Ready
    [Documentation]    Polls GET /api/v1/projects until a 200 is returned, up to ${timeout}.
    ...                The stack must already be up before this keyword is called.
    [Arguments]    ${timeout}=30s
    Wait Until Keyword Succeeds    ${timeout}    2s    GET On Session    api    /api/v1/projects    expected_status=200

Assert No Error State Visible
    [Documentation]    Asserts that common FE error states are NOT visible on the current page.
    ...                Used after navigation to confirm no stale-data error or not-found state.
    ${failed_to_load}=    Run Keyword And Return Status
    ...    Get Element    text="Failed to load project"
    Should Be Equal    ${failed_to_load}    ${FALSE}
    ${not_found}=    Run Keyword And Return Status
    ...    Get Element    text="Project not found"
    Should Be Equal    ${not_found}    ${FALSE}
