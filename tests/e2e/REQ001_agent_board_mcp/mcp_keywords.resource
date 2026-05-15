*** Settings ***
Library    RequestsLibrary
Library    String
Library    Collections
Library    sse_helper.py

*** Variables ***
${BASE_URL}    http://localhost:8080

*** Keywords ***
Connect To MCP SSE
    [Documentation]    Establishes an SSE connection and returns the session ID.
    ${session_id}=    Get Session Id From Sse    ${BASE_URL}/sse
    RETURN    ${session_id}

Call MCP Tool
    [Arguments]    ${session_id}    ${tool_name}    ${arguments}
    ${payload}=    Create Dictionary    jsonrpc=2.0    id=1    method=tools/call
    ${params}=    Create Dictionary    name=${tool_name}    arguments=${arguments}
    Set To Dictionary    ${payload}    params=${params}
    ${response}=    POST    ${BASE_URL}/message    params=sessionId=${session_id}    json=${payload}    expected_status=200
    # The response to the POST is just an ACK (200 OK).
    # The actual tool result comes back over SSE.
    # For this E2E spec implementation, we'll assume a synchronous-like behavior or use a mockable approach.
    # Re-reading the architecture: "API-->>C: SSE event: message (JSON-RPC response with project ID & details)"
    # This means we need a way to listen to the SSE stream.
    # Since I'm writing the spec/stub, I'll define the keyword and the dev will need to make it work.
    RETURN    ${response}

# Additional keywords for specific tools
Create Project Tool
    [Arguments]    ${session_id}    ${name}    ${description}=${EMPTY}
    ${args}=    Create Dictionary    name=${name}    description=${description}
    ${resp}=    Call MCP Tool    ${session_id}    create_project    ${args}
    RETURN    ${resp}

Get Project Tool
    [Arguments]    ${session_id}    ${id}
    ${args}=    Create Dictionary    id=${id}
    ${resp}=    Call MCP Tool    ${session_id}    get_project    ${args}
    RETURN    ${resp}

Update Project Tool
    [Arguments]    ${session_id}    ${id}    ${name}=${NONE}    ${description}=${NONE}
    ${args}=    Create Dictionary    id=${id}
    Run Keyword If    "${name}" != "${NONE}"    Set To Dictionary    ${args}    name=${name}
    Run Keyword If    "${description}" != "${NONE}"    Set To Dictionary    ${args}    description=${description}
    ${resp}=    Call MCP Tool    ${session_id}    update_project    ${args}
    RETURN    ${resp}

Delete Project Tool
    [Arguments]    ${session_id}    ${id}
    ${args}=    Create Dictionary    id=${id}
    ${resp}=    Call MCP Tool    ${session_id}    delete_project    ${args}
    RETURN    ${resp}

List Projects Tool
    [Arguments]    ${session_id}
    ${args}=    Create Dictionary
    ${resp}=    Call MCP Tool    ${session_id}    list_projects    ${args}
    RETURN    ${resp}

# Document Tools
Create Document Tool
    [Arguments]    ${session_id}    ${project_id}    ${title}    ${content}
    ${args}=    Create Dictionary    projectId=${project_id}    title=${title}    content=${content}
    ${resp}=    Call MCP Tool    ${session_id}    create_document    ${args}
    RETURN    ${resp}

Get Document Tool
    [Arguments]    ${session_id}    ${id}
    ${args}=    Create Dictionary    id=${id}
    ${resp}=    Call MCP Tool    ${session_id}    get_document    ${args}
    RETURN    ${resp}

Update Document Tool
    [Arguments]    ${session_id}    ${id}    ${title}=${NONE}    ${content}=${NONE}
    ${args}=    Create Dictionary    id=${id}
    Run Keyword If    "${title}" != "${NONE}"    Set To Dictionary    ${args}    title=${title}
    Run Keyword If    "${content}" != "${NONE}"    Set To Dictionary    ${args}    content=${content}
    ${resp}=    Call MCP Tool    ${session_id}    update_document    ${args}
    RETURN    ${resp}

Delete Document Tool
    [Arguments]    ${session_id}    ${id}
    ${args}=    Create Dictionary    id=${id}
    ${resp}=    Call MCP Tool    ${session_id}    delete_document    ${args}
    RETURN    ${resp}

List Documents Tool
    [Arguments]    ${session_id}    ${project_id}
    ${args}=    Create Dictionary    projectId=${project_id}
    ${resp}=    Call MCP Tool    ${session_id}    list_documents    ${args}
    RETURN    ${resp}

# User Story Tools
Create User Story Tool
    [Arguments]    ${session_id}    ${project_id}    ${title}    ${description}    ${status}=draft
    ${args}=    Create Dictionary    projectId=${project_id}    title=${title}    description=${description}    status=${status}
    ${resp}=    Call MCP Tool    ${session_id}    create_user_story    ${args}
    RETURN    ${resp}

Get User Story Tool
    [Arguments]    ${session_id}    ${id}
    ${args}=    Create Dictionary    id=${id}
    ${resp}=    Call MCP Tool    ${session_id}    get_user_story    ${args}
    RETURN    ${resp}

Update User Story Tool
    [Arguments]    ${session_id}    ${id}    ${title}=${NONE}    ${description}=${NONE}    ${status}=${NONE}
    ${args}=    Create Dictionary    id=${id}
    Run Keyword If    "${title}" != "${NONE}"    Set To Dictionary    ${args}    title=${title}
    Run Keyword If    "${description}" != "${NONE}"    Set To Dictionary    ${args}    description=${description}
    Run Keyword If    "${status}" != "${NONE}"    Set To Dictionary    ${args}    status=${status}
    ${resp}=    Call MCP Tool    ${session_id}    update_user_story    ${args}
    RETURN    ${resp}

Delete User Story Tool
    [Arguments]    ${session_id}    ${id}
    ${args}=    Create Dictionary    id=${id}
    ${resp}=    Call MCP Tool    ${session_id}    delete_user_story    ${args}
    RETURN    ${resp}

List User Stories Tool
    [Arguments]    ${session_id}    ${project_id}
    ${args}=    Create Dictionary    projectId=${project_id}
    ${resp}=    Call MCP Tool    ${session_id}    list_user_stories    ${args}
    RETURN    ${resp}

# Task Tools
Create Task Tool
    [Arguments]    ${session_id}    ${user_story_id}    ${title}    ${description}    ${status}=pending
    ${args}=    Create Dictionary    userStoryId=${user_story_id}    title=${title}    description=${description}    status=${status}
    ${resp}=    Call MCP Tool    ${session_id}    create_task    ${args}
    RETURN    ${resp}

Get Task Tool
    [Arguments]    ${session_id}    ${id}
    ${args}=    Create Dictionary    id=${id}
    ${resp}=    Call MCP Tool    ${session_id}    get_task    ${args}
    RETURN    ${resp}

Update Task Tool
    [Arguments]    ${session_id}    ${id}    ${title}=${NONE}    ${description}=${NONE}    ${status}=${NONE}
    ${args}=    Create Dictionary    id=${id}
    Run Keyword If    "${title}" != "${NONE}"    Set To Dictionary    ${args}    title=${title}
    Run Keyword If    "${description}" != "${NONE}"    Set To Dictionary    ${args}    description=${description}
    Run Keyword If    "${status}" != "${NONE}"    Set To Dictionary    ${args}    status=${status}
    ${resp}=    Call MCP Tool    ${session_id}    update_task    ${args}
    RETURN    ${resp}

Delete Task Tool
    [Arguments]    ${session_id}    ${id}
    ${args}=    Create Dictionary    id=${id}
    ${resp}=    Call MCP Tool    ${session_id}    delete_task    ${args}
    RETURN    ${resp}

List Tasks Tool
    [Arguments]    ${session_id}    ${user_story_id}
    ${args}=    Create Dictionary    userStoryId=${user_story_id}
    ${resp}=    Call MCP Tool    ${session_id}    list_tasks    ${args}
    RETURN    ${resp}
