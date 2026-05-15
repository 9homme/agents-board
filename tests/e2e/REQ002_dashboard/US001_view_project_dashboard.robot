*** Settings ***
Documentation    E2E test for viewing the project dashboard.
Library          Browser
Resource         ../REQ001_agent_board_mcp/mcp_keywords.resource

Suite Setup      New Browser    headless=True
Suite Teardown   Close Browser

*** Variables ***
${WEB_BASE_URL}    http://localhost:3000

*** Test Cases ***
US001 View Project Dashboard End-to-End
    [Documentation]    Verifies that projects created via MCP appear on the web dashboard.
    [Tags]             US001    smoke    regression
    
    # Pre-condition: Create a project to ensure the list is not empty
    ${session_id}=    Connect To MCP SSE
    ${resp}=    Create Project Tool    ${session_id}    Dashboard E2E Test    Created by Robot
    
    # Action: View dashboard
    New Page           ${WEB_BASE_URL}/
    
    # Expected: The newly created project is visible on the page
    Wait For Elements State    text="Dashboard E2E Test"    visible    timeout=10s
    Wait For Elements State    text="Created by Robot"      visible    timeout=10s
