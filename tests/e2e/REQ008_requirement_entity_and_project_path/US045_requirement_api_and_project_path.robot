*** Settings ***
Documentation    E2E tests for US045 — Requirement read API + project path on create.
...
...              E2E-045-001: POST /api/v1/projects with real directory → 201.
...              E2E-045-002: POST /api/v1/projects with missing path field → 400.
...              E2E-045-003: POST /api/v1/projects with non-existent path → 400.
...              E2E-045-004: POST /api/v1/projects with duplicate path → 409.
...              E2E-045-005: GET /api/v1/projects/:pid/requirements returns Default requirement.
...              E2E-045-006: GET /api/v1/projects/unknown/requirements for unknown project → 404.
...              E2E-045-007: MCP create_user_story with requirement_id succeeds (BREAKING CHANGE).
...              E2E-045-008: MCP create_user_story without requirement_id returns tool error.
Library          RequestsLibrary
Library          Collections
Library          String
Resource         ../REQ001_agent_board_mcp/mcp_keywords.resource

Suite Setup      Connect To MCP And Create Fixture Project
Suite Teardown   Cleanup Fixture Project

*** Variables ***
${API_BASE_URL}     %{API_BASE_URL=http://localhost:8080}
${WEB_BASE_URL}     %{WEB_BASE_URL=http://localhost:3000}
${FIXTURE_PROJ_ID}    ${EMPTY}
${FIXTURE_REQ_ID}     ${EMPTY}
${SESSION_ID}         ${EMPTY}

*** Keywords ***
Connect To MCP And Create Fixture Project
    [Documentation]    Creates an MCP session and a fixture project+requirement for tests that need them.
    ${session_id}=    Connect To MCP SSE
    Set Suite Variable    ${SESSION_ID}    ${session_id}
    # Create a project with path=/e2e/us045-fixture (volume-mounted stub dir)
    ${random}=    Generate Random String    8    [LETTERS]
    ${body}=    Create Dictionary    name=E2E045 Project ${random}    description=${EMPTY}    path=/e2e/us045-fixture
    ${proj_resp}=    POST    ${API_BASE_URL}/api/v1/projects
    ...    json=${body}
    ...    expected_status=201
    ${proj_id}=    Set Variable    ${proj_resp.json()}[id]
    Set Suite Variable    ${FIXTURE_PROJ_ID}    ${proj_id}
    # Create a requirement via MCP so create_user_story tests have a requirement_id
    ${req_args}=    Create Dictionary    project_id=${proj_id}    name=E2E REQ ${random}
    ${req_resp}=    Call MCP Tool    ${SESSION_ID}    create_requirement    ${req_args}
    ${req_text}=    Set Variable    ${req_resp.json()['result']['content'][0]['text']}
    ${req_content}=    Evaluate    json.loads($req_text)    json
    Set Suite Variable    ${FIXTURE_REQ_ID}    ${req_content}[id]

Cleanup Fixture Project
    [Documentation]    Deletes the fixture project via MCP (cascades to children).
    Run Keyword If    '${FIXTURE_PROJ_ID}' != '${EMPTY}'
    ...    Delete Project Tool    ${SESSION_ID}    ${FIXTURE_PROJ_ID}

Create Requirement Tool
    [Arguments]    ${session_id}    ${project_id}    ${name}    ${description}=${EMPTY}    ${status}=draft
    ${args}=    Create Dictionary    project_id=${project_id}    name=${name}
    Run Keyword If    '${description}' != '${EMPTY}'    Set To Dictionary    ${args}    description=${description}
    Run Keyword If    '${status}' != 'draft'    Set To Dictionary    ${args}    status=${status}
    ${resp}=    Call MCP Tool    ${session_id}    create_requirement    ${args}
    RETURN    ${resp}

Create User Story With Requirement Tool
    [Arguments]    ${session_id}    ${project_id}    ${requirement_id}    ${title}    ${description}=${EMPTY}
    ${args}=    Create Dictionary
    ...    projectId=${project_id}
    ...    requirement_id=${requirement_id}
    ...    title=${title}
    ...    description=${description}
    ...    status=draft
    ${resp}=    Call MCP Tool    ${session_id}    create_user_story    ${args}
    RETURN    ${resp}

*** Test Cases ***

E2E-045-001 POST /api/v1/projects with real directory returns 201
    [Documentation]    Verifies that POSTing a project with a real directory path
    ...                (os.Stat passes) returns 201 with the exact response shape.
    [Tags]    US045    smoke
    ${random}=    Generate Random String    8    [LETTERS]
    # /e2e/us045-test-001 is a volume-mounted stub directory that always exists
    ${body}=    Create Dictionary    name=E2E-045-001 ${random}    description=test project    path=/e2e/us045-test-001
    ${response}=    POST    ${API_BASE_URL}/api/v1/projects
    ...    json=${body}
    ...    expected_status=201
    ${resp_body}=    Set Variable    ${response.json()}
    Dictionary Should Contain Key    ${resp_body}    id
    Dictionary Should Contain Key    ${resp_body}    name
    Dictionary Should Contain Key    ${resp_body}    path
    Dictionary Should Contain Key    ${resp_body}    createdAt
    Dictionary Should Contain Key    ${resp_body}    updatedAt
    Should Be Equal    ${resp_body}[path]    /e2e/us045-test-001
    Should Not Be Empty    ${resp_body}[id]
    # Cleanup: delete via MCP
    Delete Project Tool    ${SESSION_ID}    ${resp_body}[id]

E2E-045-002 POST /api/v1/projects with missing path returns 400
    [Documentation]    Verifies that omitting the path field returns 400 VALIDATION_ERROR.
    [Tags]    US045    regression
    ${body}=    Create Dictionary    name=No Path Project
    ${response}=    POST    ${API_BASE_URL}/api/v1/projects
    ...    json=${body}
    ...    expected_status=400
    ${resp_body}=    Set Variable    ${response.json()}
    Should Be Equal    ${resp_body}[code]    VALIDATION_ERROR
    Should Contain    ${resp_body}[message]    path is required

E2E-045-003 POST /api/v1/projects with non-existent path returns 400
    [Documentation]    Verifies that os.Stat validation rejects a path that does not exist.
    [Tags]    US045    regression
    ${body}=    Create Dictionary    name=Bad Path    path=/tmp/definitely-does-not-exist-e2e-x99z-us045
    ${response}=    POST    ${API_BASE_URL}/api/v1/projects
    ...    json=${body}
    ...    expected_status=400
    ${resp_body}=    Set Variable    ${response.json()}
    Should Be Equal    ${resp_body}[code]    VALIDATION_ERROR
    Should Contain    ${resp_body}[message]    not a directory

E2E-045-004 POST /api/v1/projects with duplicate path returns 409
    [Documentation]    Verifies that a path already linked to another project returns 409 DUPLICATE_PATH.
    ...                Uses /e2e/us045-fixture which was linked by the fixture project (FIXTURE_PROJ_ID).
    [Tags]    US045    regression
    ${body}=    Create Dictionary    name=Duplicate Path    path=/e2e/us045-fixture
    ${response}=    POST    ${API_BASE_URL}/api/v1/projects
    ...    json=${body}
    ...    expected_status=409
    ${resp_body}=    Set Variable    ${response.json()}
    Should Be Equal    ${resp_body}[code]    DUPLICATE_PATH
    Should Contain    ${resp_body}[message]    already linked

E2E-045-005 GET /api/v1/projects/:pid/requirements returns requirements list
    [Documentation]    Verifies the list endpoint returns requirements in the correct shape.
    ...                The fixture project has one requirement created in suite setup.
    [Tags]    US045    smoke
    ${response}=    GET
    ...    ${API_BASE_URL}/api/v1/projects/${FIXTURE_PROJ_ID}/requirements
    ...    expected_status=200
    ${body}=    Set Variable    ${response.json()}
    Dictionary Should Contain Key    ${body}    requirements
    Should Not Be Empty    ${body}[requirements]
    ${first_req}=    Get From List    ${body}[requirements]    0
    Dictionary Should Contain Key    ${first_req}    id
    Dictionary Should Contain Key    ${first_req}    projectId
    Dictionary Should Contain Key    ${first_req}    name
    Dictionary Should Contain Key    ${first_req}    description
    Dictionary Should Contain Key    ${first_req}    status
    Dictionary Should Contain Key    ${first_req}    createdAt
    Dictionary Should Contain Key    ${first_req}    updatedAt
    Should Be Equal    ${first_req}[projectId]    ${FIXTURE_PROJ_ID}

E2E-045-006 GET /api/v1/projects/unknown/requirements returns 404
    [Documentation]    Verifies that a non-existent project returns 404 NOT_FOUND.
    [Tags]    US045    regression
    ${response}=    GET
    ...    ${API_BASE_URL}/api/v1/projects/00000000-0000-0000-0000-000000000000/requirements
    ...    expected_status=404
    ${body}=    Set Variable    ${response.json()}
    Should Be Equal    ${body}[code]    NOT_FOUND
    Should Contain    ${body}[message]    Project not found

E2E-045-007 MCP create_user_story with requirement_id succeeds (BREAKING CHANGE)
    [Documentation]    Verifies that the updated create_user_story tool accepts requirement_id
    ...                and creates the story. Then fetches via the hierarchy endpoint to confirm.
    [Tags]    US045    smoke
    ${story_resp}=    Create User Story With Requirement Tool
    ...    ${SESSION_ID}    ${FIXTURE_PROJ_ID}    ${FIXTURE_REQ_ID}    E2E-045-007 Story
    ${story_text}=    Set Variable    ${story_resp.json()['result']['content'][0]['text']}
    ${story_content}=    Evaluate    json.loads($story_text)    json
    # Assert the story has requirementId (camelCase in API response)
    Dictionary Should Contain Key    ${story_content}    requirementId
    Should Be Equal    ${story_content}[requirementId]    ${FIXTURE_REQ_ID}
    Should Be Equal    ${story_content}[projectId]    ${FIXTURE_PROJ_ID}
    # Verify via HTTP hierarchy endpoint
    ${stories_resp}=    GET
    ...    ${API_BASE_URL}/api/v1/projects/${FIXTURE_PROJ_ID}/requirements/${FIXTURE_REQ_ID}/user-stories
    ...    expected_status=200
    ${stories_body}=    Set Variable    ${stories_resp.json()}
    Should Not Be Empty    ${stories_body}[userStories]

E2E-045-008 MCP create_user_story without requirement_id returns tool error
    [Documentation]    Verifies that omitting requirement_id from create_user_story returns a tool error.
    [Tags]    US045    regression
    ${args}=    Create Dictionary
    ...    projectId=${FIXTURE_PROJ_ID}
    ...    title=No Req Story
    ...    description=should fail
    ...    status=draft
    # Intentionally omit requirement_id — this should return a tool error
    ${resp}=    Call MCP Tool    ${SESSION_ID}    create_user_story    ${args}
    ${result}=    Set Variable    ${resp.json()['result']}
    # The tool should return isError: true in the result
    Dictionary Should Contain Key    ${result}    isError
    Should Be True    ${result}[isError]
