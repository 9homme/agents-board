*** Settings ***
Documentation    Shared keywords for REQ004 project detail page e2e tests.
...              These keywords wrap common Browser-library operations specific to the
...              /projects/[id] page UI (headings, tabs, sidebar, previewer).
Library          Browser
Library          String

*** Variables ***
${WEB_BASE_URL}    http://localhost:3000

*** Keywords ***
Wait For Project Detail Page
    [Documentation]    Waits for the /projects/{id} page to finish loading the project header.
    [Arguments]    ${project_id}    ${timeout}=15s
    Wait For Elements State    css=h1    visible    timeout=${timeout}

Get Active Tab Name
    [Documentation]    Returns the text of the currently active tab (aria-selected=true).
    ${active_tab}=    Get Element    role=tab[aria-selected=true]
    ${tab_text}=      Get Text    ${active_tab}
    RETURN    ${tab_text}

Assert Tab Is Active
    [Documentation]    Asserts that the tab whose text matches ${tab_name} has aria-selected=true.
    [Arguments]    ${tab_name}
    ${tab}=    Get Element    role=tab >> text="${tab_name}"
    ${aria}=   Get Attribute    ${tab}    aria-selected
    Should Be Equal    ${aria}    true

Assert URL Contains Doc Id
    [Documentation]    Asserts that the current URL contains the given document id as the ?doc= param.
    [Arguments]    ${doc_id}
    ${url}=    Get Url
    Should Contain    ${url}    doc=${doc_id}
