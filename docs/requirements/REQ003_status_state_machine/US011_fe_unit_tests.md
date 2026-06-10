# US011 — Frontend component test specification

N/A — US011 is a backend-only quality refinement; no frontend surface is affected. No FCT-* test cases apply.

The story's scope is limited to driving `golangci-lint` to zero findings inside `services/agent-board/`. There are no changes to `web/`, no new API contracts, and no UI behaviour to assert. The frontend test suite for US008/US009/US010 remains the regression guard for any frontend-adjacent behaviour.
