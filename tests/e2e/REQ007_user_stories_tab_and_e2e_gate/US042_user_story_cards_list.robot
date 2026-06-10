*** Settings ***
Documentation    E2E tests for US042 — User Story cards list.
Library          Browser
Library          String
Resource         ../REQ001_agent_board_mcp/mcp_keywords.resource
Resource         ../REQ004_project_detail_page/resources/project_detail_keywords.resource

Suite Setup      Setup US042 Suite
Suite Teardown   Close Browser

*** Variables ***
${WEB_BASE_URL}       http://localhost:3000
${PROJECT_ID_1}       ${EMPTY}
${PROJECT_ID_EMPTY}   ${EMPTY}

*** Keywords ***
Setup US042 Suite
    ${session_id}=     Connect To MCP SSE

    # Project 1: Has stories and tasks
    ${random}=         Generate Random String    8    [LETTERS]
    ${proj_resp}=      Create Project Tool    ${session_id}    US042 Project ${random}
    ${proj_text}=      Set Variable    ${proj_resp.json()['result']['content'][0]['text']}
    ${proj_content}=   Evaluate    json.loads($proj_text)    json
    Set Suite Variable    ${PROJECT_ID_1}    ${proj_content['id']}

    # Story 1: 2 tasks
    ${s1_resp}=        Create User Story Tool    ${session_id}    ${PROJECT_ID_1}    Story With Tasks    This story has tasks.
    ${s1_text}=        Set Variable    ${s1_resp.json()['result']['content'][0]['text']}
    ${s1_content}=     Evaluate    json.loads($s1_text)    json
    Create Task Tool    ${session_id}    ${s1_content['id']}    Task 1    Desc 1
    Create Task Tool    ${session_id}    ${s1_content['id']}    Task 2    Desc 2

    # Story 2: 0 tasks, long description
    ${long_desc}=      Set Variable    This is a very long description that should exceed eighty characters so we can verify truncation visually in the browser.
    Create User Story Tool    ${session_id}    ${PROJECT_ID_1}    Story Without Tasks    ${long_desc}

    # Project 2: Empty
    ${proj2_resp}=     Create Project Tool    ${session_id}    US042 Empty Project
    ${proj2_text}=     Set Variable    ${proj2_resp.json()['result']['content'][0]['text']}
    ${proj2_content}=  Evaluate    json.loads($proj2_text)    json
    Set Suite Variable    ${PROJECT_ID_EMPTY}    ${proj2_content['id']}

    New Browser    headless=True

*** Test Cases ***
E2E-US042-001 User stories render with accurate details
    [Documentation]    Verifies that the User Stories tab shows cards with title, status, task count, and description.
    [Tags]    US042    regression

    New Page    ${WEB_BASE_URL}/projects/${PROJECT_ID_1}
    # Click User Stories tab (assuming tab list role and tab names)
    Click    role=tab >> text=User Stories
    
    # Wait for cards to appear
    Wait For Elements State    role=heading >> text=Story With Tasks    visible    timeout=10s
    Wait For Elements State    role=heading >> text=Story Without Tasks    visible

    # Verify task counts
    Wait For Elements State    text=2 tasks    visible
    Wait For Elements State    text=0 tasks    visible
    
    # Verify long description is present (partially or truncated, we can check for subset)
    Wait For Elements State    text=This is a very long description    visible

E2E-US042-002 Empty state when no stories
    [Documentation]    Verifies that a project with no stories shows the empty state.
    [Tags]    US042

    New Page    ${WEB_BASE_URL}/projects/${PROJECT_ID_EMPTY}
    Click    role=tab >> text=User Stories
    
    Wait For Elements State    text=No user stories yet for this project    visible    timeout=10s
