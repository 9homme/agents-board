*** Settings ***
Documentation    E2E tests for US041 — GitHub Actions e2e PR quality gate.
Library          OperatingSystem

*** Variables ***
${WORKFLOW_PATH}    ${CURDIR}/../../../.github/workflows/e2e.yml

*** Test Cases ***
E2E-US041-001 GitHub Actions workflow is correctly configured
    [Documentation]    Verifies that .github/workflows/e2e.yml exists and contains the correct configuration.
    [Tags]    US041

    File Should Exist    ${WORKFLOW_PATH}
    ${workflow_content}=    Get File    ${WORKFLOW_PATH}
    
    # Assert trigger
    Should Match Regexp    ${workflow_content}    (?ms)^on:.*pull_request:.*branches:\\s*\\["?main"?\\]
    Should Not Match Regexp    ${workflow_content}    (?m)^\\s*push:
    
    # Assert Makefile targets are used
    Should Contain    ${workflow_content}    make e2e-up
    Should Contain    ${workflow_content}    make e2e-seed
    Should Contain    ${workflow_content}    make e2e-run
    
    # Assert artifacts and teardown use always()
    Should Match Regexp    ${workflow_content}    (?s)if:\\s*always\\(\\).*actions/upload-artifact
    Should Match Regexp    ${workflow_content}    (?s)if:\\s*always\\(\\).*make e2e-down
