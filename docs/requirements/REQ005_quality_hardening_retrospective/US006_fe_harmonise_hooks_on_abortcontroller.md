# US006 — Harmonise `useProject`, `useProjectDocuments` on AbortController + signal-threaded `lib/api/`

**Story:** US006 — Harmonise `useProject`, `useProjectDocuments`, `useDocument` on AbortController + signal-threaded `lib/api/`
**Requirement:** REQ005
**Track:** FE
**Status:** pending
**Implements:** Scenario: `fetchProject` accepts `signal?: AbortSignal`, Scenario: all three `lib/api/` fetch functions accept `signal`, Scenario: `useProject` aborts on unmount and on id change, Scenario: `useProjectDocuments` aborts on unmount and on projectId change, Scenario: `useDocument` pattern unchanged, Scenario: aborted requests do NOT update state, Scenario: aborted-request errors are silently ignored (not surfaced as `error`), Scenario: existing tests still pass, Scenario: no end-user-visible behaviour change in the happy path
**Blocked by:** none
**Worked-by:** _(none)_

## Goal

Refactor `useProject` and `useProjectDocuments` from the `mounted` flag pattern to the AbortController + `latestIdRef` pattern that `useDocument` already exemplifies. Update `web/lib/api/projects.ts` so both `fetchProject` and `fetchProjects` accept an optional trailing `signal?: AbortSignal` parameter, matching `fetchProjectDocuments` / `fetchDocument`. After this task, all three project/document fetch hooks have race-safe cancellation and identical structure (modulo `useDocument`'s upcoming `useReducer` refactor in US010).

## Scope

- **In:** Edit `web/lib/api/projects.ts` to add `signal?: AbortSignal` as the last parameter to both `fetchProject(id, signal?)` and `fetchProjects(signal?)` and forward via `{ signal }` to `fetchClient`. Rewrite `web/hooks/useProject.ts` per architecture §4.3 (AbortController + `latestIdRef` + `controllerRef`). Rewrite `web/hooks/useProjectDocuments.ts` per §4.4 (same pattern, preserving `refetch()` via `fetchCount`). Add unit-test additions (NOT rewrites) for the abort semantics in both hooks' `.test.ts` files. Confirm `web/lib/api/documents.ts` and `web/lib/api/client.ts` need no code change (already correct per §2 US006 row).
- **Out:** Refactoring `useDocument` (it is the reference implementation — US010 covers its `useReducer` refactor and Depends-on this task per §11.4); extracting a shared `useFetch<T>` hook (D-005 — explicitly deferred); refactoring `useProjects` beyond the signature uniformity of `fetchProjects` (D-005); SSR concerns; any visual or UX change to `pages/projects/[id].tsx` or the documents tab.

## Files touched (estimated, exclusive)

- `web/lib/api/projects.ts`
- `web/hooks/useProject.ts`
- `web/hooks/useProjectDocuments.ts`
- `web/hooks/useProject.test.ts` (additions only, no rewrites of existing assertions)
- `web/hooks/useProjectDocuments.test.ts` (additions only)

US010's two FE tasks `Blocked by:` this task. This is NOT a scaffold task in the orchestrator sense (no shared file like `web/package.json` or `lib/api/types.ts` is modified), but it is the **gate task** for US010 per architecture §11.4.

## Test contract

The dev must make these tests pass (from `US006_fe_unit_tests.md`, IDs assigned by tester — FCT-*):
- FCT-* for `useProject`: id change `A → B` before A resolves results in A's request aborted, B's data committed; aborted A does NOT set `error`; cleanup on unmount calls `controller.abort()`.
- FCT-* for `useProject`: `latestIdRef` belt-and-braces guard prevents stale resolution clobber even if abort somehow fails.
- FCT-* for `useProjectDocuments`: projectId change `X → Y` before X resolves results in X aborted, Y committed; `refetch()` increments `fetchCount`, aborts the in-flight previous controller, runs a fresh fetch.
- FCT-* for both: aborted-request `AbortError` is silently swallowed (no `error` state, no console log, no `isLoading` flicker).
- FCT-* for `lib/api/projects.ts`: `fetchProject(id)` (no signal) still compiles and works (backwards-compat); `fetchProject(id, signal)` forwards `{ signal }` to `fetchClient`; same for `fetchProjects(signal?)`.
- FCT-* for `useDocument`: existing tests pass unchanged (regression guard for the reference impl).

If tester surfaces new test IDs beyond these, the dev writes them and flags the addition back to tester.

## Implementation notes

- Architecture §4.3 is the verbatim TypeScript for `useProject`. Architecture §4.4 is the verbatim TypeScript for `useProjectDocuments`. Copy them; do NOT invent variants. Existing `useDocument` (already correct) is the runtime reference.
- §4.5 locks AbortError handling rules: swallow on `controller.signal.aborted`, do not discriminate by `err.name === 'AbortError'`, layer `latestIdRef` / `latestKeyRef` belt-and-braces.
- `fetchClient` already plumbs `signal` per §2 US006 row note — verify no change is needed; if you find a hole, raise `ARCHITECTURE_GAP_FOUND`.
- MSW handler additions for delay-based abort assertions: `await delay(100)` shape; reuse `web/test/msw/` handlers; do not weaken existing handlers.
- React 18 strict mode double-invocation: the existing `useDocument` tests already handle this; copy the pattern for the new abort assertions.
- TDG skill + react-doctor skill MUST be invoked per fe-dev workflow.
- Author-side react-doctor evidence (mandatory, per fe-dev skill contract): run `npx react-doctor@latest --verbose --diff` from `web/`, paste the final score line into `## Notes` below before flipping to `in_review`. Score must not regress versus the recorded baseline (92/100, 4 errors, 19 warnings); no NEW rule fires.

## Definition of done

- All listed tests green.
- `cd web && npm run typecheck && npm test -- --watchAll=false --forceExit` clean.
- `cd web && npm test -- --coverage --watchAll=false --forceExit` — every file in `## Files touched` clears ≥ 80 % per-file line coverage OR a `## Coverage exemption` block here justifies the gap.
- No `any` types added without justification; types align with architecture §4.2 signatures.
- `pages/projects/[id].tsx`, dashboard, documents tab render identically to before — confirmed by existing Jest tests still passing.
- `scripts/review/run-gate.sh fe` exits with `REVIEW GATE: PASS`.
- `scripts/review/run-gate.sh cross` exits with `REVIEW GATE: PASS`.
- Code matches architecture §4 contract end-to-end.
- `## Notes` contains the verbatim `react-doctor --verbose --diff` final score line.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Notes

(fe-dev pastes `react-doctor --verbose --diff` final score line here before flipping to `in_review`.)

## Review log

(tech-lead appends here on each review pass)
