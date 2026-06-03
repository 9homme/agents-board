*** Settings ***
Documentation    E2E tests for US001 — Navigate to project detail page with tabs.
...              Verifies: dashboard card click-through to /projects/{id}; tab switching
...              with URL persistence; browser refresh preserves active tab.
Library          Browser
Library          String
Resource         ../REQ001_agent_board_mcp/mcp_keywords.resource
Resource         resources/project_detail_keywords.resource

Suite Setup      Setup US001 Suite
Suite Teardown   Close Browser

*** Variables ***
${WEB_BASE_URL}    http://localhost:3000
${PROJECT_ID}      ${EMPTY}
${PROJECT_NAME}    ${EMPTY}

*** Keywords ***
URL Should Contain
    [Arguments]    ${fragment}
    ${url}=    Get Url
    Should Contain    ${url}    ${fragment}

URL Should Not Contain
    [Arguments]    ${fragment}
    ${url}=    Get Url
    Should Not Contain    ${url}    ${fragment}


Setup US001 Suite
    [Documentation]    Creates a test project via MCP and opens a browser.
    ${random}=         Generate Random String    8    [LETTERS]
    ${name}=           Set Variable    REQ004 US001 E2E ${random}
    Set Suite Variable    ${PROJECT_NAME}    ${name}
    ${session_id}=     Connect To MCP SSE
    ${resp}=           Create Project Tool    ${session_id}    ${name}    E2E test project description
    ${resp_text}=      Set Variable    ${resp.json()['result']['content'][0]['text']}
    ${content}=        Evaluate    json.loads($resp_text)    json
    Set Suite Variable    ${PROJECT_ID}    ${content['id']}
    New Browser        headless=True
    New Page           ${WEB_BASE_URL}/

*** Test Cases ***
E2E-US001-001 Dashboard click-through to detail page then tab switch
    [Documentation]    Verifies: click project card navigates to /projects/{id}; project name
    ...                in heading; two tabs visible; Documents active by default; User Stories
    ...                tab shows verbatim placeholder; clicking Documents restores it.
    [Tags]    US001    smoke    regression

    # Step 1: open the dashboard and wait for the card
    New Page    ${WEB_BASE_URL}/
    Wait For Elements State    text="${PROJECT_NAME}"    visible    timeout=15s

    # Step 2: click the project card (it is a <Link> wrapping the card content)
    Click    text="${PROJECT_NAME}"

    # Step 3: wait for a DETAIL-PAGE-ONLY element before asserting URL.
    # Don't wait for `role=heading >> text=${PROJECT_NAME}` — that matches the
    # dashboard ProjectCard's h3 too and succeeds without navigation. Tabs
    # exist only on the detail page; wait for them as the navigation marker.
    Wait For Elements State    role=tab >> text=Documents    visible    timeout=15s
    Wait Until Keyword Succeeds    10s    200ms    URL Should Contain    /projects/${PROJECT_ID}
    ${url}=    Get Url
    Should Contain    ${url}    /projects/${PROJECT_ID}
    ${heading}=    Get Text    css=h1
    Should Contain    ${heading}    ${PROJECT_NAME}

    # Step 5: two tabs visible
    Wait For Elements State    role=tab >> text=Documents    visible    timeout=10s
    ${tabs}=    Get Elements    role=tab
    Length Should Be    ${tabs}    2

    # Step 6: Documents tab is active by default
    ${docs_tab}=    Get Element    role=tab >> text=Documents
    ${aria_selected}=    Get Attribute    ${docs_tab}    aria-selected
    Should Be Equal    ${aria_selected}    true

    # Step 7: switch to User Stories tab
    Click    role=tab >> text="User Stories"
    Wait Until Keyword Succeeds    10s    200ms    URL Should Contain    tab=user-stories
    ${url_after}=    Get Url
    Should Contain    ${url_after}    tab=user-stories

    # Step 8: verbatim placeholder text
    Wait For Elements State
    ...    text="Coming soon — user stories will appear here in a future release."
    ...    visible    timeout=10s

    # Step 9: switch back to Documents tab
    Click    role=tab >> text=Documents
    Wait Until Keyword Succeeds    10s    200ms    URL Should Not Contain    tab=user-stories
    ${url_docs}=    Get Url
    Should Not Contain    ${url_docs}    tab=user-stories

E2E-US001-002 Direct URL with tab=user-stories survives browser refresh
    [Documentation]    Verifies that ?tab=user-stories in the URL is preserved across a real
    ...                browser page reload (URL-as-source-of-truth).
    [Tags]    US001    regression

    # Navigate directly with tab=user-stories
    New Page    ${WEB_BASE_URL}/projects/${PROJECT_ID}?tab=user-stories
    Wait For Elements State
    ...    text="Coming soon — user stories will appear here in a future release."
    ...    visible    timeout=15s

    # Confirm User Stories tab is active
    ${us_tab}=    Get Element    role=tab >> text="User Stories"
    ${aria_before}=    Get Attribute    ${us_tab}    aria-selected
    Should Be Equal    ${aria_before}    true

    # Reload the page
    Reload

    # After reload: User Stories tab still active
    Wait For Elements State
    ...    text="Coming soon — user stories will appear here in a future release."
    ...    visible    timeout=15s
    ${us_tab_after}=    Get Element    role=tab >> text="User Stories"
    ${aria_after}=    Get Attribute    ${us_tab_after}    aria-selected
    Should Be Equal    ${aria_after}    true
