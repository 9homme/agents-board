# US003 E2E Tests — Document CRUD

## E2E Tests

| ID | Test Case | Steps | Expected Outcome |
|---|---|---|---|
| E2E-005 | Document Lifecycle | 1. Create project<br>2. Create document in project<br>3. Get document<br>4. Update document<br>5. Delete document | All steps succeed; document correctly linked to project. |
| E2E-006 | List Documents | 1. Create project<br>2. Create two documents<br>3. Call `list_documents` with project ID | Both documents are returned for that project. |
