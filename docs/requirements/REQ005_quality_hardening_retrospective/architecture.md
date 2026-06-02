# Architecture — REQ005 Quality Hardening Retrospective

**Approval:** approved
**Approved-by:** human
**Approved-at:** 2026-06-02T02:32:14Z

---

## 0. Reading guide

This is an architecture document for a **quality-hardening / dev-experience retrospective**. It is intentionally light on the usual sections (no new endpoints, no new pages, no new microservices, no data-model changes). The substance is in:

- **§2 File-level touch map** — what changes where, per story.
- **§3 Signal-cancellable lifecycle context contract** (US004) — the one new shared BE pattern.
- **§4 AbortController hook contract** (US006) — the one FE pattern being harmonised.
- **§5 Test backfill matrix** (US005) — the 16 test names and their target branches.
- **§6 e2e stack-up shape** (US008) — `docker-compose.yml` + Makefile + seed contract.
- **§7 Worktree origin contract** (US009) — path choice + fallback wording.
- **§9 Decisions** — restated D-001 through D-007 with verdicts + the new D-008..D-014.
- **§10 Open-question resolutions** — Q1/Q2/Q3 from po-ba's README answered.
- **§11 React-Doctor regression fixes** (US010) — MermaidDiagram ref-attach path, `useDocument` reducer contract, `DocumentsTab` render-time selection contract, and US006↔US010 ordering.

### 0.1 Executive summary (one-screen scan for re-approval)

This revision (rev 2, 2026-06-02) folds in **US010 — React-Doctor baseline regression fixes** per README D-008. Headline calls the human needs to confirm:

1. **MermaidDiagram fix path = ref-attach (NOT DOMPurify + suppression).** `react-doctor/no-danger` fires on the `dangerouslySetInnerHTML` token regardless of sanitisation, so DOMPurify alone does not clear the rule — it would require an inline lint suppression and still leaves an XSS-shaped construct. Ref-attach (parse `svg` via `DOMParser`, append `<svg>` node to a `ref`'d wrapper inside `useEffect`) clears the rule cleanly, adds zero dependencies, and is straightforward under React 18 strict mode. See §11.1 for the exact `useEffect` shape.
2. **`useDocument` reducer contract is frozen here.** State `{ data, isLoading, error }`, actions `FETCH_STARTED | FETCH_SUCCEEDED | FETCH_FAILED | ABORTED`. The hook's **public return shape stays `{ data, isLoading, error, refetch }`** so existing Jest consumers and `DocumentsTab` keep compiling. See §11.2.
3. **`DocumentsTab` drops the auto-select `useEffect` entirely.** The current effect's redirect intent is replicated at render time as `selectedDocId = docParam ?? documents[0]?.id`; the user-click URL write continues to live in `handleSelectDoc` (already exists). Empty-state placeholder remains the existing "No documents yet" / "This project has no documents yet". No `ARCHITECTURE_GAP_FOUND` needed — the sibling click handler already exists. See §11.3.
4. **US006 ⇄ US010 ordering: US006 lands first, US010 adds `Depends-on: US006`.** Landing the AbortController + signal-thread pattern first keeps US010's diff smaller; the reducer action set already names `ABORTED` so the merge is mechanical. See §11.4.
5. **No new product behaviour; visual parity required.** Mermaid output, document loading observable transitions, and tab navigation are byte-for-byte identical to today. See §11.5 for the test-time guarantees that catch regressions.

New open question OQ-5 (§13.2) asks the human to confirm the ref-attach default over DOMPurify+suppression.

---

## 1. Scope

### 1.1 In scope

This requirement closes nine discrete tech-debt items raised by the REQ004 quality audit. All work is **infrastructure / developer-experience / test debt**. There is **NO new product feature, NO new endpoint, NO new page, NO architectural rewrite, NO database-schema change**.

In-scope changes, grouped:

- **A. Quality-gate must-fix** (US001, US002, US003) — edits to `scripts/review/run-gate.sh` and `scripts/review/README.md` only.
- **B. Code-level tech debt** (US004, US005, US006, US007) — two Go `main.go` files, two Go repo `_test.go` files, three FE hooks + two FE API client files, one `package.json` + lockfile.
- **C. Workflow / harness** (US008, US009) — new top-level `docker-compose.yml`, new top-level `Makefile`, new `services/agent-board/Dockerfile`, optional new `web/Dockerfile`, new `tests/e2e/data/seeds/` directory, `tests/e2e/README.md` runbook, and either harness-side worktree origin fix OR a canonical-path-policy block appended to all six agent-definition files.

### 1.2 Out of scope (anti-scope)

Restating from po-ba README so the architect / tech-lead / devs share one boundary:

- **No new endpoints.** The existing REQ001–REQ004 API surface (the four `/api/v1/projects*` and `/api/v1/documents*` routes; the MCP `/sse` + `/message` routes) is unchanged in path, request shape, response shape, and status codes. **§8 API-contract impact is therefore N/A.**
- **No new pages or hooks beyond the harmonisation.** `useProjects` keeps its `mounted`-flag implementation (only its underlying `fetchProjects` gains a `signal?` param for API-uniformity).
- **No production code change in `services/agent-board/internal/repo/`.** US005 is test-only.
- **No CI runner hookup, no cloud test environment, no rewriting existing `.robot` files.** US008 delivers a local-Docker stack; CI is a follow-up REQ.
- **No new shared `useFetch<T>` extraction.** Per D-005, copy the AbortController pattern three times rather than invent a new abstraction in the same story.
- **No MSW leak root-cause investigation.** Per D-001 the `--forceExit` flag is the must-fix; the actual leak hunt is deferred (see §10 Q1 — recommended deferral to REQ006).
- **No `--detectOpenHandles` in the production gate.** That flag belongs in the leak-hunt story when it lands.
- **No sweep of `context.Background()` in test files.** Tests legitimately use it for sqlmock / httptest setup (D-003).
- **No graceful HTTP shutdown wiring.** US004 only fixes the boot-time DB ping (see D-009).
- **No migration mechanism change.** US008 reuses `psql -f` on raw migration SQL (D-010); replacing with `golang-migrate` is a future REQ.

---

## 2. File-level touch map

This is the authoritative list of files each story touches. Tech-lead uses this to scope tasks; devs use it to know whether they're in their lane. **Paths in bold are NEW files; the rest are edits.**

### US001 — `printf "--"` bug

| File | Change |
|---|---|
| `scripts/review/run-gate.sh` | Edit `printf` at line 58 (in `run_check`) and line 71 (in `run_check_warn`) — prefix format string with `-- ` so `printf` does not parse a leading `--` as an option. |

### US002 — `--forceExit` on FE Jest

| File | Change |
|---|---|
| `scripts/review/run-gate.sh` | Edit line 116 (`run_check "npm test (--watchAll=false)" ...`) to pass `--forceExit` after `--watchAll=false`. |

### US003 — Soft-warn missing `gosec` / `govulncheck`

| File | Change |
|---|---|
| `scripts/review/run-gate.sh` | In `gate_be()`: replace `require_tool gosec ...` and `require_tool govulncheck ...` with inline `command -v` checks. If missing, print one WARN line (consistent with `run_check_warn` output prefix), DO NOT call the corresponding `run_check`. `require_tool go` and `require_tool golangci-lint` remain hard. (See §9 D-011 for helper-naming guidance.) |
| `scripts/review/README.md` | Add a "Soft-warn vs. hard-required" subsection under "What it runs" listing which tools are hard (go, golangci-lint, npm, semgrep, gitleaks) vs soft-warn (gosec, govulncheck) and the install one-liners for the soft-warn pair. |

### US004 — Signal-cancellable DB-ping context

| File | Change |
|---|---|
| `services/agent-board/cmd/api-server/main.go` | Inside `run()`: build lifecycle context via `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)`; derive `pingCtx` with `context.WithTimeout(lifecycleCtx, 5*time.Second)`; pass `pingCtx` to `db.PingContext`. Replace existing `db.PingContext(context.Background())` at line 52. Add `defer stop()` and `defer cancel()` in the correct order (see §3.3). No other behaviour change. |
| `services/agent-board/cmd/mcp-server/main.go` | Same pattern as api-server, replacing line 37's `db.PingContext(context.Background())`. |

NOTE: We do NOT extract a shared `internal/lifecycle/` helper in this REQ (D-008). Two short copies are cheaper than introducing a new internal package + its tests for nine lines of glue.

### US005 — Repo error-branch tests

| File | Change |
|---|---|
| `services/agent-board/internal/repo/document_repo_test.go` | Add 8 test functions (names in §5). Test-only; no edits to `document_repo.go`. |
| `services/agent-board/internal/repo/project_repo_test.go` | Add 8 test functions (names in §5). Test-only; no edits to `project_repo.go`. |

### US006 — Harmonise FE hooks on AbortController

| File | Change |
|---|---|
| `web/lib/api/projects.ts` | `fetchProject(id, signal?)` — add `signal?: AbortSignal`, forward as `{ signal }` to `fetchClient`. `fetchProjects(signal?)` — same signature change for uniformity even though `useProjects` will not use it. |
| `web/lib/api/documents.ts` | No code change required (signatures already accept `signal`); confirm in PR. |
| `web/lib/api/client.ts` | No code change required (signal already plumbed). Confirm in PR. |
| `web/hooks/useProject.ts` | Rewrite from `mounted` flag → AbortController + `latestIdRef` pattern (see §4 contract). |
| `web/hooks/useProjectDocuments.ts` | Same — `mounted` flag → AbortController + `latestIdRef`. `refetch()` semantics preserved (still `setFetchCount(c => c+1)`; the new effect creates a new controller that aborts the previous). |
| `web/hooks/useDocument.ts` | No structural change. This is the reference implementation. |
| `web/hooks/useProjects.ts` | No structural change. `fetchProjects()` call site may pass `undefined` or omit `signal` — both work. |
| `web/hooks/useProject.test.ts`, `useProjectDocuments.test.ts` | Tester (Phase 2) will add abort-semantics test cases — story-level additions, no rewrites of existing assertions. |

### US007 — Move `@testing-library/dom` to devDependencies

| File | Change |
|---|---|
| `web/package.json` | Delete the `@testing-library/dom: "^10.4.1"` line from `dependencies`; add identical line to `devDependencies` (keep alphabetical sort within the block). |
| `web/package-lock.json` | Regenerate via `cd web && npm install` after the package.json edit. Verify diff is limited to the move + any incidental peer rewiring; if diff is large, investigate before committing. |

### US008 — Live e2e stack-up

| File | Change |
|---|---|
| **`/docker-compose.yml`** | New file at repo root. Services: `postgres`, `api-server`, `web` (containerised — D-012). Healthchecks on postgres + api-server. Bind-mount source for fast iteration. See §6.2 for service definitions. |
| **`/Makefile`** | New file at repo root. Targets: `e2e-up`, `e2e-down`, `e2e-seed`, `e2e-run`, `e2e`, `e2e-logs`. See §6.3 for exact target shapes. |
| **`/services/agent-board/Dockerfile`** | New file. Multi-stage Go build (golang:1.x-alpine builder → distroless/alpine runtime). Same binary used for both `api-server` and `mcp-server` (passed `CMD` overrides the entry). |
| **`/web/Dockerfile`** | New file. Multi-stage Node build (node:20-alpine deps → node:20-alpine runner). Runs `npm run build && npm start` for prod-like local; alternative shape `next dev` for hot-reload is a docker-compose `command:` override. D-012 picks built. |
| **`/.dockerignore`** | New file at repo root listing `node_modules/`, `.next/`, `services/agent-board/bin/`, `*.log`, `tests/e2e/results/`. |
| **`/tests/e2e/data/seeds/`** | New directory. |
| **`/tests/e2e/data/seeds/REQ000_baseline.sql`** | New file — example seed (one project named "Sample Project" with two documents) demonstrating the pattern. |
| **`/tests/e2e/README.md`** | New file. Runbook: prerequisites, target list, how to add a seed, how to debug a failing Robot run, orchestrator responsibility in Phase 3c. |
| **`/.gitignore`** | Add `tests/e2e/results/`. (Already covers `*.pid`, `*.log`, root-level Robot dumps.) |

NOTE: `services/agent-board/cmd/mcp-server` does NOT need to be in the e2e compose stack — UI e2e tests (Robot + Browser library) only hit `web` and `api-server`. MCP-server can be added later if MCP-driven seeding becomes the preferred path (D-010 says SQL for now).

### US010 — React-Doctor baseline regressions (top-3 state/effect + 1 security)

| File | Change |
|---|---|
| `web/components/ProjectDetail/MermaidDiagram.tsx` | Replace the `dangerouslySetInnerHTML` branch (line 130–142, the `renderState.status === 'success'` JSX) with a `ref`-attached wrapper `<div>`. New `useEffect` that runs when `renderState.status === 'success'` parses `svg` via `DOMParser`, extracts the `<svg>` element, clears the wrapper's current children, and appends the parsed node. Removes both the inline `// dangerouslySetInnerHTML is the single sanctioned use` comment (lines 137–138) and the `dangerouslySetInnerHTML` prop itself. NO change to the lazy-load contract, the unique-id contract, the error contract, or the `setRenderState({ status: 'success', svg, ariaLabel })` payload. See §11.1. |
| `web/hooks/useDocument.ts` | Refactor the 7-`setState` cascade in `useEffect` (lines 50–89) to a single `useReducer` with action types `FETCH_STARTED \| FETCH_SUCCEEDED \| FETCH_FAILED \| ABORTED`. **Public return shape stays `{ data, isLoading, error, refetch }`** — type alias `UseDocumentResult` unchanged. AbortController + `latestIdRef` race-safety preserved verbatim from US006's contract. See §11.2. |
| `web/components/ProjectDetail/DocumentsTab.tsx` | Delete the auto-select `useEffect` block (lines 63–78) entirely. Replace the `selectedDocId = isBogusDeepLink ? undefined : docParam` line (line 59) with `selectedDocId = isBogusDeepLink ? undefined : (docParam ?? documents?.[0]?.id)`. The render-time fallback preserves the "first item selected on load" UX without any redirect. The existing `handleSelectDoc` click handler (line 88) continues to own the URL write on user interaction; no new handler. **No** sibling escalation needed — `handleSelectDoc` is the existing click-driven counterpart, so user-triggered URL sync is already covered. See §11.3. |
| `web/package.json` | NO change — ref-attach path adds zero new dependencies. (If a future revision overrides §11.1 to pick DOMPurify, this row becomes `+ "isomorphic-dompurify": "^2.x"` in `dependencies`. Locked OUT for now.) |
| `web/components/ProjectDetail/MermaidDiagram.test.tsx` | Existing assertion (today reads `container.querySelector('svg')` or similar) continues to work — the rendered DOM still contains an `<svg>` child. Tester adds a new FCT-* assertion in `fe_unit_tests.md` that confirms `container.querySelector('[dangerouslySetInnerHTML]')` is null and that `svg.outerHTML` matches mermaid's output. Test additions only; no rewrites. |
| `web/hooks/useDocument.test.ts` | Existing assertions on `{ data, isLoading, error, refetch }` continue to pass byte-for-byte (public surface unchanged). Tester MAY add a reducer-level unit assertion that confirms `FETCH_STARTED` clears prior data, `FETCH_SUCCEEDED` commits, `FETCH_FAILED` records error, and `ABORTED` is a no-op on state. No existing assertions weakened. |
| `web/components/ProjectDetail/DocumentsTab.test.tsx` | Existing assertion that "when `?doc=` is absent and the list is non-empty, the first item is selected" continues to pass — selection is now render-time rather than effect-driven, but the observable outcome (`DocumentSidebar` receives `selectedId={documents[0].id}`, `DocumentPreviewer` receives `document=…`) is identical. The existing assertion that `router.replace` was called on auto-select **must be relaxed** to "router.replace is called only when the user clicks a sidebar item" — tester updates the spec in `fe_unit_tests.md`. This is the one external test-surface change US010 introduces; it is behaviour-preserving for the user but observable to the test. Document it in the test spec. |

NOTE: US010 is FE-only. No BE files are touched. Tracks: FE.

### US009 — Worktrees branch off local `main`

**Path choice (see §7):** Path (b) — agent-definition documentation fallback. Path (a) is **not reachable from this repo** (the worktree harness lives outside the repo, in the Claude Code runtime layer).

| File | Change |
|---|---|
| `.claude/agents/po-ba.md` | Append the `## Canonical-path edit policy (worktree workaround)` block from §7.2. |
| `.claude/agents/system-architect.md` | Same block, identical wording. |
| `.claude/agents/tech-lead.md` | Same block. |
| `.claude/agents/tester.md` | Same block. |
| `.claude/agents/be-dev.md` | Same block. |
| `.claude/agents/fe-dev.md` | Same block. |
| `docs/requirements/REQ005_quality_hardening_retrospective/README.md` | Append `### Decision: US009 path = (b)` entry to the Decision log section, stating that path (a) was not reachable and naming the agent files updated. |
| `CLAUDE.md` (= `AGENTS.md`) | Optional one-sentence pointer under "Orchestrator cheat sheet": "Sub-agent worktrees may branch off stale `origin/main`; agents follow the canonical-path edit policy in their own definition files when touching `docs/requirements/`, `tests/e2e/`, `.claude/agents/`, or `CLAUDE.md`/`AGENTS.md`." Tech-lead decides whether to add (low-priority documentation gloss). |

---

## 3. Signal-cancellable lifecycle context contract (US004)

### 3.1 Goals

1. The DB ping must NOT block indefinitely on a stuck connection. **Bound it.**
2. SIGTERM / SIGINT during boot must cancel the ping and abort startup cleanly. **Wire signals.**
3. The lifecycle context must be the parent of any future cancellable work in `run()` (for now: only the ping). **One root context.**
4. No regression to the existing happy path.

### 3.2 Signals handled

- `os.Interrupt` (= SIGINT — Ctrl+C in a foreground shell).
- `syscall.SIGTERM` (= `kill -TERM`, Kubernetes pod termination, Docker stop).

SIGHUP, SIGUSR1, SIGUSR2 are explicitly NOT handled (out of scope; no documented behaviour today).

### 3.3 DB-ping timeout value: **5 seconds** (hard-coded)

Per D-013, hard-code `5 * time.Second`. Rationale:

- Audit data shows pings against a healthy local Postgres complete in <100 ms.
- A 5-s ceiling catches a stuck network handshake well before any reasonable orchestrator's health-check grace period.
- An env-var (`DB_PING_TIMEOUT_SECONDS`) is **optional** per US004 notes; we leave it out for now. If ops later wants tuning, adding `os.Getenv("DB_PING_TIMEOUT_SECONDS")` + `strconv.Atoi` with a 5-s fallback is a trivial follow-up that does NOT require an architecture revision.

Marker: a `// TODO(REQ005): make configurable if ops needs it` comment goes on the timeout literal.

### 3.4 `run()` shape (both `api-server` and `mcp-server`)

The pattern below replaces the existing DB-ping block in both `cmd/api-server/main.go` (around line 52) and `cmd/mcp-server/main.go` (around line 37). **The two files do NOT share a helper** — copy-paste is intentional (D-008).

```go
import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"
    // existing imports
)

func run() error {
    // ... existing config / DB-URL parsing / sql.Open / defer db.Close() ...

    // Lifecycle context — cancelled on SIGINT or SIGTERM. Use as the parent for
    // any cancellable startup work (currently: the DB ping). Surviving server
    // loop (e.Start) is NOT yet wired to this context — see D-009.
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    // Bound the DB ping. TODO(REQ005): make configurable if ops needs it.
    pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    if err := db.PingContext(pingCtx); err != nil {
        return fmt.Errorf("db ping failed: %w", err)
    }

    // ... existing handler-registration + e.Start(...) ...
}
```

### 3.5 Defer order (cleanup order)

The architecture-mandated order on `run()` exit, top-down:

1. **`defer cancel()`** (innermost) — release the WithTimeout context's resources.
2. **`defer stop()`** — unregister the signal handler so a second SIGTERM during teardown does not reach our handler.
3. **`defer func() { _ = db.Close() }()`** (existing) — close the DB pool.

Go runs deferred calls LIFO, so the **declaration order in source MUST be**: `db.Close` first, then `stop`, then `cancel`. The current code already has `defer db.Close()` before the ping; new declarations come after.

Actual call order at exit becomes: cancel → stop → db.Close. Correct (release context-bound resources before tearing the pool).

### 3.6 What this contract does NOT cover

- **Graceful HTTP shutdown (`e.Shutdown(ctx)`).** Wiring the surviving Echo loop to `ctx` for in-flight-request drain on SIGTERM is **NOT in REQ005**. Per D-009 this is its own concern with its own design surface (Echo shutdown semantics, drain timeout, port re-use across restarts). REQ005 only fixes boot-time ping; the long-lived `e.Start(":" + port)` call remains a blocking foreground loop that exits on a fatal serve error. Anyone wanting graceful shutdown raises a new REQ.
- **`internal/lifecycle/` helper package.** Two call sites do not justify a new package (D-008).

### 3.7 Testability notes (for tester to spec)

The signal-handler unit test pattern: in a `t.Run` subtest, declare a `ctx, stop := signal.NotifyContext(...)`, spawn a goroutine that `syscall.Kill(syscall.Getpid(), syscall.SIGTERM)` after 50 ms, assert `<-ctx.Done()` returns within 1 s. For the timeout branch, use a `sqlmock`-driven `driver.Conn` whose `Ping(ctx)` blocks on `<-ctx.Done()` and assert the returned error matches `context.DeadlineExceeded` (via `errors.Is`). Both belong in `cmd/api-server` + `cmd/mcp-server` test packages — they live under each `cmd/*/main_test.go` (new files), not in a separate `internal/lifecycle/` package.

---

## 4. AbortController hook contract (US006)

### 4.1 Reference implementation

`web/hooks/useDocument.ts` — already correct. Do not regress it. The pattern below is a verbatim copy of its structure, restated as a contract so `useProject` and `useProjectDocuments` are written to match field-for-field.

### 4.2 `lib/api/` function signatures (frozen)

After US006, every `lib/api/` fetcher has uniform signature shape:

```ts
// web/lib/api/projects.ts
export const fetchProjects = (
  signal?: AbortSignal
): Promise<ProjectsResponse>;

export const fetchProject = (
  id: string,
  signal?: AbortSignal
): Promise<Project>;

// web/lib/api/documents.ts (already correct — no code change)
export const fetchProjectDocuments = (
  projectId: string,
  signal?: AbortSignal
): Promise<DocumentsListResponse>;

export const fetchDocument = (
  documentId: string,
  signal?: AbortSignal
): Promise<Document>;
```

`signal` is always the **last** positional parameter and always **optional**. Existing callers that omit it continue to compile.

### 4.3 Hook contract — `useProject(id)` after US006

```ts
export interface UseProjectResult {
  data: Project | null;
  isLoading: boolean;
  error: ApiError | Error | null;
  isNotFound: boolean;
}

export const useProject = (id: string | undefined): UseProjectResult => {
  const [data, setData] = useState<Project | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<ApiError | Error | null>(null);
  const [isNotFound, setIsNotFound] = useState<boolean>(false);

  const controllerRef = useRef<AbortController | null>(null);
  const latestIdRef = useRef<string | undefined>(undefined);

  useEffect(() => {
    if (id === undefined) return;

    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;
    latestIdRef.current = id;

    setIsLoading(true);
    setError(null);
    setIsNotFound(false);
    setData(null);

    fetchProject(id, controller.signal)
      .then((project) => {
        if (latestIdRef.current === id) {
          setData(project);
          setIsLoading(false);
        }
      })
      .catch((err: unknown) => {
        if (controller.signal.aborted) return; // swallow
        if (latestIdRef.current !== id) return; // stale
        if (err instanceof ApiError) {
          setError(err);
          if (err.code === 'NOT_FOUND') setIsNotFound(true);
        } else if (err instanceof Error) {
          setError(err);
        } else {
          setError(new Error('Failed to load project'));
        }
        setIsLoading(false);
      });

    return () => {
      controller.abort();
    };
  }, [id]);

  return { data, isLoading, error, isNotFound };
};
```

### 4.4 Hook contract — `useProjectDocuments(projectId)` after US006

Identical pattern, with the existing `refetch()` machinery preserved:

```ts
export interface UseProjectDocumentsResult {
  data: DocumentsListResponse | null;
  isLoading: boolean;
  error: ApiError | Error | null;
  refetch: () => void;
}

export const useProjectDocuments = (
  projectId: string | undefined
): UseProjectDocumentsResult => {
  const [data, setData] = useState<DocumentsListResponse | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<ApiError | Error | null>(null);
  const [fetchCount, setFetchCount] = useState(0);

  const controllerRef = useRef<AbortController | null>(null);
  const latestKeyRef = useRef<string | undefined>(undefined);

  const refetch = useCallback(() => {
    setFetchCount((c) => c + 1);
  }, []);

  useEffect(() => {
    if (projectId === undefined) return;

    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;
    const key = projectId; // dep that identifies "this fetch"
    latestKeyRef.current = key;

    setIsLoading(true);
    setError(null);

    fetchProjectDocuments(projectId, controller.signal)
      .then((result) => {
        if (latestKeyRef.current === key) {
          setData(result);
          setIsLoading(false);
        }
      })
      .catch((err: unknown) => {
        if (controller.signal.aborted) return; // swallow
        if (latestKeyRef.current !== key) return; // stale
        if (err instanceof ApiError) setError(err);
        else if (err instanceof Error) setError(err);
        else setError(new Error('Failed to load documents'));
        setIsLoading(false);
      });

    return () => {
      controller.abort();
    };
  }, [projectId, fetchCount]);

  return { data, isLoading, error, refetch };
};
```

### 4.5 AbortError handling — locked rules

1. If `controller.signal.aborted === true` when `.catch` fires, **swallow the error**. Do not set `error` state. Do not log to console. Aborts are control flow, not failures.
2. The fetch implementation (browser `fetch`) throws a `DOMException` with `name === 'AbortError'`. Our `lib/api/client.ts` does not currently special-case it — it propagates as a generic `Error`. The hook's `controller.signal.aborted` check above catches it without needing to discriminate by name. **Do not** add `if (err.name === 'AbortError')` checks; rely on `controller.signal.aborted`.
3. A `latestKeyRef`/`latestIdRef` belt-and-braces check after the network layer catches the rare case where `.then` resolves between the dep change and the next `controllerRef.current?.abort()` call.

### 4.6 What this contract does NOT change

- `useProjects` — keeps `mounted` flag; only `fetchProjects` signature changes (D-005). Dashboard list usage is not at risk of rapid dep change.
- Shared `useFetch<T>(key, fetcher)` extraction — explicitly deferred (D-005). The three hooks may end up looking similar; that is acceptable.
- Existing test cases — additions only, no rewrites (tester will spec the abort assertions).

---

## 5. Test backfill matrix (US005)

The audit specifies 14 backfill tests. Per po-ba D-004 and US005 AC, we land **16 functions** (the Update method is split into `_NotFound` + `_GenericError` on both repos, mirroring the source code's two error branches). Names are **verbatim** as the audit and US005 AC list them.

### 5.1 `services/agent-board/internal/repo/document_repo_test.go` — 8 new

| # | Test function | Repo method | Branch covered | Mock shape |
|---|---|---|---|---|
| D1 | `TestDocumentRepo_CreateDocument_GenericError` | `CreateDocument` | `fmt.Errorf("failed to create document: %w", err)` wrap (line 43) | `mock.ExpectQuery(...).WillReturnError(errors.New("db down"))` |
| D2 | `TestDocumentRepo_GetDocument_GenericError` | `GetDocument` | Non-`sql.ErrNoRows` error wrap (line 65) | `mock.ExpectQuery(...).WillReturnError(errors.New("db down"))` |
| D3 | `TestDocumentRepo_UpdateDocument_NotFound` | `UpdateDocument` | `sql.ErrNoRows` → `ErrNotFound` mapping (line 84-85) | `mock.ExpectQuery(...).WillReturnError(sql.ErrNoRows)` — assert `errors.Is(err, repo.ErrNotFound)` |
| D4 | `TestDocumentRepo_UpdateDocument_GenericError` | `UpdateDocument` | Generic error wrap (line 87) | `mock.ExpectQuery(...).WillReturnError(errors.New("db down"))` |
| D5 | `TestDocumentRepo_DeleteDocument_GenericError` | `DeleteDocument` | `ExecContext` error wrap (line 97) | `mock.ExpectExec(...).WillReturnError(errors.New("db down"))` |
| D6 | `TestDocumentRepo_ListDocuments_QueryError` | `ListDocuments` | `QueryContext` error wrap (line 106) | `mock.ExpectQuery(...).WillReturnError(errors.New("db down"))` |
| D7 | `TestDocumentRepo_ListDocuments_ScanError` | `ListDocuments` | `rows.Scan` error wrap inside the loop (line 113-114) | `sqlmock.NewRows(...).AddRow("not-a-uuid", ...)` — type mismatch on first column |
| D8 | `TestDocumentRepo_ListDocuments_RowsErr` | `ListDocuments` | `rows.Err()` after the loop (line 119-120) | `sqlmock.NewRows(...).AddRow(valid).RowError(0, errors.New("rows err"))` |

### 5.2 `services/agent-board/internal/repo/project_repo_test.go` — 8 new

| # | Test function | Repo method | Branch covered | Mock shape |
|---|---|---|---|---|
| P1 | `TestProjectRepo_CreateProject_GenericError` | `CreateProject` | wrap at line 45 | `WillReturnError` |
| P2 | `TestProjectRepo_GetProject_GenericError` | `GetProject` | wrap at line 66 (non-ErrNoRows) | `WillReturnError` |
| P3 | `TestProjectRepo_UpdateProject_NotFound` | `UpdateProject` | `ErrNotFound` mapping at line 85 | `WillReturnError(sql.ErrNoRows)` |
| P4 | `TestProjectRepo_UpdateProject_GenericError` | `UpdateProject` | wrap at line 87 | `WillReturnError` |
| P5 | `TestProjectRepo_DeleteProject_GenericError` | `DeleteProject` | wrap at line 97 | `ExpectExec(...).WillReturnError` |
| P6 | `TestProjectRepo_ListProjects_QueryError` | `ListProjects` | wrap at line 106 | `ExpectQuery(...).WillReturnError` |
| P7 | `TestProjectRepo_ListProjects_ScanError` | `ListProjects` | wrap at line 113-114 | `AddRow` with column-type mismatch |
| P8 | `TestProjectRepo_ListProjects_RowsErr` | `ListProjects` | wrap at line 119-120 | `RowError(0, err)` |

### 5.3 Per-test assertions (uniform across all 16)

For each test, the assertions REQUIRED by US005 AC:

1. The returned error is non-nil.
2. For `_NotFound`: `errors.Is(err, repo.ErrNotFound)` is **true**.
3. For `_GenericError` / `_QueryError` / `_ScanError` / `_RowsErr`: `errors.Is(err, repo.ErrNotFound)` is **false** AND the error message contains the `fmt.Errorf("failed to <op>"...)` substring or `"error iterating"` for `_RowsErr` (substring match is acceptable; exact match is brittle).
4. `mock.ExpectationsWereMet()` returns `nil`.
5. Production return values (Documents/Projects pointer) are nil on the error path.

### 5.4 Coverage target

- Per-file: `project_repo.go` ≥ 95%, `document_repo.go` ≥ 95% — measured via `go test ./internal/repo -coverprofile=...` + `go tool cover -func=...`.
- Package: `internal/repo` ≥ 95%.
- If any specific line cannot be reached via sqlmock (e.g. a `panic`-on-impossible-condition branch — none currently exist), document it in the test report with a one-line justification. Threshold becomes "≥ 95% modulo enumerated unreachable lines."

### 5.5 No production-code edits

Asserted via `git diff` review at code-review time. If a test reveals an actual bug (e.g. a missing `rows.Err()` check), the dev raises `ARCHITECTURE_GAP_FOUND` and we open a new story. Do not silently fix in US005.

---

## 6. e2e stack-up shape (US008)

### 6.1 Directory layout after US008

```
/                                  (repo root)
├── docker-compose.yml             ← NEW
├── Makefile                       ← NEW
├── .dockerignore                  ← NEW
├── services/agent-board/
│   ├── Dockerfile                 ← NEW (multi-stage Go build)
│   └── ... existing
├── web/
│   ├── Dockerfile                 ← NEW (multi-stage Node build)
│   └── ... existing
├── tests/e2e/
│   ├── README.md                  ← NEW (runbook)
│   ├── data/
│   │   └── seeds/
│   │       └── REQ000_baseline.sql  ← NEW (example fixture)
│   ├── resources/                 ← currently missing; created when shared keywords land
│   ├── results/                   ← gitignored — Robot's output.xml / log.html / report.html
│   └── REQ00x_*/                  ← existing per-REQ Robot suites
└── ... existing
```

### 6.2 `docker-compose.yml` — services

Three services. Bind to `127.0.0.1` only.

```yaml
# docker-compose.yml (illustrative shape — tech-lead refines exact YAML)
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: agent_board
      POSTGRES_USER: agent_board
      POSTGRES_PASSWORD: agent_board
    ports:
      - "127.0.0.1:15432:5432"
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "agent_board", "-d", "agent_board"]
      interval: 2s
      timeout: 2s
      retries: 30
    volumes:
      - postgres-data:/var/lib/postgresql/data

  api-server:
    build:
      context: ./services/agent-board
      dockerfile: Dockerfile
    command: ["./api-server"]
    environment:
      DATABASE_URL: postgres://agent_board:agent_board@postgres:5432/agent_board?sslmode=disable
      PORT: "8080"
      FRONTEND_URL: http://localhost:3000
    ports:
      - "127.0.0.1:8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "-q", "-O-", "http://localhost:8080/api/v1/projects"]
      interval: 3s
      timeout: 3s
      retries: 10

  web:
    build:
      context: ./web
      dockerfile: Dockerfile
    environment:
      NEXT_PUBLIC_API_BASE_URL: http://localhost:8080
      PORT: "3000"
    ports:
      - "127.0.0.1:3000:3000"
    depends_on:
      api-server:
        condition: service_healthy

volumes:
  postgres-data:
```

Notes:
- **Postgres external port `15432`** (not 5432) so it does not collide with a host Postgres a dev may be running for other work.
- **`api-server` healthcheck hits `/api/v1/projects`** — endpoint is unauthenticated and always returns 200 (possibly empty list) once DB is reachable. Cheapest signal of "fully ready including DB connection."
- **`web` has no healthcheck** — Next.js's HTTP server is ready as soon as it binds; Robot's `Wait For Elements State` handles UI-render readiness.
- **`mcp-server` is NOT in the compose stack** by default. The e2e suites today use it via `mcp_keywords.resource` only for REQ001's MCP-specific scenarios; those run against the locally-started `mcp-server` from `startup.sh`. Adding `mcp-server` to compose is a follow-up if a REQ wants its e2e to exercise MCP through the live stack.
- **Bind mounts vs build context:** D-012 chooses **build context** (no bind mount of source). Pros: reproducibility, single command. Cons: any source change requires `make e2e-down && make e2e-up` to rebuild. Acceptable for an e2e-after-tests workflow; dev-loop people use `startup.sh` (preserved unchanged for that use case).

### 6.3 `Makefile` — targets

Exact target names as locked in po-ba D-006. The `e2e` target uses `set -e` + `trap` to guarantee teardown.

```makefile
# Makefile (repo root)
.PHONY: e2e-up e2e-down e2e-seed e2e-run e2e e2e-logs

DOCKER_COMPOSE ?= docker compose
SEEDS_DIR := tests/e2e/data/seeds
MIGRATIONS_DIR := services/agent-board/migrations
PG_CONN := postgres://agent_board:agent_board@localhost:15432/agent_board?sslmode=disable

e2e-up:                          ## start postgres + api-server + web (compose, healthcheck-gated)
	$(DOCKER_COMPOSE) up -d --wait

e2e-down:                        ## stop and remove containers + volumes
	$(DOCKER_COMPOSE) down -v

e2e-seed:                        ## apply migrations then seed fixtures (idempotent)
	@for f in $$(ls $(MIGRATIONS_DIR)/*.up.sql | sort); do \
	  echo "→ applying migration $$f"; \
	  psql "$(PG_CONN)" -v ON_ERROR_STOP=1 -f $$f; \
	done
	@for f in $$(ls $(SEEDS_DIR)/*.sql 2>/dev/null | sort); do \
	  echo "→ applying seed $$f"; \
	  psql "$(PG_CONN)" -v ON_ERROR_STOP=1 -f $$f; \
	done

e2e-run:                         ## run Robot suites (REQ=REQ001 US=US001 narrows)
	@mkdir -p tests/e2e/results
	@INCLUDE_FLAGS=""; \
	if [ -n "$(US)" ]; then INCLUDE_FLAGS="--include $(US)"; fi; \
	if [ -n "$(REQ)" ]; then \
	  robot --outputdir tests/e2e/results $$INCLUDE_FLAGS tests/e2e/$(REQ)_*/; \
	else \
	  robot --outputdir tests/e2e/results $$INCLUDE_FLAGS tests/e2e/REQ*/; \
	fi

e2e: e2e-up e2e-seed              ## up → seed → run → ALWAYS down (trap on failure)
	@set -e; \
	trap '$(MAKE) e2e-down' EXIT; \
	$(MAKE) e2e-run

e2e-logs:                        ## stream container logs
	$(DOCKER_COMPOSE) logs -f --tail=100
```

### 6.4 Migrations runner

Per D-010 the first iteration uses raw `psql -f` against each `*.up.sql` file in lex order. Pros: no new tool dependency, transparent. Cons: any future migration that uses `\` psql meta-commands or has multi-statement edge cases will need switching to `golang-migrate` — that is its own follow-up and not blocked by US008.

### 6.5 Seed fixtures contract

- Location: `tests/e2e/data/seeds/`.
- Naming: `REQxxx_<short_name>.sql` — alphabetical sort matches REQ order.
- Idempotency: each seed uses `INSERT ... ON CONFLICT DO NOTHING` (or `TRUNCATE` + insert at the top). Re-running `make e2e-seed` on an already-seeded DB MUST NOT error.
- Example to ship: `REQ000_baseline.sql` — creates one project + two documents with deterministic UUIDs that Robot suites can reference. Real per-REQ seeds get added by future REQs.

### 6.6 Robot invocation pattern

Existing `.robot` files under `tests/e2e/REQ00x_*/` are unchanged. They tag tests with the US ID (e.g. `[Tags]    US001    smoke`). `make e2e-run REQ=REQ001 US=US001` translates to `robot --include US001 tests/e2e/REQ001_*/`. `make e2e-run REQ=REQ001` runs all stories in REQ001.

Suites that today reach `http://localhost:3000` and `http://localhost:8080` continue to work because the compose stack publishes both on `127.0.0.1` at those exact ports.

### 6.7 What this section does NOT cover

- CI integration (GitHub Actions / etc.) — follow-up REQ.
- Containerised Robot Framework — Robot runs on the host (uses the local Python install + Browser library). The `Makefile` assumes `robot` is on PATH; the runbook lists `pip install robotframework robotframework-browser robotframework-requests` as prerequisites.
- Test-data seeding via the MCP / REST API rather than SQL — explicitly rejected for v1 per D-010.

---

## 7. Worktree origin contract (US009)

### 7.1 Path choice: **Path (b) — agent-definition documentation fallback**

**Verdict: Path (a) is NOT reachable from this repo.** The worktree-creation harness is part of the Claude Code runtime / orchestrator layer; this repo cannot edit it. Therefore US009 ships path (b) only.

Rationale:

- We searched the repo for `git worktree add` invocations and found none — they happen outside the working tree, in the harness layer.
- `.claude/worktrees/` is gitignored by this repo (`.gitignore:54`); the directory's contents are created by the harness, not by repo tooling.
- An "edit the harness" PR is not landable from inside this repo's worktree.

If a future change makes path (a) reachable, that becomes its own one-line REQ.

### 7.2 Canonical-path edit policy block (path b deliverable)

The following block is appended verbatim to **all six** agent-definition files listed in §2 US009 row. Identical wording across files (a copy-paste-able block — that uniformity is intentional and is enforceable by a future lint check).

```markdown
## Canonical-path edit policy (worktree workaround)

Sub-agent worktrees may currently branch off `origin/main`, which can be stale
by N commits during an active REQ (the harness that creates worktrees lives
outside this repo and has not yet been updated to branch off local `main` HEAD).
To avoid add/add merge conflicts on spec / docs / agent-definition files that
already exist on local `main`, ALWAYS edit the following file classes at their
canonical path under `/Users/a667282/workspace/agents-board/`, never at the
worktree-local path:

  - `docs/requirements/REQ###_*/` (all README, architecture, story, task,
    spec, test-report files)
  - `tests/e2e/REQ###_*/` (Robot suites and shared resources)
  - `.claude/agents/*.md` (agent definitions including this file)
  - `CLAUDE.md` / `AGENTS.md` (the top-level orchestrator instructions)

Do NOT `git add` the worktree-local copies of these files. The orchestrator
runs `git status` after each phase and surfaces any worktree-local hits on the
above paths as a routing error.

This policy is REQ005 / US009 fallback documentation. When the harness is fixed
to branch off local `main`, this block becomes obsolete and can be removed in
the same change.
```

### 7.3 README decision-log entry

The block to append to `docs/requirements/REQ005_quality_hardening_retrospective/README.md` under the existing "Decision log" section:

```markdown
### Decision: US009 path = (b)

Path (a) (harness fix to branch off local `main`) was determined NOT reachable
from this repo by the system-architect on 2026-06-01: the worktree-creation
harness lives in the Claude Code runtime layer outside the working tree, and
no `git worktree add` invocation exists in the repo. US009 therefore ships
path (b) only: the canonical-path edit policy block is appended verbatim to
all six agent definition files (`.claude/agents/{po-ba,system-architect,
tech-lead,tester,be-dev,fe-dev}.md`).
```

### 7.4 Out of scope for US009

- Cleaning up old `.claude/worktrees/` directories.
- A "worktree freshness" linter or pre-commit hook (path (a) territory).
- Migrating away from worktrees entirely.

---

## 8. API-contract impact

**N/A.** REQ005 does not add, remove, or change any endpoint, request shape, response shape, status code, or error code. The REQ001–REQ004 API surface is preserved:

- `GET /api/v1/projects`
- `GET /api/v1/projects/{id}`
- `GET /api/v1/projects/{id}/documents`
- `GET /api/v1/documents/{id}`
- MCP `/sse`, `/message`

A code-review check at tech-lead's gate (US004 / US006 reviews) confirms that handler / response-encoding code paths are NOT touched.

---

## 9. Decisions (ADR-lite)

### Restated decisions from po-ba's intake log

| ID | Story | Decision | Verdict | Notes |
|---|---|---|---|---|
| **D-001** | US002 | `--forceExit` immediately; real MSW leak hunt deferred to a separate story | **ACCEPT** | See §10 Q1 — recommend defer to REQ006, not a US010 in REQ005. |
| **D-002** | US003 | Soft-warn `run_check_warn` pattern; `gosec` ruleset covered by `golangci-lint`; README documents substitution | **ACCEPT** | See §9 D-011 for helper-shape guidance. |
| **D-003** | US004 | Only the two production `context.Background()` sites are touched; tests legitimately use it | **ACCEPT** | |
| **D-004** | US005 | 14 backfill tests (16 functions when Update is split into NotFound + GenericError) per audit §4.4 shopping list | **ACCEPT** | See §5 matrix. |
| **D-005** | US006 | All three hooks use AbortController + signal-thread pattern; `fetchProject` gains `signal?` param; no new shared `useFetch<T>` extraction | **ACCEPT** | See §4 contract. |
| **D-006** | US008 | `docker-compose.yml` at repo root; Makefile targets `e2e-up`/`e2e-down`/`e2e-seed`/`e2e-run`/`e2e`/`e2e-logs`; SQL fixtures under `tests/e2e/data/seeds/`; runbook in `tests/e2e/README.md` | **ACCEPT** | See §6. |
| **D-007** | US009 | Two acceptance paths; (a) harness preferred, (b) agent-definition fallback | **ACCEPT then path (b) chosen** | Path (a) not reachable from repo — see §7.1. |

### New architectural decisions

#### D-008 — No `internal/lifecycle/` helper package for US004
- **Context:** Two `main.go` files need the same signal-cancellable lifecycle context pattern (~9 lines of glue each).
- **Decision:** Inline the pattern in both `cmd/api-server/main.go` and `cmd/mcp-server/main.go`. **Do NOT create a `services/agent-board/internal/lifecycle/` package.**
- **Alternatives rejected:** `internal/lifecycle/context.go` with a single `New()` constructor. Rejected because (a) two call sites do not justify a new package + its tests; (b) introducing a package widens the API surface that future REQs have to maintain; (c) the pattern is plain stdlib and reads more clearly inline.
- **Consequences:** If a third or fourth `main.go` shows up needing the same wiring, that REQ extracts the helper. Until then, copy-paste is cheaper.

#### D-009 — Graceful HTTP shutdown is out of scope
- **Context:** US004 establishes a signal-cancellable lifecycle context. The natural follow-on is to wire `e.Shutdown(ctx)` so SIGTERM during normal serving drains in-flight requests.
- **Decision:** **Out of scope for REQ005.** US004 only fixes boot-time DB ping. `e.Start(":" + port)` remains a blocking call that exits on a fatal serve error or process kill.
- **Alternatives rejected:** Bundling shutdown into US004. Rejected because graceful shutdown has its own design surface (Echo's shutdown semantics, drain-timeout policy, signal-during-Shutdown behaviour) and would inflate US004 beyond a single-INVEST story.
- **Consequences:** SIGTERM mid-serve is still a hard kill of in-flight requests. Operators have not asked for graceful shutdown yet; raise a follow-up REQ when they do.

#### D-010 — `psql -f` for migration / seed application in US008
- **Context:** US008 needs to apply `services/agent-board/migrations/*.up.sql` and `tests/e2e/data/seeds/*.sql` from a Makefile.
- **Decision:** Iterate over the SQL files in lex order and `psql -v ON_ERROR_STOP=1 -f` each one.
- **Alternatives rejected:** (a) `golang-migrate/migrate` CLI — adds a new tool dep on every dev machine; (b) a Go-based runner with `embed.FS` — yet-another binary to build and maintain; (c) docker-entrypoint-initdb scripts — only run on first container create, not idempotent across compose lifecycles.
- **Consequences:** Limited to plain SQL (no psql meta-commands). Migrations grow more complex → switch to option (a) in a follow-up REQ.

#### D-011 — US003 helper shape: inline `command -v` check (no new helper function)
- **Context:** US003 must replace two `require_tool` calls with a soft-warn behaviour.
- **Decision:** Inline the `command -v` check in `gate_be()` before each affected `run_check`. Print a WARN-prefixed line on miss; skip the `run_check`. **Do NOT add a `run_check_warn_if_missing` helper.**
- **Alternatives rejected:** New helper `run_check_warn_if_missing <tool> <name> <cmd...>`. Rejected because the call sites are exactly two and inline shell `if ! command -v X; then ...; else run_check ...; fi` is more obvious than indirection through a third helper.
- **Consequences:** If a third soft-warn-on-missing tool appears, extract the helper then.

#### D-012 — US008 web service runs as a containerised production build
- **Context:** US008 notes give two options — containerised vs host `next dev`.
- **Decision:** Containerised. `web/Dockerfile` does `npm ci && npm run build && npm start` in a multi-stage build. Compose runs the prod build.
- **Alternatives rejected:** Host `next dev` — splits the orchestrator's "one command" into "one command plus a sidecar terminal," and stale build state between compose-up and the host dev server is a real foot-gun. The `startup.sh` already covers the host-dev-loop workflow and remains unchanged.
- **Consequences:** Source changes require `make e2e-down && make e2e-up` to rebuild; that is acceptable for an "after the unit tests pass, run e2e once" workflow. FE devs in tight loop use `startup.sh` or `cd web && npm run dev`.

#### D-013 — DB-ping timeout literal `5 * time.Second`; no env-var for now
- **Context:** US004 leaves the timeout value to the architect; suggests an optional `DB_PING_TIMEOUT_SECONDS`.
- **Decision:** Hard-code `5 * time.Second` with a TODO comment. No env var.
- **Alternatives rejected:** `DB_PING_TIMEOUT_SECONDS` env-var. Rejected for now because no operator has asked for it; YAGNI; adding it later is a 4-line non-architecture change.
- **Consequences:** Ops cannot tune the ping timeout without a code change. Acceptable trade-off until someone needs it.

#### D-014 — US010 (React-Doctor regressions) folded into REQ005 — ACCEPTED
- **Context:** po-ba's README D-008 (added 2026-06-02) folds the React-Doctor baseline regression fixes — 3 state/effect clusters + 1 security — into REQ005 as US010. tech-lead's one-off scan recorded 92/100 with 4 errors and 19 warnings against current `main`; ~70% of the state/effect findings live in REQ004-shipped files. The alternatives were (a) defer to a new REQ006, (b) include in REQ005 now while REQ005 is pre-Phase-2 and scope is cheap.
- **Decision:** **Accept.** US010 is added to REQ005 and absorbed into this architecture via §11 (full contract), §2 (file touch map), §9 (this decision), §12 (react-doctor skill enforcement note), §14 (R7, R8, OQ-5, OQ-6).
- **Alternatives rejected:** Deferring to REQ006. Rejected because (a) REQ005 is pre-Phase-2 — adding a story now is free, while a separate REQ would re-litigate scope REQ005 already owns (e2e harness, FE hook harmonisation, gate fixes — US010 lives next to all of these); (b) fixing react-doctor findings before they accumulate compounding reviewer cost beats a separate cycle later; (c) US006 already touches `useDocument` — landing US010 in the same REQ keeps that file's churn coherent and well-ordered (see §11.4).
- **Consequences:** REQ005 grows from 9 stories to 10. US010 is FE-only and depends on US006 (one-directional). The four React-Doctor error rules go to zero; the 15 lower-severity baseline findings stay recorded for tracking. The architect picks ref-attach over DOMPurify for `MermaidDiagram` (rationale in §11.1.1) — confirm at re-approval via OQ-5.

---

## 10. Open-question resolutions

### Q1 — MSW leak root-cause story (add as US010 or defer?)

**Recommendation: DEFER to a follow-up REQ (suggest REQ006), do NOT add as US010 to REQ005.**

Reasoning:

- The audit already flags `--detectOpenHandles` + targeted analysis (react-markdown lazy-load? mermaid timer? jsdom global?) as the real fix. Each of those candidate causes has its own investigation arc.
- Bundling that hunt into REQ005 would re-open a story type ("investigation with uncertain time-box") that REQ005 is explicitly avoiding. REQ005's nine stories are all single-INVEST tech-debt items; an investigation does not fit that profile.
- US002's `--forceExit` already closes the operational pain. The leak is paid down as developer-time tax (Jest's `worker process has failed to exit gracefully` warning), not as gate flakiness. Acceptable to live with for one more REQ cycle.
- When the leak is hunted in REQ006, the resulting fix may want to remove `--forceExit` from US002. That coupling is clean: a single follow-up REQ that fixes the leak + removes the workaround.

### Q2 — `docker-compose.yml` location (repo root vs `tests/e2e/`)

**Recommendation: REPO ROOT.** (Locked as §6.1 layout; matches po-ba's preference in D-006.)

Reasoning:

- A repo-root `docker-compose.yml` is discoverable by anyone who clones the repo and runs `docker compose up`, with no path argument needed.
- Useful beyond e2e: a dev who wants a local Postgres without using `startup.sh` can `docker compose up postgres`. Even FE devs may want it.
- Keeps the Makefile targets simple — no `-f tests/e2e/docker-compose.yml` flag clutter.
- The e2e-specific bits (seed fixtures, Robot runbook) DO live under `tests/e2e/`; only the compose file moves up.

### Q3 — US009 harness reachability

**Resolved: NOT reachable from this repo.** US009 ships path (b) only — agent-definition documentation fallback. See §7.1 and D-007 verdict.

If the harness layer is later determined to be reachable (e.g. a `~/.claude/` config or a Claude Code update path becomes editable), the canonical-path policy block is removed and replaced with a one-line note in the README pointing at the harness commit. Until then, path (b) is the contract.

---

## 11. US010 contract — React-Doctor regression fixes

This section is frozen at re-approval. fe-dev implements against it without re-interpreting; tester binds component assertions to it. The four fixes are independent of each other at the file level and can be implemented in any order within US010, but US010 as a whole is gated by US006 (see §11.4).

### 11.1 MermaidDiagram fix path — REF-ATTACH (default; recommended)

#### 11.1.1 Why not DOMPurify

The naive read of the rule is "DOMPurify sanitises the SVG, problem solved." It does not.

- `react-doctor/no-danger` matches the JSX prop name `dangerouslySetInnerHTML` *syntactically*. It does not introspect what string is being injected. Whether or not the string passed through DOMPurify, the rule will fire because the prop is still present in the JSX.
- Clearing the rule via DOMPurify therefore requires an inline lint suppression (`// eslint-disable-next-line react-doctor/no-danger`). The AC for US010 (Scenario: react-doctor score recovers, and the seven-rule-id exclusion list) does not permit per-line suppressions — it requires the rule to genuinely not fire.
- DOMPurify also adds a runtime + dependency cost (`isomorphic-dompurify` is ~30 kB gz on top of the existing 300 kB gz mermaid chunk) for no real defence-in-depth gain here: `mermaid.initialize({ securityLevel: 'strict' })` already prevents script tags / event handlers in the rendered SVG, and mermaid's input in this app is never user-controlled (it comes from project markdown authored by the team).

If the human prefers DOMPurify + suppression (OQ-5), the rule will be re-evaluated; today's default is ref-attach.

#### 11.1.2 Ref-attach contract — exact `useEffect` shape

The behaviour-preserving change replaces the success-branch JSX (lines 130–142 in current `MermaidDiagram.tsx`) with a `ref`-attached wrapper and a second `useEffect` that performs the SVG attach. The lazy-load `useEffect` at lines 78–115 is **unchanged**.

```tsx
// Above the component, no new imports beyond what's needed:
import React, { useEffect, useId, useRef, useState } from 'react';

// Inside the component, after the existing useState<RenderState>(...):
const containerRef = useRef<HTMLDivElement | null>(null);

// NEW effect that runs on transition into 'success'. Does NOT replace the existing
// lazy-load effect — it runs in addition to it.
useEffect(() => {
  if (renderState.status !== 'success') return;
  const host = containerRef.current;
  if (!host) return;

  // Parse the mermaid-produced SVG string into a DOM node. mermaid v11 emits a
  // standalone <svg> root, so DOMParser with 'image/svg+xml' MIME type is correct.
  const parsed = new DOMParser().parseFromString(renderState.svg, 'image/svg+xml');
  const svgNode = parsed.documentElement;

  // Clear any previous child (covers the source-change re-render path) and attach.
  while (host.firstChild) host.removeChild(host.firstChild);
  host.appendChild(svgNode);

  return () => {
    if (host.firstChild === svgNode) host.removeChild(svgNode);
  };
}, [renderState]);

// And the success-branch JSX becomes:
if (renderState.status === 'success') {
  const { ariaLabel } = renderState;
  return (
    <div
      ref={containerRef}
      role="img"
      aria-label={ariaLabel || undefined}
      style={{ maxWidth: '100%', overflowX: 'auto' }}
    />
  );
}
```

Notes:
- The `DOMParser` call is browser-only. Safe here: this component is CSR-only (the parent lazy-loads it via `dynamic({ ssr: false })`) and the existing lazy-load `useEffect` already gates everything on browser-only execution.
- React 18 strict mode double-invokes effects in dev; the cleanup function removes the appended node before the second run re-appends, so no duplicate SVG renders.
- `parsed.documentElement` is the `<svg>` element (parser uses `'image/svg+xml'` MIME). If mermaid ever changes to wrapping in `<g>` only, the parse still succeeds — `documentElement` becomes the wrapper, and the visual output is identical.
- Error handling unchanged: parsing failure inside the new effect would throw, but mermaid v11 emits well-formed SVG strings. Tester adds a defensive FCT-* test using a malformed `svg` string and asserts the component does NOT throw (we wrap in try/catch if so).

#### 11.1.3 Visual parity guarantee

The DOM after the new code is `<div role="img" aria-label="..."><svg>...</svg></div>` — same wrapper, same svg subtree as today, same `role="img"`, same `aria-label`. Existing snapshot tests (if any) pass unchanged; existing `container.querySelector('svg')` assertions pass unchanged.

### 11.2 `useDocument` reducer contract

#### 11.2.1 State shape (frozen)

```ts
type State = {
  data: Document | null;
  isLoading: boolean;
  error: ApiError | Error | null;
};
```

NOT in state (computed inline at render or kept as refs):
- `latestIdRef` — stays as `useRef`, NOT in reducer state. It is request-correlation metadata, not user-observable state.
- `controllerRef` — stays as `useRef`, NOT in reducer state. Same reason.
- `fetchCount` — internal `useState` that drives the effect dep; can stay a separate `useState` next to the `useReducer`, OR fold into the reducer as a `RETRY` action that increments a counter in state. **Default: keep `fetchCount` as a separate `useState`** to keep the reducer's surface minimal. See §11.2.4.
- `hasMore` — NOT present in this hook. The user prompt mentioned `hasMore` as part of a generic reducer shape; `useDocument` is single-document fetch, not paginated, so it is omitted.

#### 11.2.2 Action types (frozen)

```ts
type Action =
  | { type: 'FETCH_STARTED' }                                    // dispatched when a new fetch begins
  | { type: 'FETCH_SUCCEEDED'; document: Document }              // dispatched on .then for the latest id
  | { type: 'FETCH_FAILED'; error: ApiError | Error }            // dispatched on .catch for the latest id (non-abort)
  | { type: 'ABORTED' };                                         // dispatched on .catch when controller.signal.aborted
```

#### 11.2.3 Reducer signature (frozen)

```ts
const initialState: State = { data: null, isLoading: false, error: null };

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'FETCH_STARTED':    return { data: null, isLoading: true,  error: null };
    case 'FETCH_SUCCEEDED':  return { data: action.document, isLoading: false, error: null };
    case 'FETCH_FAILED':     return { data: null, isLoading: false, error: action.error };
    case 'ABORTED':          return state;                       // no-op; aborts are control flow, not failures
    default:                 return state;
  }
}
```

The `ABORTED` action is dispatched but is a no-op on state. We keep it explicit (rather than just `return` early in the catch) so the reducer's action set is exhaustive and so tests can assert "on abort, no state mutation occurs."

#### 11.2.4 Public hook return shape (frozen — DO NOT CHANGE)

```ts
export interface UseDocumentResult {
  data: Document | null;
  isLoading: boolean;
  error: ApiError | Error | null;
  refetch: () => void;
}
```

**This shape MUST stay byte-identical to today's exported `UseDocumentResult`.** Every existing Jest test that destructures `{ data, isLoading, error, refetch }` continues to pass. Every consumer (`DocumentsTab.tsx` line 81) continues to compile. Internal implementation moves from `useState×3` to `useReducer`; the consumer-visible contract does not move.

#### 11.2.5 Effect body shape (target)

```ts
useEffect(() => {
  if (documentId === undefined) return;

  controllerRef.current?.abort();
  const controller = new AbortController();
  controllerRef.current = controller;
  latestIdRef.current = documentId;

  dispatch({ type: 'FETCH_STARTED' });

  fetchDocument(documentId, controller.signal)
    .then((doc) => {
      if (latestIdRef.current === documentId) {
        dispatch({ type: 'FETCH_SUCCEEDED', document: doc });
      }
    })
    .catch((err: unknown) => {
      if (controller.signal.aborted) {
        dispatch({ type: 'ABORTED' });
        return;
      }
      if (latestIdRef.current !== documentId) return;        // stale, late resolution
      const error: ApiError | Error =
        err instanceof ApiError ? err :
        err instanceof Error ? err :
        new Error('Failed to load document');
      dispatch({ type: 'FETCH_FAILED', error });
    });

  return () => { controller.abort(); };
}, [documentId, fetchCount]);
```

Compare against today (lines 50–94 in `useDocument.ts`): 7 setState calls collapse to 4 dispatch calls; the cascading-setState cluster is gone; the rule fires on `react-doctor` go to zero for this file.

#### 11.2.6 Why this also clears the named rules

- `no-cascading-set-state` — fired today because the effect contained `setIsLoading(true); setError(null); setData(null);` in sequence. Reducer pattern collapses these into one `dispatch({ type: 'FETCH_STARTED' })` — single state transition.
- `no-adjust-state-on-prop-change` × 3 — fired today because `setData(null)`, `setIsLoading(true)`, `setError(null)` were invoked in response to the `documentId` prop changing. In the reducer pattern, the prop change triggers ONE dispatch, and the reducer derives the next state from the action — not from the prop directly. The rule no longer pattern-matches.
- `rendering-usetransition-loading` — fires today on line 39 (`useState<boolean>(false)` for `isLoading`). With `useReducer`, `isLoading` is a reducer field, not a standalone `useState`. The rule scans for `useState<boolean>` named-as-loading patterns; reducer-derived booleans do not match.
- `exhaustive-deps` — the effect's deps `[documentId, fetchCount]` are unchanged. No new closures, no new refs entering the effect body. The rule passes naturally.

### 11.3 `DocumentsTab` redirect contract

#### 11.3.1 What the current `useEffect` actually does

Lines 63–78 of `DocumentsTab.tsx` fire `router.replace` when **(a)** documents have loaded and **(b)** `docParam` is undefined. This is auto-selection on first load — it ensures the URL ends up reflecting the selected document.

The `handleSelectDoc` function (line 88) is the user-click counterpart: when the user clicks a sidebar item, `handleSelectDoc(id)` runs `router.replace` to update the URL. **The two paths perform identical URL writes; only the trigger differs.**

#### 11.3.2 The fix — render-time computed `selectedDocId`; no auto-select effect

The cleanest behaviour-preserving change drops the effect entirely:

```tsx
// Compute selectedDocId at render time. Fallback to first document on initial load
// (when docParam is absent) replicates today's auto-select observable outcome.
const selectedDocId = isBogusDeepLink
  ? undefined
  : (docParam ?? documents?.[0]?.id);
```

This is the **single line change** to line 59. Lines 63–78 (the auto-select effect) are deleted.

Observable behaviour at render time:
- Initial load, `?doc=` absent, list non-empty: `selectedDocId = documents[0].id`. `DocumentSidebar` shows item 0 highlighted; `DocumentPreviewer` fetches document 0. **Same as today.**
- User clicks sidebar item N: `handleSelectDoc(documents[N].id)` runs, `router.replace` writes `?doc=...`, `docParam` becomes that id on next render, `selectedDocId` becomes `docParam`. **Same as today.**
- Bogus deep-link (`?doc=` set, not in list): `isBogusDeepLink === true`, `selectedDocId = undefined`, previewer renders `isNotFound`. **Same as today.**
- Empty list: short-circuited by the `documents.length === 0` early return at line 165, unchanged.
- List undefined (loading): `documents?.[0]?.id` is `undefined`, but the `listLoading` early return at line 112 fires first anyway.

#### 11.3.3 What about the URL not getting updated on initial load?

Today's effect writes `?doc=<first>` into the URL on initial load. The new render-time selection does NOT write to the URL — the user sees the first document in the previewer but the URL stays bare (`/projects/123?tab=documents`, no `&doc=...`).

This is acceptable for two reasons:
- (a) The URL-as-source-of-truth comment at the top of `DocumentsTab.tsx` (lines 20–27) describes URL-driven *selection*, not URL-driven *defaulting*. Defaulting to the first item is render policy; only user actions write to the URL.
- (b) The user can still deep-link by sharing the URL after clicking an item — the click handler writes the URL on every interaction.

**If po-ba pushes back at sign-off and wants the URL written on initial load,** that becomes either (a) a one-line addition to `handleSelectDoc` invoked from a `useEffect` (re-introducing the rule violation — rejected), or (b) a `useLayoutEffect` with `router.replace` gated to the first mount only (also re-introduces `nextjs-no-client-side-redirect`), or (c) accept that bare URL is the new default. Default verdict: (c). See OQ-6 if the human prefers otherwise.

#### 11.3.4 Why this clears the named rules

- `no-event-handler` (line 40) — fires on `refetch: refetchList` being destructured into a variable that is later passed as an event handler without being wrapped. The fix is unrelated to the auto-select effect; tech-lead's review should confirm `onClick={refetchList}` at line 151 still satisfies the rule (it does — it's a stable `useCallback` from the hook). If the rule still fires, fe-dev wraps the call site as `onClick={() => refetchList()}` — trivial, no architectural impact. Tester adds an FCT-* assertion that the retry click triggers a refetch.
- `nextjs-no-client-side-redirect` (line 69) — fires on `router.replace` inside a `useEffect`. The effect is deleted; the rule cannot fire. The remaining `router.replace` in `handleSelectDoc` (a user-click handler, not an effect) does not match the rule's pattern.
- `exhaustive-deps` (line 78) — fires today because the existing effect has `// eslint-disable-line react-hooks/exhaustive-deps`. With the effect deleted, the rule cannot fire.

### 11.4 US006 ⇄ US010 ordering

**Decision: US006 lands first; US010 lists `Depends-on: US006` in its task header.**

Rationale:

- US006 harmonises three FE hooks (`useProject`, `useProjectDocuments`, `useDocument`) on AbortController + signal-thread. `useDocument` is the *reference implementation* in US006 — its current state-machine shape is what `useProject` and `useProjectDocuments` are rewritten to match.
- US010 refactors `useDocument` from `useState×3` to `useReducer`. After US010, `useDocument` is **no longer structurally similar** to the other two hooks.
- If US010 lands first: US006 still has to rewrite `useProject` and `useProjectDocuments` to match a `useState×3 + AbortController` template (the current `useDocument`), but the live reference (`useDocument`) has moved on to `useReducer` — confusing for the dev, doubles the cognitive cost.
- If US006 lands first: US006's contract (§4 of this architecture) is satisfied by the existing `useDocument`. US010 then mechanically swaps `useState×3` for `useReducer` inside `useDocument` while preserving the AbortController flow (its action set already names `ABORTED`). The diff is small and isolated to one file.

**Concrete task-graph consequences for tech-lead (Phase 2):**

- US006 tasks have NO dependency on US010 tasks.
- US010 tasks declare `Depends-on: US006_fe_use_document_abort_controller` (or whichever specific US006 FE task touches `useDocument`).
- This dependency is enforced at Phase 3a queue-build time — US010's fe-dev does not pick up the task until US006's matching task is `completed`.
- If US010 is delayed (e.g. blocked at review), US006 is NOT held up — it can complete independently. The ordering is one-directional.

### 11.5 What US010 does NOT change

- The lazy-load contract for mermaid (`dynamic({ ssr: false })` parent wrapping, in-effect `import('mermaid')`, module-level cache, `securityLevel: 'strict'`).
- The error UX of `MermaidDiagram` — the `<div role="alert">Could not render diagram</div>` + `<pre><code>` fallback is preserved verbatim.
- The public TypeScript shape of `UseDocumentResult` — every existing consumer keeps compiling, every existing Jest assertion keeps passing.
- The AbortController + `latestIdRef` race-safety in `useDocument` — preserved through the reducer refactor; the `ABORTED` action exists specifically to make the abort path a first-class part of the reducer's action set without state mutation.
- The URL-as-source-of-truth model in `DocumentsTab` for user-driven selection — `handleSelectDoc` continues to write `?doc=` on every click.
- The `isBogusDeepLink` semantics — preserved unchanged.
- The mermaid library version, config, or load mode.
- Any of the 15 lower-severity React-Doctor baseline findings (out of scope per US010 anti-scope).
- `web/package.json` dependencies (default path: ref-attach adds nothing).

---

## 12. Skill / hook usage (cross-cutting reminder)

The existing skill enforcements in `.claude/agents/*.md` continue to apply for the BE + FE code-touching stories:

- **TDG skill** (Test-Driven Generation, vendored at `.claude/skills/`): enforced on **be-dev** and **fe-dev** when implementing US004, US005 (test-only, but the methodology of "write the failing assertion first, confirm the right failure mode, then run") still applies, US006, and US010.
- **react-doctor skill**: enforced on **fe-dev** for US006 AND **US010** — for US006 it covers React effect-cleanup correctness, hook-deps soundness, and AbortController lifetime within `useEffect`; for US010 it is the *primary review gate* (`npx react-doctor scan web/` + `react-doctor --diff` against the recorded baseline must show only removals on the four targeted error rules).
- **Gate script (run-gate.sh)**: tech-lead runs it on every review. US001 / US002 / US003 IMPROVE this gate; US004 / US005 / US006 / US007 / US010 are reviewed THROUGH it.

No new skills are introduced by REQ005.

---

## 13. Cross-cutting

### 13.1 Env vars touched

- **Existing, no change:** `DATABASE_URL`, `DB_URL`, `PORT`, `FRONTEND_URL`, `NEXT_PUBLIC_API_BASE_URL`.
- **New, optional, NOT introduced now (D-013):** `DB_PING_TIMEOUT_SECONDS`. Left as a TODO comment in code; ops can add it later without an architecture revision.
- **US008 compose-local:** `POSTGRES_USER=agent_board`, `POSTGRES_PASSWORD=agent_board`, `POSTGRES_DB=agent_board`. These are local-only and never reach prod.
- **US010:** no env vars touched.

### 13.2 Logging

- **No new logging keys** introduced. The `log.Printf("api-server exited with error: %v", err)` and `log.Printf("Starting api-server on port %s", safePort)` lines in `cmd/api-server/main.go` remain unchanged.
- US004 may add one line on the ping-timeout / signal-cancel branch via the wrapped error message in the `return fmt.Errorf("db ping failed: %w", err)`. The outer `log.Printf("api-server exited with error: %v", err)` will print the wrapped chain.

### 13.3 Metrics / observability

- **N/A.** No metrics framework today; REQ005 does not add one.

### 13.4 Error model

- BE error wrapping convention (`fmt.Errorf("failed to ...: %w", err)`) is preserved and extended into US004's new ping-failure wrap.
- FE error model (`ApiError` with `code` + `message`) is preserved unchanged; AbortError handling is layered in front (see §4.5) AND remains intact through US010's reducer refactor (the `FETCH_FAILED` action carries `ApiError | Error`; the `ABORTED` action is the explicit non-error swallow path — see §11.2.2).

### 13.5 CORS

- Unchanged. `services/agent-board/cmd/api-server/main.go` continues to honour `FRONTEND_URL` for `AllowOrigins`. Compose's `web` service sets `FRONTEND_URL=http://localhost:3000`.

---

## 14. Risks & open questions

### 14.1 Risks (with mitigations)

- **R1: `docker-compose.yml` build of `api-server` requires a `Dockerfile` that does not exist yet.** Mitigation: tech-lead splits US008 into a BE-Dockerfile task + a compose+Makefile task; the Dockerfile is a 20-line multi-stage build, low risk.
- **R2: US007 (`@testing-library/dom` move) may cause a non-trivial lockfile churn.** Mitigation: AC requires the diff to be inspected before commit; if churn is large, dev raises `ARCHITECTURE_GAP_FOUND` (e.g. if mermaid actually depends on it at the top level via a peer chain we are unaware of).
- **R3: US005's coverage assertion (`≥ 95% per file`) may be optimistic if any line is genuinely unreachable via sqlmock.** Mitigation: US005 AC already allows the threshold to become "≥ 95% modulo enumerated unreachable lines"; tester surfaces any unreachable lines in the test report.
- **R4: US006's abort-semantics tests may flake on slow CI due to MSW timing.** Mitigation: tester uses `delay(100)`-style explicit delays in MSW handlers + `await waitFor(...)` rather than fixed sleeps. The existing `useDocument` tests are the working precedent.
- **R5: Docker not installed on a dev's machine breaks `make e2e-*`.** Mitigation: `tests/e2e/README.md` runbook lists Docker as a prerequisite; the Makefile prints a friendly error (`docker compose: command not found`) and exits non-zero; nothing else regresses.
- **R6: US009 path (b) accumulates lint debt** — six identical blocks across six files are now in lockstep. If one drifts, the workaround stops working consistently. Mitigation: tech-lead enforces identical-wording check during review (a `diff <(grep -A 12 'Canonical-path edit policy' file1) ...` one-liner suffices).
- **R7: US010 ref-attach path may double-mount under React 18 strict mode.** The new `useEffect` in `MermaidDiagram.tsx` (§11.1.2) appends a parsed `<svg>` node to a `ref`'d host. Strict mode invokes effects twice in dev; the cleanup function removes the appended node before the second run re-appends. Mitigation: cleanup is spelled out in §11.1.2; tester adds an FCT-* assertion that asserts exactly one `<svg>` child after mount under `<React.StrictMode>`.
- **R8: US010's `DocumentsTab` change drops the initial-load URL write.** Tests that asserted "`router.replace` is called when list loads with no `?doc=`" must be updated to "`router.replace` is called only when the user clicks a sidebar item." Mitigation: §11.3 enumerates this; tester's `fe_unit_tests.md` revision for US010 covers it. If po-ba pushes back on the bare-URL behaviour at sign-off, raise OQ-6.

### 14.2 Open questions surfaced for the human (please confirm at approval)

- **OQ-1 (US008 web build mode).** D-012 chooses containerised `next start` (production build). If the human prefers `next dev` inside the container for faster iteration when running `make e2e`, raise it now — flip D-012, change the Dockerfile's `CMD`, no other architecture impact.
- **OQ-2 (US004 env var).** D-013 hard-codes the 5-s DB-ping timeout. If ops policy already requires every timeout to be env-driven, raise it now — add `DB_PING_TIMEOUT_SECONDS` with a 5-s default. Trivial.
- **OQ-3 (US009 path).** §7.1 confirms path (b). If the human knows that the harness IS reachable somehow (an internal Claude Code config the architect couldn't see), please name the location at approval time and we add a path (a) task.
- **OQ-4 (MSW leak hunt).** §10 Q1 recommends DEFER to REQ006. If the human prefers an in-REQ stub now (was previously phrased as "US010" — note that US010 is now React-Doctor regressions per D-008, so the MSW hunt stub would be US011 if folded in), please say so.
- **OQ-5 (US010 MermaidDiagram fix path — NEW).** §11.1 recommends **ref-attach** as the default and rejects DOMPurify + inline lint suppression (won't clear `react-doctor/no-danger`, needs eslint disable comment, adds dep + bundle weight). If the human prefers DOMPurify + suppression anyway (e.g. team policy is "always sanitise mermaid output as defence-in-depth even though input is trusted"), confirm at approval; the architect will revise §11.1 + the §2 US010 row and add `isomorphic-dompurify` to the dep matrix. Default verdict if no objection: **ref-attach**.
- **OQ-6 (US010 DocumentsTab initial-load URL behaviour — NEW).** §11.3.3 documents that the URL is no longer auto-written on initial load (it stays bare `/projects/:id?tab=documents` until the user clicks a sidebar item). Today's behaviour writes `?doc=<first>` on initial load via the `useEffect` that is being deleted. If po-ba or the human considers the bare-URL initial state a behaviour regression rather than an acceptable simplification, raise it now and the architect will explore a fix that does not re-introduce `nextjs-no-client-side-redirect` (likely: a `useLayoutEffect`-driven one-shot via `router.replace` with a `// react-doctor-allow` and a rationale — least bad of the bad options). Default verdict if no objection: **bare URL on initial load is acceptable**.

---

## 15. Approval log

### Revision 1 — 2026-06-01 — author: system-architect

- Initial draft.
- Accepted D-001 through D-007 as locked by po-ba's intake; added D-008 through D-013 as new architectural decisions for REQ005-specific gaps.
- §3 specifies the US004 signal-cancellable lifecycle context contract (signals, timeout, defer order, no shared helper).
- §4 specifies the US006 AbortController hook contract for `useProject` and `useProjectDocuments`, plus the frozen `lib/api/` function signatures.
- §5 lists the 16 backfill test names verbatim with the branch and mock shape per test.
- §6 specifies the docker-compose service topology, Makefile target shapes, migration / seed strategy.
- §7 chooses path (b) for US009 — path (a) is not reachable from this repo.
- §10 resolves Q1 / Q2 / Q3 from the README's open questions.
- §13 (now §14) raises four open questions for the human to confirm at approval.

### Revision 2 — 2026-06-02 — driver: human feedback pass 1 (po-ba added US010)

- **Trigger:** po-ba folded US010 (React-Doctor baseline regression fixes — top-3 state/effect + 1 security) into REQ005 README per D-008. Architecture must absorb it without re-litigating REQ005's other eight stories.
- **Added §0.1 Executive summary** at the top of the document so a re-approval scan in 60 seconds catches the five US010-driven calls (ref-attach over DOMPurify; reducer state/action contract; auto-select effect deletion; US006→US010 ordering; visual parity guarantees).
- **Added US010 row to §2 File-level touch map** — four touched files (MermaidDiagram.tsx, useDocument.ts, DocumentsTab.tsx, plus the three matching test files with explicit notes on what changes vs stays). `web/package.json` row says NO CHANGE under the default ref-attach path.
- **Added §11 — "US010 contract: React-Doctor regression fixes"** as a new, full-substance section. Covers:
  - §11.1 MermaidDiagram ref-attach fix path (default, recommended), with exact `useEffect` shape + rationale for rejecting DOMPurify + inline suppression.
  - §11.2 `useDocument` reducer contract — state, actions, reducer signature, effect body, public-shape preservation. Locks the action set as `FETCH_STARTED | FETCH_SUCCEEDED | FETCH_FAILED | ABORTED` and clarifies why `hasMore` is not in scope (this is single-document fetch, not paginated).
  - §11.3 `DocumentsTab` redirect contract — delete the auto-select `useEffect`, replace with render-time `selectedDocId = docParam ?? documents?.[0]?.id`. Calls out that the existing `handleSelectDoc` click handler is the existing sibling click path, so **no `ARCHITECTURE_GAP_FOUND` is required**. Notes the side-effect of bare URL on initial load and raises OQ-6.
  - §11.4 US006 ⇄ US010 ordering — US006 first, US010 declares `Depends-on: US006_fe_use_document_abort_controller`. Rationale provided.
  - §11.5 What US010 does NOT change.
- **Added D-014 to §9 Decisions** mirroring po-ba README's D-008 with architect-level confirmation. (Numbering: po-ba's "D-008 (US010 inclusion)" in the README maps to this architecture's **D-014** to avoid colliding with the existing architecture D-008 "no `internal/lifecycle/` helper".) See §9.
- **Renumbered §11→§12, §12→§13, §13→§14, §14→§15.** Section content otherwise preserved verbatim. §12 (skill / hook usage) gained one line: react-doctor skill is now the primary review gate for US010 (was already enforced on US006).
- **Added R7 + R8** to §14.1 risks (strict-mode double-mount of the ref-attach effect; loss of initial-load URL write in DocumentsTab).
- **Added OQ-5 + OQ-6** to §14.2 (ref-attach vs DOMPurify default; bare-URL acceptability). OQ-4's wording updated to note that "US010" now means React-Doctor regressions, so an MSW-leak-hunt stub would be **US011** if the human wants to fold it in.
- **Frontmatter set back to `Approval: pending_approval`** — this revision returns to the human for re-approval, NOT auto-approved.

Sections **untouched** by this revision: §1 Scope, §3 US004 lifecycle context contract, §4 US006 AbortController hook contract, §5 US005 test backfill matrix, §6 US008 e2e stack-up, §7 US009 worktree origin contract, §8 API-contract impact (still N/A), §10 Open-question resolutions Q1/Q2/Q3, and D-001–D-013 in §9. Tasks that downstream agents may already have drafted for those untouched sections do NOT need rolling back.

US010-specific downstream implications the orchestrator should be aware of:
- tech-lead (Phase 2) will produce two US010 FE tasks (typical split: one for `MermaidDiagram` ref-attach + one for `useDocument` reducer + `DocumentsTab` selection — exact split is tech-lead's call) with `Depends-on:` pointing at the corresponding US006 task.
- tester (Phase 2) will add FCT-* entries to `US010_fe_unit_tests.md` covering: ref-attach `<svg>` child assertion under strict mode, reducer-action-level unit tests, render-time selection without URL write, react-doctor diff baseline check.
- No e2e impact — visual parity means existing Robot suites pass unchanged. No new e2e files needed.
- No BE impact — US010 is FE-only.

### Revision 3 — 2026-06-02 — driver: human approval

- Approved by human at 2026-06-02T02:32:14Z.
- OQ-5 (ref-attach over DOMPurify+suppression) and OQ-6 (bare URL on initial load) accepted as proposed.
- Architecture is now LOCKED for Phase 2. Any downstream gap surfaces via `ARCHITECTURE_GAP_FOUND` → re-enters HARD STOP.
