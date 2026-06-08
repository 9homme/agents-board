*** Settings ***
Documentation    E2E tests for US001 — api-server runs DB migrations at startup.
...              Verifies that the API server is up and its endpoints succeed without undefined-table errors,
...              proving that migrations ran successfully before the server started accepting traffic.
Library          RequestsLibrary
Library          Collections

*** Variables ***
${API_BASE_URL}    http://localhost:8080

*** Test Cases ***
E2E-US001-001 API starts and serves requests immediately on a fresh migrated database
    [Documentation]    Verifies that a GET to /api/v1/projects returns 200 OK.
    ...                If tables were not created by migrations at startup, this would return a 500
    ...                database error.
    [Tags]    US001    smoke
    
    Create Session    api    ${API_BASE_URL}
    ${response}=    GET On Session    api    /api/v1/projects    expected_status=200
    Should Be Equal As Strings    ${response.status_code}    200
    # Response JSON should contain 'projects'
    ${json}=    Set Variable    ${response.json()}
    Dictionary Should Contain Key    ${json}    projects
