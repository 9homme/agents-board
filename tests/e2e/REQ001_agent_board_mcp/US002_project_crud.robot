*** Settings ***
Resource          mcp_keywords.resource
Test Tags         US002

*** Test Cases ***
Scenario: Full Project CRUD Lifecycle
    [Documentation]    Create, read, update, list and delete a project via MCP tools.
    [Tags]    E2E-002
    ${session_id}=    Connect To MCP SSE
    
    # Create
    ${resp}=    Create Project Tool    ${session_id}    Test Project    Initial description
    ${project}=    Set Variable    ${resp.json()['result']['content'][0]['text']}
    # Note: text is JSON string according to architecture
    ${project_json}=    Evaluate    json.loads('''${project}''')    json
    ${project_id}=    Set Variable    ${project_json['id']}
    Should Be Equal    ${project_json['name']}    Test Project
    
    # Get
    ${resp}=    Get Project Tool    ${session_id}    ${project_id}
    ${project_text}=    Set Variable    ${resp.json()['result']['content'][0]['text']}
    ${project_json}=    Evaluate    json.loads('''${project_text}''')    json
    Should Be Equal    ${project_json['id']}    ${project_id}
    
    # Update
    ${resp}=    Update Project Tool    ${session_id}    ${project_id}    name=Updated Project
    ${project_text}=    Set Variable    ${resp.json()['result']['content'][0]['text']}
    ${project_json}=    Evaluate    json.loads('''${project_text}''')    json
    Should Be Equal    ${project_json['name']}    Updated Project
    
    # List
    ${resp}=    List Projects Tool    ${session_id}
    ${list_text}=    Set Variable    ${resp.json()['result']['content'][0]['text']}
    ${list_json}=    Evaluate    json.loads('''${list_text}''')    json
    # Check if project_id is in the list
    ${found}=    Set Variable    ${FALSE}
    FOR    ${p}    IN    @{list_json['projects']}
        IF    '${p['id']}' == '${project_id}'
            ${found}=    Set Variable    ${TRUE}
            BREAK
        END
    END
    Should Be True    ${found}
    
    # Delete
    ${resp}=    Delete Project Tool    ${session_id}    ${project_id}
    ${delete_text}=    Set Variable    ${resp.json()['result']['content'][0]['text']}
    ${delete_json}=    Evaluate    json.loads('''${delete_text}''')    json
    Should Be True    ${delete_json['success']}
    
    # Verify Deleted
    ${resp}=    Get Project Tool    ${session_id}    ${project_id}
    Should Be True    ${resp.json()['result']['isError']}
