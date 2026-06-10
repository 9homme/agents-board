*** Settings ***
Documentation    E2E tests for US044 — Introduce Requirement entity + re-parent model.
...
...              E2E-044-001: Running server returns projects with `path` field after migration.
...              E2E-044-002: Default requirement exists for a seeded project after migration.
Library          RequestsLibrary
Library          Collections
Library          String

*** Variables ***
${API_BASE_URL}    %{API_BASE_URL=http://localhost:8080}

*** Test Cases ***

E2E-044-001 Projects list includes path field after migration
    [Documentation]    Verifies that the migration added the `path` column and the
    ...                GET /api/v1/projects endpoint returns it on every item.
    [Tags]    US044    smoke
    ${response}=    GET    ${API_BASE_URL}/api/v1/projects    expected_status=200
    ${body}=    Set Variable    ${response.json()}
    Should Not Be Empty    ${body}[projects]    msg=Expected at least one seeded project
    ${first_project}=    Get From List    ${body}[projects]    0
    Dictionary Should Contain Key    ${first_project}    path
    Should Not Be Empty    ${first_project}[path]    msg=path field must not be empty

E2E-044-002 Seeded project has Default requirement after migration backfill
    [Documentation]    Verifies that the migration backfill created a Default requirement
    ...                for every existing project; fetches the first project's requirements
    ...                and asserts the Default requirement is present.
    [Tags]    US044    regression
    # Step 1: get first project id
    ${projects_resp}=    GET    ${API_BASE_URL}/api/v1/projects    expected_status=200
    ${projects}=    Set Variable    ${projects_resp.json()}[projects]
    Should Not Be Empty    ${projects}    msg=Seed data must include at least one project
    ${project_id}=    Set Variable    ${projects}[0][id]

    # Step 2: list requirements for that project
    ${req_resp}=    GET    ${API_BASE_URL}/api/v1/projects/${project_id}/requirements    expected_status=200
    ${body}=    Set Variable    ${req_resp.json()}
    Dictionary Should Contain Key    ${body}    requirements
    Should Not Be Empty    ${body}[requirements]    msg=Migrated project must have at least the Default requirement

    # Step 3: assert Default requirement shape
    ${first_req}=    Get From List    ${body}[requirements]    0
    Should Be Equal    ${first_req}[name]    Default
    Should Be Equal    ${first_req}[status]    draft
    Should Be Equal    ${first_req}[projectId]    ${project_id}
    Dictionary Should Contain Key    ${first_req}    id
    Dictionary Should Contain Key    ${first_req}    createdAt
    Dictionary Should Contain Key    ${first_req}    updatedAt
