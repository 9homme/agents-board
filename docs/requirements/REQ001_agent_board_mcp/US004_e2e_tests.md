# US004 E2E Tests — User Story CRUD

## E2E Tests

| ID | Test Case | Steps | Expected Outcome |
|---|---|---|---|
| E2E-007 | User Story Lifecycle | 1. Create project<br>2. Create user story<br>3. Get user story<br>4. Update user story status<br>5. Delete user story | All steps succeed; user story correctly linked to project. |
| E2E-008 | List User Stories | 1. Create project<br>2. Create two user stories<br>3. Call `list_user_stories` | Both stories are returned. |
