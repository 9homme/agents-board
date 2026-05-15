---
description: Human approval gate for architecture.md. Flips Approval to `approved` and stamps Approved-by + Approved-at. Pre-requisite for /phase2.
argument-hint: <REQ_ID> [approver name]
---

# /approve-architecture — human gate

You are the orchestrator running the human approval step for an architecture document.

## Input

`$ARGUMENTS` is `<REQ_ID> [approver name]`. If `<REQ_ID>` is missing, ask the user once and wait.

## Steps

1. **Resolve the REQ folder.** Use Glob to find `docs/requirements/REQ[ID]_*/architecture.md`. If multiple match, list them and ask the user to disambiguate. If none match, abort and report.
2. **Read the file.** Confirm `Approval:` is currently `pending_approval`. If it's already `approved`, tell the user and stop. If it's `draft` or `changes_requested`, abort — the architect must re-submit it as `pending_approval` first.
3. **Confirm with the user one more time** via `AskUserQuestion`:
   - "About to approve `<path>`. The recorded approver will be `<approver name or 'human'>` and the timestamp will be `<ISO now>`. Proceed?"
   - Approve / Cancel.
   - This double-check is intentional — flipping to `approved` unblocks Phase 2.
4. **Edit the file** to set:
   - `Approval: approved`
   - `Approved-by: <approver name or 'human'>`
   - `Approved-at: <ISO timestamp now>`
5. **Append an entry** to the `## Approval log` section:
   ```
   ### Revision N — YYYY-MM-DD — driver: human approval
   - Approved by <approver name or 'human'> at <ISO timestamp>.
   ```
   (`N` = next number after the highest existing revision entry.)
6. **Report back** to the user: path, new approval status, next command (`/phase2 <REQ_ID>`).

## Rules

- This command is the ONLY way `Approval: approved` should ever be set. No agent (including system-architect) flips it themselves.
- If the user wants changes instead of approval, tell them to run `/phase1 <REQ_ID>` again with feedback (the architect re-enters HARD STOP loop).
