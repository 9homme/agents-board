*** Settings ***
Resource          mcp_keywords.resource
Test Tags         US005

*** Test Cases ***
Scenario: Full Task CRUD Lifecycle
    [Documentation]    Create, read, update, list and delete a task via MCP tools.
    [Tags]    E2E-005
    ${session_id}=    Connect To MCP SSE
    
    # Pre-requisites: Create Project and Story
    ${resp}=    Create Project Tool    ${session_id}    Task Project
    ${project_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    ${project_id}=    Set Variable    ${project_json['id']}
    
    ${resp}=    Create User Story Tool    ${session_id}    ${project_id}    Task Story    Description
    ${story_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    ${story_id}=    Set Variable    ${story_json['id']}
    
    # Create Task
    ${resp}=    Create Task Tool    ${session_id}    ${story_id}    Test Task    Task description    pending
    ${task_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    ${task_id}=    Set Variable    ${task_json['id']}
    Should Be Equal    ${task_json['title']}    Test Task
    Should Be Equal    ${task_json['status']}    pending
    
    # Get Task
    ${resp}=    Get Task Tool    ${session_id}    ${task_id}
    ${task_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    Should Be Equal    ${task_json['id']}    ${task_id}
    
    # Update Task
    ${resp}=    Update Task Tool    ${session_id}    ${task_id}    status=in_progress
    ${task_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    Should Be Equal    ${task_json['status']}    in_progress
    
    # List Tasks
    ${resp}=    List Tasks Tool    ${session_id}    ${story_id}
    ${list_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    ${found}=    Set Variable    ${FALSE}
    FOR    ${t}    IN    @{list_json['tasks']}
        IF    '${t['id']}' == '${task_id}'
            ${found}=    Set Variable    ${TRUE}
            BREAK
        END
    END
    Should Be True    ${found}
    
    # Delete Task
    ${resp}=    Delete Task Tool    ${session_id}    ${task_id}
    ${delete_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    Should Be True    ${delete_json['success']}
    
    # Verify Deleted
    ${resp}=    Get Task Tool    ${session_id}    ${task_id}
    Should Be True    ${resp.json()['result']['isError']}
