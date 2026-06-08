*** Settings ***
Documentation    E2E tests for US002 — Makefile e2e-up health-check fix + data-only e2e-seed.
Library          OperatingSystem
Library          Process
Library          String

*** Variables ***
${MAKEFILE_PATH}    ${CURDIR}/../../../Makefile

*** Test Cases ***
E2E-US002-001 mcp-server health-check is bounded and e2e-seed is data-only
    [Documentation]    Verifies that the Makefile has been updated so that the mcp-server probe
    ...                is bounded with --max-time 5 and that e2e-seed has removed the migration step.
    [Tags]    US002

    ${makefile_content}=    Get File    ${MAKEFILE_PATH}
    
    # Assert e2e-up mcp probe is bounded
    Should Match Regexp    ${makefile_content}    (?s)e2e-up:.*curl -sf --max-time 5 http://localhost:8081/sse
    
    # Assert e2e-seed does not contain migration loop
    ${make_dry_run}=    Run Process    make    -n    e2e-seed    cwd=${CURDIR}/../../..
    Should Not Contain    ${make_dry_run.stdout}    applying migration
    Should Contain    ${make_dry_run.stdout}    psql
