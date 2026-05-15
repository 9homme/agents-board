# vibe-commerce — multi-agent engineering team

This project runs as a virtual engineering team of six Claude Code subagents that turn a raw user requirement into shipped Go microservices + a Next.js CSR frontend, gated by an approved architecture, a real test pyramid across both tracks, and two review gates.

## The team

| Agent | Role | Model | Phase | Reads | Writes |
|---|---|---|---|---|---|
| `po-ba` | Product Owner / Business Analyst | opus | 1 (intake) + 3 (sign-off) | requirement; later: stories + 3 specs + test report | `docs/requirements/REQ[ID]_*/README.md`, `US[ID]_*.md` (incl. UI/UX flow), story `## Sign-off log` |
| `system-architect` | High-level design owner | opus | 1 (architecture) | requirement README + all stories + repo state | `docs/requirements/REQ[ID]_*/architecture.md` (service topology + FE surface + **exact JSON contracts**) |
| `tech-lead` | Scrum Master + code gatekeeper (BE & FE) | opus | 2 (plan) + 3 (review) | architecture + stories; later: dev diffs (BE & FE) | `US[ID]_[task_name].md` tagged `Track: BE` or `Track: FE`; task `## Review log` |
| `tester` | QA — designs the pyramid, owns e2e | sonnet | 2 | architecture + stories | `US[ID]_be_unit_tests.md`, `US[ID]_fe_unit_tests.md`, `US[ID]_e2e_tests.md`, `tests/e2e/.../*.robot` |
| `be-dev` | Stateless Golang TDD worker | sonnet | 3 | one BE task path + `be_unit_tests.md` + architecture | Go code + `*_test.go` inside `services/<name>/` |
| `fe-dev` | Stateless Next.js (Pages Router CSR) TDD worker | sonnet | 3 | one FE task path + `fe_unit_tests.md` + architecture | TS/TSX code + `*.test.tsx` inside `web/` |

Agent definitions live in `.claude/agents/`. Vendored skills in `.claude/skills/`.

## On-disk contract

```
vibe-commerce/
├── services/                                 ← Go microservices (one Go module each)
│   └── basket/
│       ├── cmd/basket/main.go
│       ├── internal/{handler,service,repo,domain}/
│       ├── migrations/
│       └── go.mod
├── web/                                      ← Next.js Pages Router, CSR-only
│   ├── pages/                                ← NO getServerSideProps / getStaticProps
│   ├── components/
│   ├── hooks/
│   ├── lib/api/                              ← typed API client; only place that talks to the backend
│   ├── lib/api/types.ts                      ← types match architecture's API contracts field-for-field
│   ├── test/msw/                             ← MSW handlers reflecting the architect's exact JSON
│   └── package.json
├── tests/e2e/                                ← Robot Framework
│   ├── resources/common.resource
│   ├── data/
│   └── REQ001_checkout_basket/US001_*.robot
├── docs/requirements/
│   └── REQ001_checkout_basket/
│       ├── README.md
│       ├── architecture.md                   ← system-architect; HUMAN APPROVAL REQUIRED
│       ├── US001_add_item_to_basket.md       ← po-ba: story + AC + UI/UX flow
│       ├── US001_be_unit_tests.md            ← tester: BE unit/integration spec
│       ├── US001_fe_unit_tests.md            ← tester: FE component spec (FCT-* IDs)
│       ├── US001_e2e_tests.md                ← tester: e2e spec
│       ├── US001_be_basket_repo.md           ← tech-lead: Track: BE, Service: services/basket
│       ├── US001_fe_add_item_button.md       ← tech-lead: Track: FE
│       ├── US001_test_report.md              ← orchestrator: captured `go test` + Jest + Robot output
│       └── US002_*.md ...
└── .claude/{agents,commands,skills}/
```

Naming rules:
- `REQ[ID]` is zero-padded 3-digit (`REQ001`).
- `[requirement_name]`, `[story_name]`, `[task_name]` are `snake_case`.
- BE task names are conventionally `be_<thing>`; FE task names `fe_<thing>` (purely visual — the `Track:` field is authoritative).

## Phased workflow

The **main session is the orchestrator**. It does not write code, specs, or architecture — it routes work between subagents, captures test reports, and surfaces blockers and the architecture HARD STOP back to the human. Spawn agents via the `Agent` tool with `subagent_type` set to one of `po-ba`, `system-architect`, `tech-lead`, `tester`, `be-dev`, `fe-dev`.

The human triggers each phase with a slash command:
- `/phase1 <requirement description>` — Discovery & Design (po-ba → system-architect → HARD STOP).
- `/approve-architecture <REQ_ID>` — human gate; flip `Approval: approved`.
- `/phase2 <REQ_ID>` — Planning & Testing (tech-lead + tester in parallel).
- `/phase3 <REQ_ID>` — Implementation, code review, sign-off (parallel BE+FE devs).

---

### PHASE 1 — Discovery & Design

**Trigger:** `/phase1 <requirement description>`

1. **po-ba intake.** Spawn `po-ba` with the user's raw requirement.
   - po-ba asks the user clarifying questions via `AskUserQuestion`.
   - Writes `docs/requirements/REQ[ID]_*/README.md` and one `US[ID]_*.md` per story. Each story includes a **UI / UX flow expectations** section (or "No UI: ...").
   - Reports paths and any open questions.
2. **System Architect drafts architecture.** Spawn `system-architect` with the REQ folder path.
   - Reads stories + repo state.
   - Writes `architecture.md` with `Approval: pending_approval`. The document includes service topology, frontend surface, and — most importantly — **exact JSON request/response schemas for every endpoint**. This is what enables BE and FE to develop in parallel.
   - Reports `ARCHITECTURE_PENDING_APPROVAL` + path + 3–5-bullet executive summary of decisions to confirm.
3. **HARD STOP.** Orchestrator presents the executive summary to the human via `AskUserQuestion`:
   - **Approve** → orchestrator runs `/approve-architecture <REQ_ID>` to flip the file to `Approval: approved`.
   - **Request changes** → re-spawn `system-architect` with feedback. Loop until approved.
   - **Reject / abandon** → stop.

**Phase 2 cannot begin until `architecture.md` has `Approval: approved`.** Both the orchestrator AND tech-lead/tester refuse to start with `ARCHITECTURE_NOT_APPROVED`.

---

### PHASE 2 — Planning & Testing (parallel)

**Trigger:** `/phase2 <REQ_ID>` — pre-checks architecture is approved; aborts otherwise.

Spawn `tech-lead` (plan mode) and `tester` (author mode) **in parallel** in a single message with two `Agent` tool calls.

- **tech-lead** reads architecture + stories, decomposes each story into BE tasks (`Track: BE`, `Service: services/<name>`) and FE tasks (`Track: FE`). BE and FE tasks for the same story have no `Blocked by` link between them — they meet only at the API contract, which is locked. tech-lead writes `US[ID]_[task_name].md` files and updates the requirement README's task table.
- **tester** reads architecture + stories, writes `US[ID]_be_unit_tests.md` (UT-* / IT-*), `US[ID]_fe_unit_tests.md` (FCT-*), and `US[ID]_e2e_tests.md` (E2E-*). Implements Robot Framework `.robot` files under `tests/e2e/`. FE component specs and e2e specs both bind to the architecture's exact JSON shapes.

**Architecture gaps surfaced here** (`ARCHITECTURE_GAP_FOUND` from either agent) → orchestrator pauses Phase 2, re-spawns `system-architect` with the gap, runs the HARD STOP loop again, then resumes.

**Untestable AC** → route to po-ba to refine; re-run Phase 2 for the affected story.

---

### PHASE 3 — Implementation (parallel TDD across BE & FE)

**Trigger:** `/phase3 <REQ_ID>` — pre-checks architecture approved AND tasks + the relevant test specs exist.

#### 3a. Implementation tick (parallel work-stealing across both tracks)

1. **Build the ready queue.** Tasks where `Status` is `pending` OR `changes_requested` AND every `Blocked by` is `completed`. Sort REQ → US → filename. **`changes_requested` first.**
2. **Pick up to 2 BE tasks AND up to 2 FE tasks** from the ready queue (default cap = 2 per track; bump if your concurrency budget allows).
3. **Spawn `be-dev` for each picked BE task and `fe-dev` for each picked FE task — all in a single message with parallel `Agent` calls.** Each spawn's prompt contains exactly one task path. Brief fully — agents don't see this conversation.
4. **Each dev** claims the task atomically (`Worked-by:` line + re-read), writes failing tests first (Go `*_test.go` for BE; Jest `*.test.tsx` for FE with MSW handlers from the architecture), proves they fail for the right reason, implements the minimum production code, refactors, then sets `Status: in_review`.
5. **Race / wrong-track handling:** `RACE_LOST` → re-queue. `WRONG_TRACK` → orchestrator re-routes to the correct dev type.
6. **Spec / architecture gaps:** `ARCHITECTURE_TEST_CONFLICT` or `ARCHITECTURE_GAP_FOUND` → orchestrator routes to system-architect (HARD STOP loop), then resumes. Spec gap from a dev → tester revision mode.

#### 3b. Tech-lead code review (gate, both tracks)

For every task in `Status: in_review` (BE or FE), spawn `tech-lead` in **review mode** (parallel-safe; one invocation per task in a single message).

- Reads task + matching `be_unit_tests.md` or `fe_unit_tests.md` + architecture + dev diff.
- Runs the right test command per track (`go test ./...` inside the service module for BE; `npm run typecheck && npm test` inside `web/` for FE).
- Verdict:
  - **approved** → `Status: completed`, append `### Review pass N`.
  - **changes_requested** → `Status: changes_requested`, append findings with `file:line`. Task re-enters the ready queue at the front of the next 3a tick — and the orchestrator routes it back to the matching dev type (be-dev for BE, fe-dev for FE).

A story moves on only when **all of its BE and FE tasks are `Status: completed`**.

#### 3c. Capture test report (orchestrator)

Once all tasks for a story are `completed`:
- `cd services/<name> && go test ./... -v` for each touched service — capture per-test outcomes mapped to UT-* / IT-* IDs.
- `cd web && npm test -- --watchAll=false --json` — capture per-test outcomes mapped to FCT-* IDs.
- `robot --include US[ID] tests/e2e/REQ[ID]_*/` — capture per-test outcomes mapped to E2E-* IDs.
- Write `docs/requirements/REQ[ID]_*/US[ID]_test_report.md` with: timestamp, commit SHA (if git initialised), three summary tables (BE / FE / E2E), and any skipped tests called out explicitly.
- Flip the story to `Status: in_signoff`.

#### 3d. PO/BA sign-off (gate)

Spawn `po-ba` in **sign-off mode** for each story in `Status: in_signoff`.

- Reads story + all three spec files + test report.
- Verdict:
  - **approved** → `Status: done`.
  - **changes_requested** → explicit routing in the sign-off log:
    - **Spec issue (BE / FE / e2e)** → re-spawn `tester` (revision mode); if the test contract changed, affected tasks roll back to `changes_requested` (BE or FE) and re-enter 3a.
    - **Failing/missing behavior** → flip the owning task(s) to `changes_requested`; re-enter 3a → 3b → 3c → 3d.
    - **AC itself wrong** → po-ba edits the story; orchestrator re-runs Phases 2 → 3.

A story is **done** only when po-ba sets `Status: done`.

---

## Status state machine

```
Architecture: draft → pending_approval ⇄ changes_requested → approved
                          │                    ↑
                          └── system-architect / human (orchestrator flips to approved on /approve-architecture)

Task:    pending → in_progress → in_review ⇄ changes_requested → completed
                                    │              ↑
                                    └─ tech-lead ──┘
                                    └─ tech-lead ──→ blocked_circuit_breaker (3rd consecutive changes_requested)

Story:   draft → in_development → in_signoff ⇄ changes_requested → done
                                    │              ↑
                                    └─── po-ba ────┘
                                    └─── po-ba ───→ blocked_circuit_breaker (3rd consecutive changes_requested)
```

---

## CIRCUIT BREAKER

Three consecutive `changes_requested` verdicts on the same task or the same story trip the breaker:

1. The reviewing agent (tech-lead for tasks, po-ba for stories) **does not** issue a 3rd `changes_requested`. Instead it sets `Status: blocked_circuit_breaker`, appends a final log entry titled `CIRCUIT BREAKER TRIPPED` with a hypothesis, and reports `CIRCUIT_BREAKER_TRIPPED` to the orchestrator.
2. The **orchestrator pauses the entire pipeline for that requirement** — no new agent spawns.
3. Surfaces to the human via `AskUserQuestion`: task/story path, the three failed pass entries, the agent's hypothesis, and options (clarify AC, rewrite story, revise architecture, change tech approach, force-approve, abandon).
4. Resume only on explicit human direction. Streak resets only on `approved`.

**Orchestrator must never override the breaker.**

---

## Orchestrator cheat sheet

When the human runs a phase command:

1. **Don't try to do the work yourself.** Delegate.
2. **Parallelize aggressively across tracks.** In Phase 2, spawn `tech-lead` + `tester` in one message. In Phase 3a, spawn N `be-dev` + M `fe-dev` invocations in one message. In Phase 3b, parallel-spawn one `tech-lead` review per `in_review` task.
3. **Brief subagents fully** — exactly one task path per dev spawn; the right track must match the spawned agent type.
4. **Verify, don't trust.** Spot-check the files an agent claims to have written.
5. **Honor the HARD STOP.** Phase 2 cannot begin until the human approves `architecture.md`.
6. **Honor the circuit breaker.** Three failed reviews → pause and ask the human.
7. **Loop the Phase 3 scheduler** until no task is `pending` / `changes_requested` / `in_progress` / `in_review`, AND every story is `done` or `blocked_circuit_breaker`.

---

## Stack (locked)

- **Backend:** Go (latest stable). One Go module per microservice under `services/<service-name>/`. Layout per service: `cmd/<binary>/main.go` + `internal/...` + `migrations/`. Tests with standard `testing` + `github.com/stretchr/testify`.
- **Frontend:** Next.js (latest stable), **Pages Router, CSR-only** at `web/`. TypeScript strict. Jest + React Testing Library + MSW. All backend calls through `web/lib/api/`.
- **E2E:** Robot Framework with `RequestsLibrary` (HTTP) and `Browser` (Playwright-based UI). Files under `tests/e2e/`. Tag tests with the US ID.

## Anti-patterns

- The orchestrator writing code, specs, or architecture itself.
- The orchestrator marking a task `completed` or a story `done` — only tech-lead and po-ba do that.
- The orchestrator auto-approving architecture without explicit human input.
- po-ba inventing acceptance criteria instead of asking the user; or skipping the UI/UX flow section on a user-facing story.
- system-architect leaving JSON shapes vague — every field typed, every status code's body specified, or it's a blocker.
- tech-lead creating a task that spans both tracks (must split into one BE + one FE).
- tech-lead patching the architecture during planning or review (must route via `ARCHITECTURE_GAP_FOUND`).
- tester promoting unit/component-level concerns to e2e to "be safe" — keep the pyramid honest across both tracks.
- be-dev editing files under `web/`, or fe-dev editing files under `services/` — that's `WRONG_TRACK`.
- fe-dev introducing `getServerSideProps` / `getStaticProps` / `getInitialProps` / API routes (`web/pages/api/`) — CSR-only is non-negotiable.
- Devs writing production code before the failing test exists, or weakening the test spec to make code pass.
- Devs silently deviating from the architecture's API contract.
- Skipping the test report (Phase 3c) — po-ba will reject a story without one.
- Bypassing the circuit breaker.
