*** Settings ***
Documentation    E2E tests for US005 — User story detail + tasks in side drawer.
...                E2E-US005-001: clicking a card opens the drawer with detail + tasks.
...                E2E-US005-002: switching stories updates the drawer; Story 2 (0 tasks) asserts
...                the empty-state "No tasks for this story." as a sub-step; Escape closes the drawer.
...                (Standalone E2E-US005-003 removed — empty-state is a sub-assertion of E2E-002.)
Library          Browser
Library          String
Resource         ../REQ001_agent_board_mcp/mcp_keywords.resource
Resource         ../REQ004_project_detail_page/resources/project_detail_keywords.resource

Suite Setup      Setup US005 Suite
Suite Teardown   Close Browser

*** Variables ***
${WEB_BASE_URL}       http://localhost:3000
${PROJECT_ID}         ${EMPTY}

*** Keywords ***
Setup US005 Suite
    ${session_id}=     Connect To MCP SSE
    ${random}=         Generate Random String    8    [LETTERS]
    ${proj_resp}=      Create Project Tool    ${session_id}    US005 Project ${random}
    ${proj_text}=      Set Variable    ${proj_resp.json()['result']['content'][0]['text']}
    ${proj_content}=   Evaluate    json.loads($proj_text)    json
    Set Suite Variable    ${PROJECT_ID}    ${proj_content['id']}

    # Story 1: 2 tasks
    ${s1_resp}=        Create User Story Tool    ${session_id}    ${PROJECT_ID}    US005 Story 1    Full description for story 1
    ${s1_text}=        Set Variable    ${s1_resp.json()['result']['content'][0]['text']}
    ${s1_content}=     Evaluate    json.loads($s1_text)    json
    Create Task Tool    ${session_id}    ${s1_content['id']}    Task A    Description A
    Create Task Tool    ${session_id}    ${s1_content['id']}    Task B    Description B

    # Story 2: 0 tasks
    Create User Story Tool    ${session_id}    ${PROJECT_ID}    US005 Story 2    Full description for story 2

    New Browser    headless=True

*** Test Cases ***
E2E-US005-001 Clicking a card opens detail drawer with tasks
    [Documentation]    Verifies that clicking a card opens the drawer and shows detail + tasks.
    [Tags]    US005    regression

    New Page    ${WEB_BASE_URL}/projects/${PROJECT_ID}
    Click    role=tab >> text=User Stories
    
    # Wait for cards to appear
    Wait For Elements State    role=heading >> text=US005 Story 1    visible    timeout=10s
    
    # Click the card
    Click    text=US005 Story 1
    
    # Verify drawer opens (role=dialog)
    Wait For Elements State    role=dialog    visible    timeout=5s
    
    # Verify full description and tasks
    Wait For Elements State    role=dialog >> text=Full description for story 1    visible
    Wait For Elements State    role=dialog >> text=Task A    visible
    Wait For Elements State    role=dialog >> text=Description A    visible
    Wait For Elements State    role=dialog >> text=Task B    visible
    
    # List is still visible
    Wait For Elements State    role=heading >> text=US005 Story 2    visible

E2E-US005-002 Switching stories and closing the drawer
    [Documentation]    Verifies clicking another card updates drawer; Story 2 (0 tasks) shows empty-state; Escape closes the drawer.
    [Tags]    US005

    # Start from a clean state
    New Page    ${WEB_BASE_URL}/projects/${PROJECT_ID}
    Click    role=tab >> text=User Stories
    Wait For Elements State    role=heading >> text=US005 Story 1    visible    timeout=10s
    Click    text=US005 Story 1
    Wait For Elements State    role=dialog >> text=Full description for story 1    visible    timeout=5s

    # Click Story 2 to switch context
    Click    text=US005 Story 2
    
    # Drawer should update
    Wait For Elements State    role=dialog >> text=Full description for story 2    visible    timeout=5s
    
    # Story 2 has 0 tasks, verify empty state
    Wait For Elements State    role=dialog >> text=No tasks for this story    visible
    
    # Close the drawer
    Keyboard Key    press    Escape
    
    # Drawer should unmount
    Wait For Elements State    role=dialog    hidden    timeout=5s
