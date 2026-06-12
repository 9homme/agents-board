*** Settings ***
Documentation    E2E tests for US046 — Add Project from web by linking a local path.
...
...              E2E-046-001: Full user journey — form opens, fills, submits, project appears (golden path).
...              E2E-046-002: Server rejects non-existent path → inline error rendered by FE.
...              E2E-046-003: Duplicate path returns 409 → correct inline DUPLICATE_PATH error.
Library          Browser
Library          RequestsLibrary
Library          Collections
Library          String
Resource         ../REQ001_agent_board_mcp/mcp_keywords.resource

Suite Setup      Open Browser For US046
Suite Teardown   Close Browser

*** Variables ***
${WEB_BASE_URL}     %{WEB_BASE_URL=http://localhost:3000}
${API_BASE_URL}     %{API_BASE_URL=http://localhost:8080}
${SESSION_ID}       ${EMPTY}
${DUPE_PROJ_ID}     ${EMPTY}

*** Keywords ***
Open Browser For US046
    ${session_id}=    Connect To MCP SSE
    Set Suite Variable    ${SESSION_ID}    ${session_id}
    New Browser    headless=True
    New Context

Create Dupe Fixture Project
    [Documentation]    Creates a project with path=/e2e/us046-fixture so duplicate-path test can conflict with it.
    ${random}=    Generate Random String    8    [LETTERS]
    ${body}=    Create Dictionary    name=Dupe Fixture ${random}    path=/e2e/us046-fixture
    ${response}=    POST    ${API_BASE_URL}/api/v1/projects
    ...    json=${body}
    ...    expected_status=201
    ${proj_id}=    Set Variable    ${response.json()}[id]
    Set Test Variable    ${DUPE_PROJ_ID}    ${proj_id}
    RETURN    ${proj_id}

Delete Dupe Fixture Project
    Run Keyword If    '${DUPE_PROJ_ID}' != '${EMPTY}'
    ...    Delete Project Tool    ${SESSION_ID}    ${DUPE_PROJ_ID}

Open Add Project Dialog
    [Documentation]    Navigates to dashboard and clicks "Add Project" to open the form.
    New Page    ${WEB_BASE_URL}
    Wait For Elements State    role=button >> text=/Add Project/i    visible    timeout=10s
    Click    role=button >> text=/Add Project/i
    Wait For Elements State    role=dialog    visible    timeout=5s

*** Test Cases ***

E2E-046-001 Add Project golden path: form opens, submits, project appears in list
    [Documentation]    Full user journey: click "Add Project", fill path (/tmp) and name,
    ...                submit, verify dialog closes and new project appears in the list.
    [Tags]    US046    smoke
    Open Add Project Dialog
    # Fill path with /tmp (guaranteed real directory in container)
    ${random}=    Generate Random String    6    [LETTERS]
    Fill Text    id=add-project-path    /tmp
    # Name should auto-fill with "tmp" (basename); override with a unique name
    Fill Text    id=add-project-name    E2E046 Project ${random}
    # Submit
    Click    role=dialog >> role=button >> text=/create|add|submit/i
    # Dialog should close
    Wait For Elements State    role=dialog    hidden    timeout=10s
    # New project should appear in dashboard list (look for the name we submitted)
    Wait For Elements State    text=E2E046 Project ${random}    visible    timeout=10s
    # Cleanup via HTTP delete (find project id from API)
    ${projects_resp}=    GET    ${API_BASE_URL}/api/v1/projects    expected_status=200
    ${projects}=    Set Variable    ${projects_resp.json()}[projects]
    FOR    ${p}    IN    @{projects}
        IF    '${p}[name]' == 'E2E046 Project ${random}'
            Delete Project Tool    ${SESSION_ID}    ${p}[id]
            BREAK
        END
    END

E2E-046-002 Add Project: server rejects non-existent path shows inline error
    [Documentation]    Types a path that does not exist on the server; verifies inline error
    ...                is shown and the dialog stays open.
    [Tags]    US046    regression
    Open Add Project Dialog
    Fill Text    id=add-project-path    /tmp/this-path-does-not-exist-e2e-us046-x99z
    # Name auto-fills; we don't need to touch it
    Click    role=dialog >> role=button >> text=/create|add|submit/i
    # Inline error should appear within the dialog
    Wait For Elements State
    ...    role=dialog >> text=/not a directory|does not exist/i
    ...    visible    timeout=5s
    # Dialog must remain open (not closed)
    Get Element States    role=dialog    contains    visible

E2E-046-003 Add Project: duplicate path shows DUPLICATE_PATH inline error
    [Documentation]    Creates a fixture project with path=/e2e/us046-fixture, then submits the form
    ...                with the same path and expects the 409 DUPLICATE_PATH error inline.
    [Tags]    US046    regression
    [Setup]    Create Dupe Fixture Project
    [Teardown]    Delete Dupe Fixture Project
    Open Add Project Dialog
    Fill Text    id=add-project-path    /e2e/us046-fixture
    Click    role=dialog >> role=button >> text=/create|add|submit/i
    # The 409 DUPLICATE_PATH error message should appear inline
    Wait For Elements State
    ...    role=dialog >> text=/already linked/i
    ...    visible    timeout=5s
    # Dialog still open
    Get Element States    role=dialog    contains    visible
