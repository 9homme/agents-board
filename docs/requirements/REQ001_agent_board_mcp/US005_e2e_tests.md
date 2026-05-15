# US005 E2E Tests — Task CRUD

## E2E Tests

| ID | Test Case | Steps | Expected Outcome |
|---|---|---|---|
| E2E-009 | Task Lifecycle | 1. Create project & user story<br>2. Create task under story<br>3. Get task<br>4. Update task status<br>5. Delete task | All steps succeed; task correctly linked to story. |
| E2E-010 | List Tasks | 1. Create project & story<br>2. Create two tasks<br>3. Call `list_tasks` | Both tasks are returned. |
