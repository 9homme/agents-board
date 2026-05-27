*** Settings ***
Documentation    E2E tests for US003 — Markdown and mermaid rendering in the previewer.
...              Verifies: syntax-highlighted code fence has language-X / hljs classes in a real
...              browser; mermaid diagram renders as an <svg> element; raw mermaid source text
...              is not visible to the user.
Library          Browser
Library          String
Resource         ../../REQ001_agent_board_mcp/mcp_keywords.resource
Resource         resources/project_detail_keywords.resource

Suite Setup      Setup US003 Suite
Suite Teardown   Close Browser

*** Variables ***
${WEB_BASE_URL}    http://localhost:3000
${PROJECT_ID}      ${EMPTY}
${DOC_ID}          ${EMPTY}

*** Keywords ***
Setup US003 Suite
    [Documentation]    Creates a test project and a document containing a Go code fence and
    ...                a mermaid diagram via MCP, then opens a browser.
    ${random}=        Generate Random String    8    [LETTERS]
    ${project_name}=  Set Variable    REQ004 US003 E2E ${random}
    ${session_id}=    Connect To MCP SSE

    ${proj_resp}=    Create Project Tool    ${session_id}    ${project_name}    US003 markdown rendering test
    ${proj_content}=    Evaluate    json.loads('''${proj_resp.json()['result']['content'][0]['text']}''')    json
    Set Suite Variable    ${PROJECT_ID}    ${proj_content['id']}

    # Document content with a Go code fence and a mermaid diagram
    ${doc_content}=    Set Variable
    ...    # Rendering test\n\n## Code block\n\n\`\`\`go\nfunc main() { fmt.Println("hello") }\n\`\`\`\n\n## Diagram\n\n\`\`\`mermaid\ngraph TD; Start-->End;\n\`\`\`\n

    ${doc_resp}=    Create Document Tool    ${session_id}    ${PROJECT_ID}    Rendering test document    ${doc_content}
    ${doc_content_json}=    Evaluate    json.loads('''${doc_resp.json()['result']['content'][0]['text']}''')    json
    Set Suite Variable    ${DOC_ID}    ${doc_content_json['id']}

    New Browser    headless=True

*** Test Cases ***
E2E-US003-001 Previewer renders syntax-highlighted code fence and mermaid SVG
    [Documentation]    Verifies: real browser shows hljs/language-go class on code fence;
    ...                mermaid source replaced by a live <svg> element in the previewer.
    [Tags]    US003    smoke    regression

    # Navigate directly to the document via deep link
    New Page    ${WEB_BASE_URL}/projects/${PROJECT_ID}?tab=documents&doc=${DOC_ID}

    # Wait for the document title to appear in the previewer
    Wait For Elements State    text="Rendering test document"    visible    timeout=15s

    # Assert the document heading rendered as an HTML heading
    Wait For Elements State    css=h1,h2    visible    timeout=5s
    ${heading_el}=    Get Element    css=h1:has-text("Rendering test"), css=h2:has-text("Rendering test")
    Get Text    ${heading_el}    # just confirm it exists without error

    # Assert the Go code block has language-go or hljs class (syntax highlighting applied)
    ${code_el}=    Wait For Elements State
    ...    css=pre > code[class*="language-go"], css=pre > code[class*="hljs"]
    ...    visible    timeout=10s

    # Assert the mermaid SVG is present (mermaid is async; allow generous timeout)
    Wait For Elements State    css=svg    visible    timeout=30s

    # Assert raw mermaid source is NOT visible as user-readable text
    ${raw_mermaid_count}=    Get Element Count    text="graph TD; Start-->End;"
    Should Be Equal As Integers    ${raw_mermaid_count}    0
