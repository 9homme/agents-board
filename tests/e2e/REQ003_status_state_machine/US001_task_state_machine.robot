*** Settings ***
Documentation    E2E tests for US001 - Task State Machine
Library          RequestsLibrary
Library          Collections
Resource         ../REQ001_agent_board_mcp/mcp_keywords.resource

Suite Setup      Connect And Create Base Entities

*** Variables ***
${PROJECT_ID}    ${EMPTY}
${STORY_ID}      ${EMPTY}
${SESSION_ID}    ${EMPTY}

*** Keywords ***
Connect And Create Base Entities
    ${session_id}=    Connect To MCP SSE
    Set Suite Variable    ${SESSION_ID}    ${session_id}
    
    ${proj_resp}=    Create Project Tool    ${SESSION_ID}    Test Project US001    State Machine Test Project
    ${proj_content}=    Evaluate    json.loads('''${proj_resp.json()['params']['result']['content'][0]['text']}''')    json
    Set Suite Variable    ${PROJECT_ID}    ${proj_content['id']}

    ${story_resp}=    Create User Story Tool    ${SESSION_ID}    ${PROJECT_ID}    Test Story US001    Story for task testing    draft
    ${story_content}=    Evaluate    json.loads('''${story_resp.json()['params']['result']['content'][0]['text']}''')    json
    Set Suite Variable    ${STORY_ID}    ${story_content['id']}

*** Test Cases ***
E2E-001 Valid task state machine transitions
    [Tags]    US001    regression
    # 1. Create task with status pending
    ${create_resp}=    Create Task Tool    ${SESSION_ID}    ${STORY_ID}    Task 1    Test task    pending
    ${task_content}=    Evaluate    json.loads('''${create_resp.json()['params']['result']['content'][0]['text']}''')    json
    ${task_id}=    Set Variable    ${task_content['id']}
    Should Be Equal As Strings    ${task_content['status']}    pending

    # 2. Update to in_progress
    ${up1_resp}=    Update Task Tool    ${SESSION_ID}    ${task_id}    status=in_progress
    ${up1_result}=    Set Variable    ${up1_resp.json()['params']['result']}
    Should Not Contain    ${up1_result}    isError
    
    # 3. Update to in_review
    ${up2_resp}=    Update Task Tool    ${SESSION_ID}    ${task_id}    status=in_review
    ${up2_result}=    Set Variable    ${up2_resp.json()['params']['result']}
    Should Not Contain    ${up2_result}    isError

    # 4. Update to changes_requested
    ${up3_resp}=    Update Task Tool    ${SESSION_ID}    ${task_id}    status=changes_requested
    ${up3_result}=    Set Variable    ${up3_resp.json()['params']['result']}
    Should Not Contain    ${up3_result}    isError

    # 5. Update to in_progress (cycle back)
    ${up4_resp}=    Update Task Tool    ${SESSION_ID}    ${task_id}    status=in_progress
    ${up4_result}=    Set Variable    ${up4_resp.json()['params']['result']}
    Should Not Contain    ${up4_result}    isError
    
    # Update to in_review again
    ${up5_resp}=    Update Task Tool    ${SESSION_ID}    ${task_id}    status=in_review
    
    # 6. Update to completed
    ${up6_resp}=    Update Task Tool    ${SESSION_ID}    ${task_id}    status=completed
    ${up6_result}=    Set Variable    ${up6_resp.json()['params']['result']}
    Should Not Contain    ${up6_result}    isError

E2E-002 Invalid task state machine transition rejected
    [Tags]    US001    regression
    # 1. Create task with status pending
    ${create_resp}=    Create Task Tool    ${SESSION_ID}    ${STORY_ID}    Task 2    Test invalid    pending
    ${task_content}=    Evaluate    json.loads('''${create_resp.json()['params']['result']['content'][0]['text']}''')    json
    ${task_id}=    Set Variable    ${task_content['id']}

    # 2. Update to completed directly from pending
    ${up_resp}=    Update Task Tool    ${SESSION_ID}    ${task_id}    status=completed
    ${up_result}=    Set Variable    ${up_resp.json()['params']['result']}
    
    # Expected to have isError: true
    Dictionary Should Contain Key    ${up_result}    isError
    Should Be True    ${up_result['isError']}
    ${error_text}=    Set Variable    ${up_result['content'][0]['text']}
    Should Not Be Empty    ${error_text}

E2E-003 Enforce initial state on task creation
    [Tags]    US001
    # 1. Create task with status completed
    ${create_resp}=    Create Task Tool    ${SESSION_ID}    ${STORY_ID}    Task 3    Test initial    completed
    ${create_result}=    Set Variable    ${create_resp.json()['params']['result']}
    
    # It should either fail with an error or default to pending. 
    # Let's check if it failed. If it didn't fail, it should be pending.
    ${has_error}=    Run Keyword And Return Status    Dictionary Should Contain Key    ${create_result}    isError
    
    Run Keyword If    not ${has_error}    Verify Defaulted To Pending    ${create_result}

*** Keywords ***
Verify Defaulted To Pending
    [Arguments]    ${result}
    ${task_content}=    Evaluate    json.loads('''${result['content'][0]['text']}''')    json
    Should Be Equal As Strings    ${task_content['status']}    pending
