*** Settings ***
Documentation    E2E tests for US006 — Harmonise FE hooks on AbortController.
...
...              E2E-US006-001: Rapid project navigation does not show stale data or error state.
...
...              Why e2e: The AbortController refactor is internally correct (proven by
...              FCT-US006-003 through FCT-US006-010 with MSW-simulated delay). One scenario
...              belongs at e2e: on the live stack, rapidly navigating between two projects must
...              not produce stale data rendered under the wrong project URL nor surface any
...              visible error state. Real network latency and the Next.js Pages Router CSR
...              lifecycle interact with AbortController in ways that MSW mocks cannot fully
...              replicate. See US006_e2e_tests.md §"Why e2e" for the full rationale.
...
...              Preconditions: stack is UP via `make e2e-up && make e2e-seed`.
...              This suite creates its own isolated test projects via REST API so it is not
...              coupled to any particular seed file name. mcp-server is NOT required
...              (architecture §6.2 — mcp-server is not in the REQ005 compose stack).
Library          Browser
Library          RequestsLibrary
Library          String
Resource         resources/req005_keywords.resource
Resource         ../../REQ004_project_detail_page/resources/project_detail_keywords.resource

Suite Setup      Setup US006 Suite
Suite Teardown   Teardown US006 Suite

*** Variables ***
${WEB_BASE_URL}    http://localhost:3000
${API_BASE_URL}    http://localhost:8080
${PROJECT_1_ID}    ${EMPTY}
${PROJECT_1_NAME}    ${EMPTY}
${PROJECT_2_ID}    ${EMPTY}
${PROJECT_2_NAME}    ${EMPTY}

*** Keywords ***
Setup US006 Suite
    [Documentation]    Creates two distinct projects via REST API for the rapid-navigation test.
    ...                Opens a Browser instance in headless mode.
    Create API Session
    ${random}=         Generate Random String    8    [LETTERS]
    ${name1}=          Set Variable    REQ005 US006 Alpha ${random}
    ${name2}=          Set Variable    REQ005 US006 Beta ${random}
    Set Suite Variable    ${PROJECT_1_NAME}    ${name1}
    Set Suite Variable    ${PROJECT_2_NAME}    ${name2}
    ${proj1}=          Create Project Via API    ${name1}    First project for rapid-navigation test
    ${proj2}=          Create Project Via API    ${name2}    Second project for rapid-navigation test
    Set Suite Variable    ${PROJECT_1_ID}    ${proj1['id']}
    Set Suite Variable    ${PROJECT_2_ID}    ${proj2['id']}
    New Browser        chromium    headless=True

Teardown US006 Suite
    [Documentation]    Closes the browser. Stack teardown is done externally by `make e2e-down`.
    Close Browser

*** Test Cases ***
E2E-US006-001 Rapid project navigation does not show stale data or error state
    [Documentation]    Navigates to Project 1, then immediately navigates to Project 2 before
    ...                the first project's data has finished rendering. Asserts that the page
    ...                ultimately shows Project 2's name (no stale Project 1 data) and no
    ...                error state (AbortController swallows the cancelled in-flight fetch).
    ...
    ...                If network is fast enough that both fetches complete before the second
    ...                navigation fires, the test still has value: it proves the page is stable
    ...                after rapid navigation and the final state is correct. The unit tests
    ...                (FCT-US006-004/006) are the authoritative abort-correctness proofs.
    [Tags]    US006    REQ005    regression

    # Step 1 — open dashboard and wait for the project list to render
    New Page    ${WEB_BASE_URL}/
    # Multiple project-card elements present (we created at least 2 plus seed);
    # nth=0 disambiguates for Playwright strict mode while still asserting that
    # the dashboard rendered at least one card.
    Wait For Elements State    css=[data-testid="project-card"] >> nth=0    visible    timeout=15s

    # Step 2 — navigate to Project 1
    Go To    ${WEB_BASE_URL}/projects/${PROJECT_1_ID}

    # Step 3 — immediately navigate to Project 2 (simulate rapid navigation;
    #           intentionally NO wait between the two Go To calls so the first
    #           fetch may still be in-flight when the second starts)
    Go To    ${WEB_BASE_URL}/projects/${PROJECT_2_ID}

    # Step 4 — wait for Project 2's heading to become visible
    Wait For Elements State    css=h1    visible    timeout=10s

    # Step 5 — assert the heading text is Project 2's name (no stale Project 1 data)
    ${heading}=    Get Text    css=h1
    Should Contain    ${heading}    ${PROJECT_2_NAME}
    Should Not Contain    ${heading}    ${PROJECT_1_NAME}

    # Step 6 — assert no error state is visible
    Assert No Error State Visible

    # Step 7 — assert we are on the correct URL
    ${url}=    Get Url
    Should Contain    ${url}    /projects/${PROJECT_2_ID}
