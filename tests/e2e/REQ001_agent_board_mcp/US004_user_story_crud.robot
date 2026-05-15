*** Settings ***
Resource          mcp_keywords.resource
Test Tags         US004

*** Test Cases ***
Scenario: Full User Story CRUD Lifecycle
    [Documentation]    Create, read, update, list and delete a user story via MCP tools.
    [Tags]    E2E-004
    ${session_id}=    Connect To MCP SSE
    
    # Pre-requisite: Create Project
    ${resp}=    Create Project Tool    ${session_id}    Story Project
    ${project_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    ${project_id}=    Set Variable    ${project_json['id']}
    
    # Create Story
    ${resp}=    Create User Story Tool    ${session_id}    ${project_id}    Test Story    Story description    draft
    ${story_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    ${story_id}=    Set Variable    ${story_json['id']}
    Should Be Equal    ${story_json['title']}    Test Story
    Should Be Equal    ${story_json['status']}    draft
    
    # Get Story
    ${resp}=    Get User Story Tool    ${session_id}    ${story_id}
    ${story_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    Should Be Equal    ${story_json['id']}    ${story_id}
    
    # Update Story
    ${resp}=    Update User Story Tool    ${session_id}    ${story_id}    status=in_development
    ${story_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    Should Be Equal    ${story_json['status']}    in_development
    
    # List Stories
    ${resp}=    List User Stories Tool    ${session_id}    ${project_id}
    ${list_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    ${found}=    Set Variable    ${FALSE}
    FOR    ${s}    IN    @{list_json['userStories']}
        IF    '${s['id']}' == '${story_id}'
            ${found}=    Set Variable    ${TRUE}
            BREAK
        END
    END
    Should Be True    ${found}
    
    # Delete Story
    ${resp}=    Delete User Story Tool    ${session_id}    ${story_id}
    ${delete_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    Should Be True    ${delete_json['success']}
    
    # Verify Deleted
    ${resp}=    Get User Story Tool    ${session_id}    ${story_id}
    Should Be True    ${resp.json()['result']['isError']}
