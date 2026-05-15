*** Settings ***
Resource          mcp_keywords.resource
Test Tags         US003

*** Test Cases ***
Scenario: Full Document CRUD Lifecycle
    [Documentation]    Create, read, update, list and delete a document via MCP tools.
    [Tags]    E2E-003
    ${session_id}=    Connect To MCP SSE
    
    # Pre-requisite: Create Project
    ${resp}=    Create Project Tool    ${session_id}    Doc Project
    ${project_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    ${project_id}=    Set Variable    ${project_json['id']}
    
    # Create Document
    ${resp}=    Create Document Tool    ${session_id}    ${project_id}    Test Doc    Some content
    ${doc_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    ${doc_id}=    Set Variable    ${doc_json['id']}
    Should Be Equal    ${doc_json['title']}    Test Doc
    
    # Get Document
    ${resp}=    Get Document Tool    ${session_id}    ${doc_id}
    ${doc_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    Should Be Equal    ${doc_json['id']}    ${doc_id}
    
    # Update Document
    ${resp}=    Update Document Tool    ${session_id}    ${doc_id}    title=Updated Doc
    ${doc_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    Should Be Equal    ${doc_json['title']}    Updated Doc
    
    # List Documents
    ${resp}=    List Documents Tool    ${session_id}    ${project_id}
    ${list_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    ${found}=    Set Variable    ${FALSE}
    FOR    ${d}    IN    @{list_json['documents']}
        IF    '${d['id']}' == '${doc_id}'
            ${found}=    Set Variable    ${TRUE}
            BREAK
        END
    END
    Should Be True    ${found}
    
    # Delete Document
    ${resp}=    Delete Document Tool    ${session_id}    ${doc_id}
    ${delete_json}=    Evaluate    json.loads('''${resp.json()['result']['content'][0]['text']}''')    json
    Should Be True    ${delete_json['success']}
    
    # Verify Deleted
    ${resp}=    Get Document Tool    ${session_id}    ${doc_id}
    Should Be True    ${resp.json()['result']['isError']}
