# US004 — E2E test specification (Robot Framework)

N/A — no behavioural change in production code; existing e2e suites for US001/US002/US003 remain the regression guard. No E2E-* cases for US004.

US004 makes no change to any API contract, MCP tool response shape, or user-observable flow. All verifiable claims (lint clean, race-free tests, per-finding-category code shape) are provable at the static-analysis and unit layer inside `services/agent-board/` and are fully captured by UT-001 through UT-010 in `US004_be_unit_tests.md`. Promoting any of those checks to e2e would violate the test-pyramid policy (no user-facing behaviour to exercise end-to-end).

No `.robot` file is created for US004.
