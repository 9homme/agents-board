*** Settings ***
Documentation    E2E tests for US047 — Requirement-level navigation on the project detail page.
...
...              E2E-047-001: Project detail page shows linked path in header and requirements list.
...              E2E-047-002: Clicking a requirement scopes the user stories tab.
...              E2E-047-003: New project with no requirements shows empty-state message.
Library          Browser
Library          RequestsLibrary
Library          Collections
Library          String
Resource         ../REQ001_agent_board_mcp/mcp_keywords.resource

Suite Setup      Setup US047 Suite
Suite Teardown   Teardown US047 Suite

*** Variables ***
${WEB_BASE_URL}       %{WEB_BASE_URL=http://localhost:3000}
${API_BASE_URL}       %{API_BASE_URL=http://localhost:8080}
${SESSION_ID}         ${EMPTY}
${PROJ_ID}            ${EMPTY}
${REQ_ID}             ${EMPTY}
${EMPTY_PROJ_ID}      ${EMPTY}

*** Keywords ***
Setup US047 Suite
    [Documentation]    Creates a project + requirement + user story for navigation tests.
    ${session_id}=    Connect To MCP SSE
    Set Suite Variable    ${SESSION_ID}    ${session_id}
    ${random}=    Generate Random String    8    [LETTERS]

    # Create project with path=/e2e/us047-proj (volume-mounted stub dir)
    ${proj_body}=    Create Dictionary    name=US047 Project ${random}    path=/e2e/us047-proj
    ${proj_resp}=    POST    ${API_BASE_URL}/api/v1/projects
    ...    json=${proj_body}
    ...    expected_status=201
    ${proj_id}=    Set Variable    ${proj_resp.json()}[id]
    Set Suite Variable    ${PROJ_ID}    ${proj_id}

    # Create requirement via MCP
    ${req_args}=    Create Dictionary    project_id=${proj_id}    name=US047 REQ ${random}
    ${req_resp}=    Call MCP Tool    ${SESSION_ID}    create_requirement    ${req_args}
    ${req_text}=    Set Variable    ${req_resp.json()['result']['content'][0]['text']}
    ${req_content}=    Evaluate    json.loads($req_text)    json
    Set Suite Variable    ${REQ_ID}    ${req_content}[id]

    # Create user story under the requirement via MCP
    ${story_args}=    Create Dictionary
    ...    projectId=${proj_id}
    ...    requirement_id=${req_content}[id]
    ...    title=US047 Story ${random}
    ...    description=Story for e2e navigation test
    ...    status=draft
    Call MCP Tool    ${SESSION_ID}    create_user_story    ${story_args}

    # Create a second project with NO requirements (freshly created, post-migration)
    ${random2}=    Generate Random String    8    [LETTERS]
    ${empty_body}=    Create Dictionary    name=US047 Empty ${random2}    path=/e2e/us047-empty
    ${empty_proj_resp}=    POST    ${API_BASE_URL}/api/v1/projects
    ...    json=${empty_body}
    ...    expected_status=201
    Set Suite Variable    ${EMPTY_PROJ_ID}    ${empty_proj_resp.json()}[id]

    New Browser    headless=True
    New Context

Teardown US047 Suite
    Close Browser
    Run Keyword If    '${PROJ_ID}' != '${EMPTY}'
    ...    Delete Project Tool    ${SESSION_ID}    ${PROJ_ID}
    Run Keyword If    '${EMPTY_PROJ_ID}' != '${EMPTY}'
    ...    Delete Project Tool    ${SESSION_ID}    ${EMPTY_PROJ_ID}

*** Test Cases ***

E2E-047-001 Project detail page shows linked path in header and requirements list
    [Documentation]    Opens a real project detail page; asserts the linked path (/tmp)
    ...                appears in the header and the requirements area has at least one item.
    [Tags]    US047    smoke
    New Page    ${WEB_BASE_URL}/projects/${PROJ_ID}
    # Wait for the page to load (project header visible)
    Wait For Elements State    text=/US047 Project/i    visible    timeout=10s
    # Assert path is shown in the header (substring match avoids regex slash ambiguity)
    Wait For Elements State    text=us047-proj    visible    timeout=5s
    # Assert requirements list has at least one item
    Wait For Elements State    text=/US047 REQ/i    visible    timeout=5s

E2E-047-002 Clicking a requirement scopes the user stories tab
    [Documentation]    Clicks on the requirement and verifies the user story created under it
    ...                appears in the user stories list/tab.
    [Tags]    US047    regression
    New Page    ${WEB_BASE_URL}/projects/${PROJ_ID}
    Wait For Elements State    text=/US047 REQ/i    visible    timeout=10s
    # Click the requirement to select it
    Click    text=/US047 REQ/i
    # Navigate to user stories tab if not auto-shown
    ${tab_visible}=    Run Keyword And Return Status
    ...    Wait For Elements State    role=tab >> text=/user stor/i    visible    timeout=3s
    Run Keyword If    ${tab_visible}    Click    role=tab >> text=/user stor/i
    # The user story created in suite setup should be visible
    Wait For Elements State    text=/US047 Story/i    visible    timeout=10s

E2E-047-003 New project with no requirements shows empty-state message
    [Documentation]    Freshly created project (no requirements) should display
    ...                "No requirements yet" or equivalent empty-state in the requirements area.
    [Tags]    US047    regression
    New Page    ${WEB_BASE_URL}/projects/${EMPTY_PROJ_ID}
    # Wait for page to load
    Wait For Elements State    text=/US047 Empty/i    visible    timeout=10s
    # Empty state for requirements area
    Wait For Elements State    text=/no requirements/i    visible    timeout=5s
    # Project header and path should still render (substring match avoids regex slash ambiguity)
    Wait For Elements State    text=us047-empty    visible    timeout=3s
