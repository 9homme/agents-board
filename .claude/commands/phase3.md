---
description: Phase 3 — Implementation. Runs the work-stealing scheduler with parallel BE+FE devs, tech-lead review gate (per track), test report capture (Go + Jest + Robot), and po-ba sign-off gate, until every story is `done` or the circuit breaker trips.
argument-hint: <REQ_ID>
---

# /phase3 — Implementation, review, sign-off

You are the orchestrator. Run Phase 3 of the vibe-commerce pipeline (see `CLAUDE.md`).

## Input

`$ARGUMENTS` is `<REQ_ID>`. If missing, ask once and wait.

## Pre-checks (abort if any fail)

1. Resolve `docs/requirements/REQ[ID]_*/` via Glob. Exactly one match.
2. `architecture.md` exists with `Approval: approved`.
3. At least one `US[ID]_[task_name].md` task file exists; each has `Track:` set.
4. The relevant test specs exist for each story (`US[ID]_be_unit_tests.md` for any story with BE tasks; `US[ID]_fe_unit_tests.md` for any story with FE tasks; `US[ID]_e2e_tests.md` for all stories).

## Concurrency cap

Default to **2 parallel `be-dev` + 2 parallel `fe-dev`** per scheduler tick (4 devs total) and **N parallel `tech-lead` reviews** where N = current `in_review` count.

## The loop

Repeat until: no task is in `pending` / `changes_requested` / `in_progress` / `in_review`, AND every story is `done` or `blocked_circuit_breaker`.

### 3a. Implementation tick (parallel work-stealing across both tracks)

1. **Build the ready queue.** Tasks with `Status: pending` OR `Status: changes_requested` AND every `Blocked by` resolved to `completed`. Sort REQ → US → filename. **`changes_requested` first.**
2. From the ready queue, pick:
   - up to 2 tasks with `Track: BE`,
   - up to 2 tasks with `Track: FE`.
   Distinct task paths only.
3. **Spawn `be-dev` for each picked BE task and `fe-dev` for each picked FE task — all in a single message with parallel `Agent` calls.** Each prompt contains exactly one task path, plus a one-line note pointing to the latest `### Review pass N` if it's rework.
4. Collect dev reports. Handle:
   - `RACE_LOST` → re-queue, re-pick on next tick.
   - `WRONG_TRACK` → orchestrator routing bug — re-spawn the correct dev type for that task.
   - `MISSING_TASK_PATH` → orchestrator briefing bug — re-brief properly.
   - `ARCHITECTURE_TEST_CONFLICT` or `ARCHITECTURE_GAP_FOUND` → spawn `system-architect` to resolve (HARD STOP loop), then resume.
   - Spec gap reports → spawn `tester` (revision mode) for the right spec file (`be_unit_tests.md` if BE dev raised it; `fe_unit_tests.md` if FE dev raised it), then resume.

### 3b. Tech-lead code review (gate, per track)

For each task now in `Status: in_review`:

1. **Spawn `tech-lead` in review mode** (parallel-safe — one invocation per task in a single message with parallel `Agent` calls). Brief tech-lead with the task path; tech-lead infers track from the file.
2. Collect verdicts:
   - `approved` → task `completed`. Continue.
   - `changes_requested` → task back in the ready queue for the next 3a tick. The track field still says BE or FE, so the next pick automatically routes to be-dev or fe-dev.
   - `CIRCUIT_BREAKER_TRIPPED` → STOP THE PIPELINE for this requirement.

### 3c. Capture test report (orchestrator)

When all tasks for a story are `Status: completed`:

1. For each touched service, run `cd services/<name> && go test ./... -v` — capture per-test outcomes mapped back to UT-* / IT-* IDs from `be_unit_tests.md`.
2. Run `cd web && npm test -- --watchAll=false --json` — capture per-test outcomes mapped to FCT-* IDs from `fe_unit_tests.md`.
3. Run `robot --include US[ID] tests/e2e/REQ[ID]_*/` — capture per-test outcomes mapped to E2E-* IDs.
4. Write `docs/requirements/REQ[ID]_*/US[ID]_test_report.md` with:
   - timestamp, commit SHA if git initialised,
   - **BE summary table** (UT-* / IT-* → pass / fail),
   - **FE summary table** (FCT-* → pass / fail),
   - **E2E summary table** (E2E-* → pass / fail),
   - any skipped tests called out explicitly across all three.
5. Flip the story to `Status: in_signoff`.

If a toolchain isn't installed (no Go, no Node, no Robot) report this back to the user and pause — do not fake a report.

### 3d. PO/BA sign-off (gate)

For each story now in `Status: in_signoff`:

1. **Spawn `po-ba` in sign-off mode.** Brief with the story path.
2. Collect verdict:
   - `approved` → story is `done`.
   - `changes_requested` → po-ba's sign-off log entry tells you where to route:
     - **tester (spec)** → spawn `tester` (revision mode) inline within this command; affected tasks roll back to `changes_requested` (BE or FE per the changed spec) and the loop re-enters 3a.
     - **dev (failing/missing behavior)** → po-ba already flipped the affected task(s) to `changes_requested`; the loop re-enters 3a naturally and routes to the matching dev type.
     - **po-ba (AC rewrite)** → po-ba edits the story. Halt this command, surface the AC change to the user, and instruct them to run `/phase2 <REQ_ID>` (so tester and tech-lead regenerate specs and tasks against the new AC) then `/phase3 <REQ_ID>` again. Do not try to continue 3a with a stale plan against rewritten AC.
   - `CIRCUIT_BREAKER_TRIPPED` → STOP THE PIPELINE.

## Circuit-breaker handling

If any agent reports `CIRCUIT_BREAKER_TRIPPED`:

1. **Halt all loops.** Do not spawn another agent for this requirement.
2. Read the relevant `## Review log` or `## Sign-off log` final entry.
3. Surface to the human via `AskUserQuestion` with options: clarify AC, rewrite story, revise architecture, change tech approach, force-approve, abandon.
4. Resume only on explicit human direction.

## Reporting back

End your turn with:
- counts: BE tasks completed / in-flight / blocked, FE tasks completed / in-flight / blocked, stories done / in-signoff / blocked,
- any test-report path written this run,
- any unresolved blocker,
- next command (usually nothing — Phase 3 finishing means the requirement is done).
