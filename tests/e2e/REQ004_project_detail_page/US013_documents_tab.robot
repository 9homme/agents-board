*** Settings ***
Documentation    E2E tests for US013 — Documents tab: list documents and view content.
...              Verifies: auto-select first document; click sidebar item updates previewer
...              and URL; browser refresh with ?doc= rehydrates same document.
Library          Browser
Library          String
Resource         ../REQ001_agent_board_mcp/mcp_keywords.resource
Resource         resources/project_detail_keywords.resource

Suite Setup      Setup US013 Suite
Suite Teardown   Close Browser

*** Keywords ***
URL Should Contain
    [Arguments]    ${fragment}
    ${url}=    Get Url
    Should Contain    ${url}    ${fragment}

*** Variables ***
${WEB_BASE_URL}    http://localhost:3000
${PROJECT_ID}      ${EMPTY}
${DOC1_ID}         ${EMPTY}
${DOC2_ID}         ${EMPTY}

*** Keywords ***
Setup US013 Suite
    [Documentation]    Creates a test project with two documents via MCP.
    ...                Doc1 is created first, Doc2 second — so Doc2 has a later updated_at
    ...                and should appear first in the sidebar (updatedAt DESC).
    ${random}=         Generate Random String    8    [LETTERS]
    ${project_name}=   Set Variable    REQ004 US013 E2E ${random}
    ${session_id}=     Connect To MCP SSE

    ${proj_resp}=    Create Project Tool    ${session_id}    ${project_name}    US013 E2E test project
    # Pass the text via `$var` (Python variable reference) NOT interpolation —
    # the doc content contains literal \n which Python would interpret as newline,
    # breaking JSON validity if used inside triple-quoted Evaluate expressions.
    ${proj_text}=    Set Variable    ${proj_resp.json()['result']['content'][0]['text']}
    ${proj_content}=    Evaluate    json.loads($proj_text)    json
    Set Suite Variable    ${PROJECT_ID}    ${proj_content['id']}

    # Create Doc1 first (older updated_at)
    ${doc1_content_body}=    Set Variable    \# First\n\nHello from doc 1.
    ${doc1_resp}=    Create Document Tool    ${session_id}    ${PROJECT_ID}    First Document    ${doc1_content_body}
    ${doc1_text}=    Set Variable    ${doc1_resp.json()['result']['content'][0]['text']}
    ${doc1_content}=    Evaluate    json.loads($doc1_text)    json
    Set Suite Variable    ${DOC1_ID}    ${doc1_content['id']}

    # Create Doc2 second (newer updated_at — will be listed first by the API)
    ${doc2_content_body}=    Set Variable    \# Second\n\nHello from doc 2.
    ${doc2_resp}=    Create Document Tool    ${session_id}    ${PROJECT_ID}    Second Document    ${doc2_content_body}
    ${doc2_text}=    Set Variable    ${doc2_resp.json()['result']['content'][0]['text']}
    ${doc2_content}=    Evaluate    json.loads($doc2_text)    json

    Set Suite Variable    ${DOC2_ID}    ${doc2_content['id']}

    New Browser    headless=True

*** Test Cases ***
E2E-US013-001 Documents tab auto-selects first document; click another updates previewer and URL
    [Documentation]    Verifies: list loads ordered by updatedAt DESC; first item auto-selected
    ...                and in URL; clicking a different item updates the previewer and ?doc=.
    [Tags]    US013    smoke    regression

    New Page    ${WEB_BASE_URL}/projects/${PROJECT_ID}
    # Documents tab is active by default (US012)

    # Wait for sidebar to list both documents
    Wait For Elements State    role=heading >> text=Second Document    visible    timeout=15s
    Wait For Elements State    role=option >> text=First Document    visible    timeout=5s

    # Second Document should be first in the sidebar (created later → newer updated_at)
    ${sidebar_items}=    Get Elements    role=option
    # Fallback: look for the first text content in the sidebar
    ${first_item_text}=    Get Text    ${sidebar_items}[0]
    Should Contain    ${first_item_text}    Second Document

    # Second Document is auto-selected (rendered in previewer); URL stays bare
    # on initial load per REQ005/US010 OQ-6 — auto-write of ?doc= was removed.
    # User-driven click writes the URL (asserted below).
    ${url_initial}=    Get Url
    Should Not Contain    ${url_initial}    doc=

    # Previewer shows Second Document content
    Wait For Elements State    main >> text=Hello from doc 2    visible    timeout=10s

    # Click First Document in the sidebar
    Click    role=option >> text=First Document
    Wait For Elements State    main >> text=Hello from doc 1    visible    timeout=10s

    # URL updated to doc1. Click → URL race: wait for the URL to reflect the change.
    Wait Until Keyword Succeeds    10s    200ms    URL Should Contain    doc=${DOC1_ID}
    ${url_after}=    Get Url
    Should Contain    ${url_after}    doc=${DOC1_ID}

    # First Document sidebar item is now active
    ${first_doc_btn}=    Get Element    role=option >> text=First Document
    ${aria}=    Get Attribute    ${first_doc_btn}    aria-selected
    Should Be Equal    ${aria}    true

E2E-US013-002 Refresh with ?doc= rehydrates the same document
    [Documentation]    Verifies that a deep-linked ?doc= query param survives a real browser
    ...                page reload and rehydrates the same document in the previewer.
    [Tags]    US013    regression

    # Navigate directly to Doc1 via deep link
    New Page    ${WEB_BASE_URL}/projects/${PROJECT_ID}?tab=documents&doc=${DOC1_ID}

    # Wait for First Document to appear in the previewer
    Wait For Elements State    main >> text=Hello from doc 1    visible    timeout=10s
    ${first_doc_btn}=    Get Element    role=option >> text=First Document
    ${aria_before}=    Get Attribute    ${first_doc_btn}    aria-selected
    Should Be Equal    ${aria_before}    true

    # Reload the page
    Reload

    # After reload: same document still shown
    Wait For Elements State    main >> text=Hello from doc 1    visible    timeout=10s
    ${first_doc_btn_after}=    Get Element    role=option >> text=First Document
    ${aria_after}=    Get Attribute    ${first_doc_btn_after}    aria-selected
    Should Be Equal    ${aria_after}    true

    # URL still contains doc1 id
    ${url_after}=    Get Url
    Should Contain    ${url_after}    doc=${DOC1_ID}
