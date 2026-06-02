# US010 — `useDocument` → `useReducer` + `DocumentsTab` render-time selection

**Story:** US010 — Fix React-Doctor baseline regressions (top-3 state/effect + 1 security)
**Requirement:** REQ005
**Track:** FE
**Status:** in_review
**Implements:** Scenario: `useDocument` stops cascading state and adjusting state on prop change, Scenario: `DocumentsTab` no longer redirects from inside `useEffect`, Scenario: react-doctor score recovers (the six non-`no-danger` rule IDs portion), Scenario: no regressions in baseline noise (useDocument + DocumentsTab portion), Scenario: tests still pass and hook contract is unchanged
**Blocked by:** US006_fe_harmonise_hooks_on_abortcontroller.md
**Worked-by:** fe-dev-2026-06-02T00:00:00Z-a1b2

## Goal

Refactor `web/hooks/useDocument.ts` from `useState×3` to `useReducer` per architecture §11.2 (state shape `{ data, isLoading, error }`; actions `FETCH_STARTED | FETCH_SUCCEEDED | FETCH_FAILED | ABORTED`; public return shape `{ data, isLoading, error, refetch }` byte-identical). Refactor `web/components/ProjectDetail/DocumentsTab.tsx` per §11.3: delete the auto-select `useEffect` (lines 63–78) entirely; replace the `selectedDocId` computation at line 59 with `const selectedDocId = isBogusDeepLink ? undefined : (docParam ?? documents?.[0]?.id);`. After this task, the six non-`no-danger` rule IDs from architecture §11 (cascading state, adjust-state-on-prop-change, rendering-usetransition-loading, no-event-handler, nextjs-no-client-side-redirect, exhaustive-deps) no longer fire on `useDocument.ts` or `DocumentsTab.tsx`; observable user behaviour is unchanged except for the bare-URL-on-initial-load side-effect (R8, OQ-6 accepted at Rev 3 approval).

## Scope

- **In:** Rewrite `web/hooks/useDocument.ts` per §11.2 — state shape, action types, reducer signature, effect body verbatim from §11.2.1–§11.2.5. Public hook return shape stays `{ data, isLoading, error, refetch }`. AbortController + `latestIdRef` race-safety preserved verbatim (depends on US006's `useDocument` baseline; US006 leaves `useDocument` unchanged structurally, this task mutates it). Edit `web/components/ProjectDetail/DocumentsTab.tsx` per §11.3 — delete lines 63–78, edit line 59 to the new render-time fallback. `handleSelectDoc` (line 88) is UNCHANGED. Update `web/hooks/useDocument.test.ts` with reducer-action-level assertions (additions only) and the test-spec change for `DocumentsTab` documented in §2 US010 row (relax `router.replace`-on-auto-select assertion to `router.replace` only on user click).
- **Out:** MermaidDiagram changes (separate task `US010_fe_mermaid_diagram_ref_attach.md`); changing the public TypeScript shape `UseDocumentResult` (§11.2.4 — locked); adding `hasMore` or pagination to state (§11.2.1 — not applicable, single-document fetch); changing mermaid library; changing AbortController contract from US006 (must be preserved); fixing the 15 lower-severity baseline findings; introducing `useReducer` anywhere else in the codebase.

## Files touched (estimated, exclusive)

- `web/hooks/useDocument.ts`
- `web/hooks/useDocument.test.ts` (additions only)
- `web/components/ProjectDetail/DocumentsTab.tsx`
- `web/components/ProjectDetail/DocumentsTab.test.tsx` (one targeted assertion change per §2 US010 row — relax `router.replace`-on-auto-select to user-click only; NO weakening of behaviour-preserving assertions; tester re-specs this in `fe_unit_tests.md` before code lands)

Independent of `US010_fe_mermaid_diagram_ref_attach.md`. Both US010 tasks `Blocked by:` US006's FE task per §11.4 and are NOT blocked by each other.

## Test contract

The dev must make these tests pass (from `US010_fe_unit_tests.md`, IDs assigned by tester):
- FCT-* reducer-level: `FETCH_STARTED` action clears prior `data` (returns `{ data: null, isLoading: true, error: null }`).
- FCT-* reducer-level: `FETCH_SUCCEEDED` action commits document (returns `{ data: action.document, isLoading: false, error: null }`).
- FCT-* reducer-level: `FETCH_FAILED` action records error (returns `{ data: null, isLoading: false, error: action.error }`).
- FCT-* reducer-level: `ABORTED` action is a no-op on state (returns the previous state unchanged).
- FCT-* public shape: every existing destructure of `{ data, isLoading, error, refetch }` from `useDocument` continues to compile and pass.
- FCT-* AbortController contract preserved: id change `A → B` aborts A, commits B; `controller.signal.aborted` swallow path dispatches `ABORTED`; `latestIdRef` belt-and-braces guard still in place.
- FCT-* `DocumentsTab`: when `?doc=` absent and list non-empty, first item is selected at render-time (`DocumentSidebar` receives `selectedId={documents[0].id}`, `DocumentPreviewer` receives the right document) — same observable outcome as today's auto-select effect.
- FCT-* `DocumentsTab`: bogus deep-link (`?doc=` set, not in list) leaves `selectedDocId = undefined`; previewer renders `isNotFound`.
- FCT-* `DocumentsTab`: `router.replace` is called ONLY when the user clicks a sidebar item (via `handleSelectDoc`) — NOT on initial auto-select (this is the relaxed assertion per §2 US010 row + R8).
- FCT-* `DocumentsTab`: bare URL on initial load is acceptable (OQ-6 accepted) — no `router.replace` fires before user interaction.
- FCT-* react-doctor diff: zero new rule fires; six targeted rules cleared from these files.

If tester surfaces new test IDs beyond these, the dev writes them and flags the addition back to tester.

## Implementation notes

- Architecture §11.2.1 through §11.2.5 is the verbatim TypeScript for state shape, action types, reducer signature, public return shape, and effect body. Copy them; do NOT invent variants.
- Architecture §11.3.2 is the verbatim TSX edit for `DocumentsTab` line 59 and the deletion of lines 63–78.
- `fetchCount` stays as a separate `useState` (NOT folded into reducer state per §11.2.1 default choice).
- `latestIdRef` and `controllerRef` stay as `useRef` (NOT in reducer state).
- The `ABORTED` action MUST be dispatched explicitly in the `.catch` when `controller.signal.aborted`, even though it is a no-op on state. This makes the action set exhaustive and lets tests assert "on abort, no state mutation occurs" (§11.2.3).
- `handleSelectDoc` in `DocumentsTab` is UNCHANGED — it already owns the URL write on user click (§11.3.1).
- §11.3.3 documents that the URL is no longer auto-written on initial load. R8 + OQ-6 — accepted bare-URL initial state per Rev 3 approval. If po-ba pushes back at sign-off, that becomes a separate ticket; do NOT pre-empt with a `useLayoutEffect` workaround.
- Empty-state placeholder ("No documents yet" / "This project has no documents yet") is UNCHANGED.
- Preserve the existing `<DocumentSidebar>` / `<DocumentPreviewer>` JSX structure verbatim; only the `selectedDocId` derivation and the auto-select effect change.
- TDG skill + react-doctor skill MUST be invoked per fe-dev workflow.

### React-Doctor rule IDs that MUST NOT appear in the post-change scan

Per architecture §11 the seven rule IDs targeted across US010 are:

1. `react-doctor/no-danger`
2. `react-doctor/no-cascading-set-state`
3. `react-doctor/no-adjust-state-on-prop-change`
4. `react-doctor/rendering-usetransition-loading`
5. `react-doctor/no-event-handler`
6. `react-doctor/nextjs-no-client-side-redirect`
7. `react-doctor/exhaustive-deps`

This task specifically clears #2, #3, #4 on `web/hooks/useDocument.ts` and #5, #6, #7 on `web/components/ProjectDetail/DocumentsTab.tsx`. Rule #1 (`no-danger`) is cleared by the sibling MermaidDiagram task. After this task lands, the post-change scan must show zero findings on `useDocument.ts` for rules #2–#4 and zero findings on `DocumentsTab.tsx` for rules #5–#7.

Per §11.2.6 — reducer pattern collapses the `setIsLoading(true); setError(null); setData(null)` cascade into one `dispatch({ type: 'FETCH_STARTED' })` (clears #2); prop change triggers ONE dispatch and the reducer derives next state from the action, not from the prop (clears #3); `isLoading` is now a reducer field, not a standalone `useState<boolean>` (clears #4); the deps `[documentId, fetchCount]` are unchanged so #7 passes naturally on this file.

Per §11.3.4 — `nextjs-no-client-side-redirect` cannot fire because the `router.replace` is no longer inside a `useEffect`; `no-event-handler` is unrelated to the auto-select effect (see §11.3.4 note on `refetchList` — verify `onClick={refetchList}` still satisfies the rule; if not, wrap as `onClick={() => refetchList()}`); `exhaustive-deps` cannot fire because the effect is deleted.

### Author-side react-doctor evidence (mandatory)

Run `npx react-doctor@latest --verbose --diff` from `web/` before flipping to `in_review`. Paste the verbatim final score line into `## Notes` below. Score must not regress versus the recorded baseline (92/100); no NEW rule fires in the diff; rules #2–#7 must be gone from the two changed files.

## Definition of done

- All listed tests green.
- `cd web && npm run typecheck && npm test -- --watchAll=false --forceExit` clean.
- `cd web && npm test -- --coverage --watchAll=false --forceExit` — every file in `## Files touched` clears ≥ 80 % per-file line coverage OR a `## Coverage exemption` block here justifies the gap.
- `npx react-doctor scan web/` reports zero findings on `useDocument.ts` for rules #2–#4 and zero findings on `DocumentsTab.tsx` for rules #5–#7.
- `npx react-doctor --diff` against baseline shows only removals (no NEW rule fires).
- `UseDocumentResult` public shape is byte-identical to today; every existing consumer (`DocumentsTab.tsx` line 81 originally, now wherever it lands after the line-59 edit) compiles.
- AbortController contract from US006 preserved verbatim (race-safety unchanged).
- `scripts/review/run-gate.sh fe` exits with `REVIEW GATE: PASS`.
- `scripts/review/run-gate.sh cross` exits with `REVIEW GATE: PASS`.
- Code matches architecture §11.2 and §11.3 contracts end-to-end.
- `## Notes` contains the verbatim `react-doctor --verbose --diff` final score line.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Notes

### Implementation pass 1

**react-doctor `--diff` final score:** `100 / 100 Great — No issues found!` (React Doctor v0.2.16, scanning changes: worktree-agent-abfedecc9649afcb3 → main)

Score improved from 92/100 baseline to 100/100. No new errors, no new warnings introduced by this diff.

**Files touched:**
- `web/hooks/useDocument.ts` — refactored from `useState×3` to `useReducer` per §11.2. State `{ data, isLoading, error }`, actions `FETCH_STARTED|FETCH_SUCCEEDED|FETCH_FAILED|ABORTED`. Public shape `{ data, isLoading, error, refetch }` unchanged. AbortController + `latestIdRef` race-safety preserved verbatim from US006.
- `web/hooks/useDocument.test.ts` — added FCT-US010-005 through FCT-US010-009 (reducer action assertions + public shape check).
- `web/components/ProjectDetail/DocumentsTab.tsx` — deleted auto-select `useEffect` (lines 63–78); replaced `selectedDocId` with render-time fallback `docParam ?? documents?.[0]?.id` per §11.3.2; added `type="button"` + `onClick={() => refetchList()}` to Retry button.
- `web/components/ProjectDetail/DocumentsTab.test.tsx` — relaxed FCT-US002-003 per R8 (no more router.replace on mount assertion); added FCT-US010-010 through FCT-US010-012.

**Tests added:** 8 new tests (FCT-US010-005 through FCT-US010-012)

**Test counts:** 125 tests passing, 17 suites, 0 failures

**react-doctor rules cleared:**
- `no-cascading-set-state` — gone from `useDocument.ts` (single `dispatch` replaces triple `setState` cascade)
- `no-adjust-state-on-prop-change` × 3 — gone from `useDocument.ts` (reducer derives state from action, not prop)
- `rendering-usetransition-loading` — gone from `useDocument.ts` (`isLoading` is now reducer field, not `useState<boolean>`)
- `no-event-handler` — `onClick={refetchList}` wrapped as `onClick={() => refetchList()}`; `type="button"` added
- `nextjs-no-client-side-redirect` — gone from `DocumentsTab.tsx` (auto-select `useEffect` deleted)
- `exhaustive-deps` — gone from `DocumentsTab.tsx` (effect with `// eslint-disable-line` deleted)

**CSR invariants:** `grep -RE 'getServerSideProps|getStaticProps|getInitialProps' web/pages/` returns comment-only hit, no actual SSR.

**Gate results:** `scripts/review/run-gate.sh fe` → `REVIEW GATE: PASS`; `scripts/review/run-gate.sh cross` → `REVIEW GATE: PASS`

**Side effect documented:** URL no longer auto-written on initial load (bare `/projects/:id?tab=documents` until user clicks sidebar item). This is the accepted behavior per OQ-6 / R8.

**FCT-US010-013 (react-doctor meta-assertion):** Verified — zero findings on `useDocument.ts` for rules #2–#4; zero findings on `DocumentsTab.tsx` for rules #5–#7. Score 100/100.

## Review log

(tech-lead appends here on each review pass)
