# US010 — E2E test specification

**Owner:** tester. No Robot file is created for US010 — see rationale below.

## Why e2e does not apply

US010 is a FE-only internal refactor with no new product behaviour. The four fixes are:

1. **MermaidDiagram ref-attach** — the rendered DOM is identical to today (`<div role="img"><svg>...</svg></div>`). The e2e layer would see exactly the same output as before the change. FCT-US010-001/002/003 cover the structural correctness.

2. **useDocument reducer** — the public hook return shape `{ data, isLoading, error, refetch }` is byte-identical (architecture §11.2.4 — "MUST stay byte-identical to today's exported `UseDocumentResult`"). Any user-observable state transition (loading → loaded → error) that an e2e test could assert was already asserted by pre-existing Robot suites against the old implementation. Those suites continue to pass unchanged.

3. **DocumentsTab auto-select effect deletion** — architecture §11.3 documents the one user-observable difference: the URL is no longer auto-written to `?doc=<first>` on initial load. The page still shows the first document in the previewer (behaviour preserved); only the URL stays bare. An e2e test asserting the URL after initial load would need to be RELAXED (not added), and that relaxation is captured in FCT-US010-010's notes on the existing `router.replace` assertion change.

4. **No new visual states** — architecture §11.5 confirms: mermaid output identical, document loading transitions identical, tab navigation identical.

Promoting any of these to e2e would duplicate what existing Robot suites already cover (visual parity is guaranteed by those suites continuing to pass — see architecture §11.5 "No e2e impact — visual parity means existing Robot suites pass unchanged"). Adding a new e2e suite for US010 would create redundant coverage.

**Verdict: No e2e scenarios and no new Robot file. Visual parity for US010 is verified by existing Robot suites continuing to pass after the refactor. The URL initial-load behaviour change (bare URL instead of `?doc=<first>`) is documented in FCT-US010-010 and in US010's sign-off notes for po-ba.**
