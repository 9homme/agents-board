*** Settings ***
Documentation    Shared keywords for REQ005 quality hardening retrospective e2e tests.
...              These keywords use RequestsLibrary for REST API calls against the
...              api-server (port 8080) to set up test data without needing the MCP
...              server (which is NOT in the REQ005 compose stack per architecture §6.2).
Library          RequestsLibrary
Library          Browser
Library          String
Library          Collections

*** Variables ***
${API_BASE_URL}    http://localhost:8080
${WEB_BASE_URL}    http://localhost:3000

*** Keywords ***
Create API Session
    [Documentation]    Creates a persistent RequestsLibrary session named 'api'.
    Create Session    api    ${API_BASE_URL}    verify=False

Create Project Via API
    [Documentation]    POSTs to /api/v1/projects and returns the created project dict.
    [Arguments]    ${name}    ${description}=${EMPTY}
    ${body}=    Create Dictionary    name=${name}    description=${description}
    ${resp}=    POST On Session    api    /api/v1/projects    json=${body}    expected_status=201
    RETURN    ${resp.json()}

Create Document Via API
    [Documentation]    POSTs to /api/v1/projects/{project_id}/documents and returns the created document dict.
    [Arguments]    ${project_id}    ${title}    ${content}
    ${body}=    Create Dictionary    title=${title}    content=${content}
    ${resp}=    POST On Session    api    /api/v1/projects/${project_id}/documents    json=${body}    expected_status=201
    RETURN    ${resp.json()}

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
