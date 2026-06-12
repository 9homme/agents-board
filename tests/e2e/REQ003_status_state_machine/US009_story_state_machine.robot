*** Settings ***
Documentation    E2E tests for US009 - Story State Machine
Library          RequestsLibrary
Library          Collections
Resource         ../REQ001_agent_board_mcp/mcp_keywords.resource

Suite Setup      Connect And Create Base Entities

*** Variables ***
${PROJECT_ID}      ${EMPTY}
${REQUIREMENT_ID}  ${EMPTY}
${SESSION_ID}      ${EMPTY}

*** Keywords ***
Connect And Create Base Entities
    ${session_id}=    Connect To MCP SSE
    Set Suite Variable    ${SESSION_ID}    ${session_id}

    ${proj_resp}=    Create Project Tool    ${SESSION_ID}    Test Project US009    State Machine Test Project    /e2e/us009-state
    ${proj_content}=    Evaluate    json.loads('''${proj_resp.json()['result']['content'][0]['text']}''')    json
    Set Suite Variable    ${PROJECT_ID}    ${proj_content['id']}

    ${req_resp}=    Create Requirement Tool    ${SESSION_ID}    ${PROJECT_ID}    Default
    ${req_content}=    Evaluate    json.loads('''${req_resp.json()['result']['content'][0]['text']}''')    json
    Set Suite Variable    ${REQUIREMENT_ID}    ${req_content['id']}

*** Test Cases ***
E2E-001 Valid story state machine transitions
    [Tags]    US009    regression
    # 1. Create story with status draft
    ${create_resp}=    Create User Story Tool    ${SESSION_ID}    ${PROJECT_ID}    ${REQUIREMENT_ID}    Story 1    Test story    draft
    ${story_content}=    Evaluate    json.loads('''${create_resp.json()['result']['content'][0]['text']}''')    json
    ${story_id}=    Set Variable    ${story_content['id']}
    Should Be Equal As Strings    ${story_content['status']}    draft

    # 2. Update to in_development
    ${up1_resp}=    Update User Story Tool    ${SESSION_ID}    ${story_id}    status=in_development
    ${up1_result}=    Set Variable    ${up1_resp.json()['result']}
    Should Not Contain    ${up1_result}    isError

    # 3. Update to in_signoff
    ${up2_resp}=    Update User Story Tool    ${SESSION_ID}    ${story_id}    status=in_signoff
    ${up2_result}=    Set Variable    ${up2_resp.json()['result']}
    Should Not Contain    ${up2_result}    isError

    # 4. Update to changes_requested
    ${up3_resp}=    Update User Story Tool    ${SESSION_ID}    ${story_id}    status=changes_requested
    ${up3_result}=    Set Variable    ${up3_resp.json()['result']}
    Should Not Contain    ${up3_result}    isError

    # 5. Update to in_development (cycle back)
    ${up4_resp}=    Update User Story Tool    ${SESSION_ID}    ${story_id}    status=in_development
    ${up4_result}=    Set Variable    ${up4_resp.json()['result']}
    Should Not Contain    ${up4_result}    isError

    # Update to in_signoff again
    ${up5_resp}=    Update User Story Tool    ${SESSION_ID}    ${story_id}    status=in_signoff

    # 6. Update to done
    ${up6_resp}=    Update User Story Tool    ${SESSION_ID}    ${story_id}    status=done
    ${up6_result}=    Set Variable    ${up6_resp.json()['result']}
    Should Not Contain    ${up6_result}    isError

E2E-002 Invalid story state machine transition rejected
    [Tags]    US009    regression
    # 1. Create story with status draft
    ${create_resp}=    Create User Story Tool    ${SESSION_ID}    ${PROJECT_ID}    ${REQUIREMENT_ID}    Story 2    Test invalid    draft
    ${story_content}=    Evaluate    json.loads('''${create_resp.json()['result']['content'][0]['text']}''')    json
    ${story_id}=    Set Variable    ${story_content['id']}

    # 2. Update to done directly from draft
    ${up_resp}=    Update User Story Tool    ${SESSION_ID}    ${story_id}    status=done
    ${up_result}=    Set Variable    ${up_resp.json()['result']}

    # Expected to have isError: true
    Dictionary Should Contain Key    ${up_result}    isError
    Should Be True    ${up_result['isError']}
    ${error_text}=    Set Variable    ${up_result['content'][0]['text']}
    Should Not Be Empty    ${error_text}

E2E-003 Enforce initial state on story creation
    [Tags]    US009
    # 1. Create story with status done
    ${create_resp}=    Create User Story Tool    ${SESSION_ID}    ${PROJECT_ID}    ${REQUIREMENT_ID}    Story 3    Test initial    done
    ${create_result}=    Set Variable    ${create_resp.json()['result']}

    # It should either fail with an error or default to draft.
    ${has_error}=    Run Keyword And Return Status    Dictionary Should Contain Key    ${create_result}    isError

    Run Keyword If    not ${has_error}    Verify Defaulted To Draft    ${create_result}

*** Keywords ***
Verify Defaulted To Draft
    [Arguments]    ${result}
    ${story_content}=    Evaluate    json.loads('''${result['content'][0]['text']}''')    json
    Should Be Equal As Strings    ${story_content['status']}    draft
