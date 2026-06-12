*** Settings ***
Documentation    E2E tests for US010 - State change audit trail
Library          RequestsLibrary
Library          Collections
Resource         ../REQ001_agent_board_mcp/mcp_keywords.resource

Suite Setup      Connect And Create Base Entities

*** Variables ***
${PROJECT_ID}      ${EMPTY}
${REQUIREMENT_ID}  ${EMPTY}
${STORY_ID}        ${EMPTY}
${SESSION_ID}      ${EMPTY}

*** Keywords ***
Connect And Create Base Entities
    ${session_id}=    Connect To MCP SSE
    Set Suite Variable    ${SESSION_ID}    ${session_id}

    ${proj_resp}=    Create Project Tool    ${SESSION_ID}    Test Project US010    Audit Trail Test Project    /e2e/us010-audit
    ${proj_content}=    Evaluate    json.loads('''${proj_resp.json()['result']['content'][0]['text']}''')    json
    Set Suite Variable    ${PROJECT_ID}    ${proj_content['id']}

    ${req_resp}=    Create Requirement Tool    ${SESSION_ID}    ${PROJECT_ID}    Default
    ${req_content}=    Evaluate    json.loads('''${req_resp.json()['result']['content'][0]['text']}''')    json
    Set Suite Variable    ${REQUIREMENT_ID}    ${req_content['id']}

    ${story_resp}=    Create User Story Tool    ${SESSION_ID}    ${PROJECT_ID}    ${REQUIREMENT_ID}    Test Story US010    Story for task audit testing    draft
    ${story_content}=    Evaluate    json.loads('''${story_resp.json()['result']['content'][0]['text']}''')    json
    Set Suite Variable    ${STORY_ID}    ${story_content['id']}

Get Task Audit Trail Tool
    [Arguments]    ${session_id}    ${task_id}
    ${args}=    Create Dictionary    taskId=${task_id}
    ${resp}=    Call MCP Tool    ${session_id}    get_task_audit_trail    ${args}
    RETURN    ${resp}

Get User Story Audit Trail Tool
    [Arguments]    ${session_id}    ${story_id}
    ${args}=    Create Dictionary    userStoryId=${story_id}
    ${resp}=    Call MCP Tool    ${session_id}    get_user_story_audit_trail    ${args}
    RETURN    ${resp}

*** Test Cases ***
E2E-001 Retrieve task audit trail after valid transitions
    [Tags]    US010    regression
    # 1. Create task
    ${create_resp}=    Create Task Tool    ${SESSION_ID}    ${STORY_ID}    Task 1    Test task audit    pending
    ${task_content}=    Evaluate    json.loads('''${create_resp.json()['result']['content'][0]['text']}''')    json
    ${task_id}=    Set Variable    ${task_content['id']}

    # 2. Update to in_progress
    Update Task Tool    ${SESSION_ID}    ${task_id}    status=in_progress

    # 3. Update to in_review
    Update Task Tool    ${SESSION_ID}    ${task_id}    status=in_review

    # 4. Get audit trail
    ${audit_resp}=    Get Task Audit Trail Tool    ${SESSION_ID}    ${task_id}
    ${audit_content}=    Evaluate    json.loads('''${audit_resp.json()['result']['content'][0]['text']}''')    json
    
    # Expected at least 2 entries. 
    # Entry 1: pending -> in_progress
    # Entry 2: in_progress -> in_review
    ${trail}=    Set Variable    ${audit_content['auditTrail']}
    Length Should Be    ${trail}    2
    
    Should Be Equal As Strings    ${trail[0]['fromStatus']}    pending
    Should Be Equal As Strings    ${trail[0]['toStatus']}      in_progress
    Should Be Equal As Strings    ${trail[0]['entityType']}    task

    Should Be Equal As Strings    ${trail[1]['fromStatus']}    in_progress
    Should Be Equal As Strings    ${trail[1]['toStatus']}      in_review

E2E-002 Retrieve story audit trail after valid transitions
    [Tags]    US010    regression
    # 1. Create story
    ${create_resp}=    Create User Story Tool    ${SESSION_ID}    ${PROJECT_ID}    ${REQUIREMENT_ID}    Story 2    Test story audit    draft
    ${story_content}=    Evaluate    json.loads('''${create_resp.json()['result']['content'][0]['text']}''')    json
    ${story_id}=    Set Variable    ${story_content['id']}

    # 2. Update to in_development
    Update User Story Tool    ${SESSION_ID}    ${story_id}    status=in_development

    # 3. Get audit trail
    ${audit_resp}=    Get User Story Audit Trail Tool    ${SESSION_ID}    ${story_id}
    ${audit_content}=    Evaluate    json.loads('''${audit_resp.json()['result']['content'][0]['text']}''')    json
    
    ${trail}=    Set Variable    ${audit_content['auditTrail']}
    Length Should Be    ${trail}    1
    
    Should Be Equal As Strings    ${trail[0]['fromStatus']}    draft
    Should Be Equal As Strings    ${trail[0]['toStatus']}      in_development
    Should Be Equal As Strings    ${trail[0]['entityType']}    user_story

E2E-003 Audit record not created on invalid transition
    [Tags]    US010
    # 1. Create task
    ${create_resp}=    Create Task Tool    ${SESSION_ID}    ${STORY_ID}    Task 3    Test invalid audit    pending
    ${task_content}=    Evaluate    json.loads('''${create_resp.json()['result']['content'][0]['text']}''')    json
    ${task_id}=    Set Variable    ${task_content['id']}

    # 2. Invalid update to completed
    Update Task Tool    ${SESSION_ID}    ${task_id}    status=completed

    # 3. Get audit trail
    ${audit_resp}=    Get Task Audit Trail Tool    ${SESSION_ID}    ${task_id}
    ${audit_content}=    Evaluate    json.loads('''${audit_resp.json()['result']['content'][0]['text']}''')    json
    
    # Trail should be empty
    ${trail}=    Set Variable    ${audit_content['auditTrail']}
    Should Be Empty    ${trail}
