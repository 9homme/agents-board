# US002 E2E Tests — Project CRUD

## E2E Tests

| ID | Test Case | Steps | Expected Outcome |
|---|---|---|---|
| E2E-003 | Project Lifecycle | 1. Create project "Test Project"<br>2. Get project by ID<br>3. Update project name<br>4. Delete project | All steps succeed; data matches at each step; final delete returns success. |
| E2E-004 | List Projects | 1. Create two projects<br>2. Call `list_projects` | Both projects are present in the list. |
