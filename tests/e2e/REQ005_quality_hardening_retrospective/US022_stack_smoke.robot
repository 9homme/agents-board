*** Settings ***
Documentation    E2E tests for US022 — Live e2e stack-up.
...
...              E2E-US022-001: API server is healthy and returns project list (RequestsLibrary).
...              E2E-US022-002: Web UI is reachable and renders the dashboard (Browser library).
...
...              E2E-US022-003 (documented in US022_e2e_tests.md) is NOT a Robot test case —
...              it is a make-level invocation whose outcome is captured in the test report.
...              The orchestrator runs `make e2e-run REQ=REQ001` (or similar) and records the
...              exit code. That step is outside this suite.
...
...              Preconditions (NOT steps — must be satisfied before running):
...                - `make e2e-up && make e2e-seed` completed successfully.
...                - Postgres (127.0.0.1:15432), api-server (127.0.0.1:8080), and web
...                  (127.0.0.1:3000) containers are healthy per compose healthchecks.
...                - `robot` is on PATH with robotframework-browser and
...                  robotframework-requests installed.
...
...              Architecture cite: §6.2 (service topology), §6.3 (Makefile targets), §6.6 (Robot
...              invocation pattern).
Library          RequestsLibrary
Library          Browser
Resource         resources/req005_keywords.resource

Suite Setup      Create API Session
Suite Teardown   Delete All Sessions

*** Variables ***
${API_BASE_URL}    http://localhost:8080
${WEB_BASE_URL}    http://localhost:3000

*** Test Cases ***
E2E-US022-001 Stack smoke: API server is healthy and returns project list
    [Documentation]    Asserts that a GET /api/v1/projects returns HTTP 200 and a body that
    ...                contains the "projects" key. This mirrors the api-server healthcheck
    ...                that compose uses (architecture §6.2).
    [Tags]    US022    REQ005    smoke

    ${resp}=    GET On Session    api    /api/v1/projects    expected_status=200
    ${body}=    Set Variable    ${resp.json()}
    Dictionary Should Contain Key    ${body}    projects

E2E-US022-002 Stack smoke: web UI is reachable and renders the dashboard
    [Documentation]    Asserts that the Next.js web container is serving the dashboard page:
    ...                the <body> is visible and the page title is non-empty.
    ...                Architecture cite: §6.2 web service at port 3000; D-012 containerised
    ...                Next.js production build.
    [Tags]    US022    REQ005    smoke

    New Browser    chromium    headless=True
    New Page    ${WEB_BASE_URL}/
    Wait For Elements State    css=body    visible    timeout=10s
    ${title}=    Get Title
    Should Not Be Empty    ${title}
    Close Browser
