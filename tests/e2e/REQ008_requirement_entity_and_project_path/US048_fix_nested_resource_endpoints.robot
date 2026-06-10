*** Settings ***
Documentation    E2E tests for US048 — Migrate to fully-nested REST hierarchy.
...
...              E2E-048-001: Full hierarchy leaf endpoints return 200 (golden path).
...              E2E-048-002: Chain mismatch returns 404 — no cross-resource leakage.
...              E2E-048-003: All 8 removed flat routes return non-200.
Library          RequestsLibrary
Library          Collections
Library          String
Resource         ../REQ001_agent_board_mcp/mcp_keywords.resource

Suite Setup      Setup US048 Suite
Suite Teardown   Teardown US048 Suite

*** Variables ***
${API_BASE_URL}     %{API_BASE_URL=http://localhost:8080}
${SESSION_ID}       ${EMPTY}
${PROJ_P}           ${EMPTY}
${PROJ_P2}          ${EMPTY}
${REQ_R}            ${EMPTY}
${REQ_R2}           ${EMPTY}
${STORY_S}          ${EMPTY}
${TASK_T}           ${EMPTY}
${DOC_D}            ${EMPTY}

*** Keywords ***
Setup US048 Suite
    [Documentation]    Creates: Project P, Requirement R (in P), UserStory S (in R),
    ...                Task T (in S), Document D (in R). Also creates P2 + R2 (in P2)
    ...                for chain-mismatch test.
    ${session_id}=    Connect To MCP SSE
    Set Suite Variable    ${SESSION_ID}    ${session_id}
    ${random}=    Generate Random String    8    [LETTERS]

    # Project P with path=/tmp
    ${p_resp}=    POST    ${API_BASE_URL}/api/v1/projects
    ...    json={"name": "US048 P ${random}", "path": "/tmp"}
    ...    expected_status=201
    Set Suite Variable    ${PROJ_P}    ${p_resp.json()}[id]

    # Requirement R in P
    ${r_args}=    Create Dictionary    project_id=${PROJ_P}    name=US048 R ${random}
    ${r_resp}=    Call MCP Tool    ${SESSION_ID}    create_requirement    ${r_args}
    ${r_text}=    Set Variable    ${r_resp.json()['result']['content'][0]['text']}
    ${r_content}=    Evaluate    json.loads($r_text)    json
    Set Suite Variable    ${REQ_R}    ${r_content}[id]

    # UserStory S in R
    ${s_args}=    Create Dictionary
    ...    projectId=${PROJ_P}
    ...    requirementId=${REQ_R}
    ...    title=US048 Story ${random}
    ...    description=hierarchy e2e story
    ...    status=draft
    ${s_resp}=    Call MCP Tool    ${SESSION_ID}    create_user_story    ${s_args}
    ${s_text}=    Set Variable    ${s_resp.json()['result']['content'][0]['text']}
    ${s_content}=    Evaluate    json.loads($s_text)    json
    Set Suite Variable    ${STORY_S}    ${s_content}[id]

    # Task T in S
    ${t_args}=    Create Dictionary
    ...    userStoryId=${STORY_S}
    ...    title=US048 Task ${random}
    ...    description=hierarchy e2e task
    ...    status=pending
    ${t_resp}=    Call MCP Tool    ${SESSION_ID}    create_task    ${t_args}
    ${t_text}=    Set Variable    ${t_resp.json()['result']['content'][0]['text']}
    ${t_content}=    Evaluate    json.loads($t_text)    json
    Set Suite Variable    ${TASK_T}    ${t_content}[id]

    # Document D in R
    ${d_args}=    Create Dictionary
    ...    projectId=${PROJ_P}
    ...    requirementId=${REQ_R}
    ...    title=US048 Doc ${random}
    ...    content=# US048 E2E Document
    ${d_resp}=    Call MCP Tool    ${SESSION_ID}    create_document    ${d_args}
    ${d_text}=    Set Variable    ${d_resp.json()['result']['content'][0]['text']}
    ${d_content}=    Evaluate    json.loads($d_text)    json
    Set Suite Variable    ${DOC_D}    ${d_content}[id]

    # Project P2 for chain-mismatch test, path=/var/tmp
    ${p2_resp}=    POST    ${API_BASE_URL}/api/v1/projects
    ...    json={"name": "US048 P2 ${random}", "path": "/var/tmp"}
    ...    expected_status=201
    Set Suite Variable    ${PROJ_P2}    ${p2_resp.json()}[id]

    # Requirement R2 in P2 (used to prove cross-project mismatch → 404)
    ${r2_args}=    Create Dictionary    project_id=${PROJ_P2}    name=US048 R2 ${random}
    ${r2_resp}=    Call MCP Tool    ${SESSION_ID}    create_requirement    ${r2_args}
    ${r2_text}=    Set Variable    ${r2_resp.json()['result']['content'][0]['text']}
    ${r2_content}=    Evaluate    json.loads($r2_text)    json
    Set Suite Variable    ${REQ_R2}    ${r2_content}[id]

Teardown US048 Suite
    Run Keyword If    '${PROJ_P}' != '${EMPTY}'
    ...    Delete Project Tool    ${SESSION_ID}    ${PROJ_P}
    Run Keyword If    '${PROJ_P2}' != '${EMPTY}'
    ...    Delete Project Tool    ${SESSION_ID}    ${PROJ_P2}

Assert Response Is Not 200
    [Arguments]    ${path}
    ${response}=    GET    ${API_BASE_URL}${path}    expected_status=any
    Should Not Be Equal As Integers    ${response.status_code}    200
    ...    msg=Expected non-200 for removed route ${path} but got 200

*** Test Cases ***

E2E-048-001 Full hierarchy leaf endpoints return 200 (golden path)
    [Documentation]    Verifies each new canonical hierarchy endpoint returns 200 with the correct shape:
    ...                user-story detail, task detail, document detail.
    [Tags]    US048    smoke

    # Sub-case A: user-story detail
    ${us_resp}=    GET
    ...    ${API_BASE_URL}/api/v1/projects/${PROJ_P}/requirements/${REQ_R}/user-stories/${STORY_S}
    ...    expected_status=200
    ${us_body}=    Set Variable    ${us_resp.json()}
    Should Be Equal    ${us_body}[id]    ${STORY_S}
    Should Be Equal    ${us_body}[projectId]    ${PROJ_P}
    Should Be Equal    ${us_body}[requirementId]    ${REQ_R}
    Dictionary Should Not Contain Key    ${us_body}    taskCount    msg=Detail endpoint must NOT include taskCount

    # Sub-case B: task detail
    ${task_resp}=    GET
    ...    ${API_BASE_URL}/api/v1/projects/${PROJ_P}/requirements/${REQ_R}/user-stories/${STORY_S}/tasks/${TASK_T}
    ...    expected_status=200
    ${task_body}=    Set Variable    ${task_resp.json()}
    Should Be Equal    ${task_body}[id]    ${TASK_T}
    Should Be Equal    ${task_body}[userStoryId]    ${STORY_S}
    Dictionary Should Not Contain Key    ${task_body}    requirementId    msg=Task shape must not include requirementId

    # Sub-case C: document detail (incl. content)
    ${doc_resp}=    GET
    ...    ${API_BASE_URL}/api/v1/projects/${PROJ_P}/requirements/${REQ_R}/documents/${DOC_D}
    ...    expected_status=200
    ${doc_body}=    Set Variable    ${doc_resp.json()}
    Should Be Equal    ${doc_body}[id]    ${DOC_D}
    Should Be Equal    ${doc_body}[projectId]    ${PROJ_P}
    Should Be Equal    ${doc_body}[requirementId]    ${REQ_R}
    Dictionary Should Contain Key    ${doc_body}    content
    Should Not Be Empty    ${doc_body}[content]

E2E-048-002 Chain mismatch returns 404 (no cross-resource leakage)
    [Documentation]    Uses REQ_R (which belongs to PROJ_P) but passes PROJ_P2 as :pid.
    ...                Verifies the server returns 404 NOT_FOUND and does not leak R's user stories.
    [Tags]    US048    regression
    # R belongs to P, not P2 — this is a cross-project mismatch
    ${response}=    GET
    ...    ${API_BASE_URL}/api/v1/projects/${PROJ_P2}/requirements/${REQ_R}/user-stories
    ...    expected_status=404
    ${body}=    Set Variable    ${response.json()}
    Should Be Equal    ${body}[code]    NOT_FOUND
    Should Contain    ${body}[message]    not found

E2E-048-003 All 8 removed flat routes return non-200
    [Documentation]    Asserts each of the 8 routes removed by US048 no longer resolves
    ...                (the router returns 404 or 405, NOT 200 with data).
    [Tags]    US048    regression
    # 1. GET /api/v1/projects/:id/user-stories (flat project-scoped list)
    Assert Response Is Not 200    /api/v1/projects/${PROJ_P}/user-stories

    # 2. GET /api/v1/projects/:id/documents (flat project-scoped list)
    Assert Response Is Not 200    /api/v1/projects/${PROJ_P}/documents

    # 3. GET /api/v1/user-stories/:id (top-level detail)
    Assert Response Is Not 200    /api/v1/user-stories/${STORY_S}

    # 4. GET /api/v1/user-stories/:id/tasks (flat sub-resource)
    Assert Response Is Not 200    /api/v1/user-stories/${STORY_S}/tasks

    # 5. GET /api/v1/tasks/:id (top-level task detail)
    Assert Response Is Not 200    /api/v1/tasks/${TASK_T}

    # 6. GET /api/v1/documents/:id (top-level document detail)
    Assert Response Is Not 200    /api/v1/documents/${DOC_D}

    # 7. GET /api/v1/requirements/:rid/user-stories (intermediate draft route)
    Assert Response Is Not 200    /api/v1/requirements/${REQ_R}/user-stories

    # 8. GET /api/v1/requirements/:rid/documents (intermediate draft route)
    Assert Response Is Not 200    /api/v1/requirements/${REQ_R}/documents
