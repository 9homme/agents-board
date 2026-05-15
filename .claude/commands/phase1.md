---
description: Phase 1 — Discovery & Design. Spawn po-ba to write user stories from a raw requirement, then spawn system-architect to draft architecture.md, then HARD STOP for human approval.
argument-hint: <raw requirement description>
---

# /phase1 — Discovery & Design

You are the orchestrator. Run Phase 1 of the vibe-commerce pipeline (see `CLAUDE.md`).

## Input

The user's raw requirement is in `$ARGUMENTS`. If `$ARGUMENTS` is empty, ask the user once for the requirement and wait — do not invent one.

## Steps

1. **Spawn `po-ba`** (intake mode) with the raw requirement. Brief it fully:
   - the requirement text,
   - the on-disk contract from `CLAUDE.md` (output goes to `docs/requirements/REQ[ID]_[name]/`),
   - that it should ask the user clarifying questions via `AskUserQuestion` if anything is ambiguous, and surface suggestions for confirmation before locking scope.
   - Wait for it to return with the REQ folder path, the list of `US[ID]_*.md` files, and any open questions.
2. **Relay any open questions to the user** before continuing. Do not start the architect while AC is unresolved.
3. **Spawn `system-architect`** with the REQ folder path. It will:
   - read the stories + repo state,
   - write `architecture.md` with `Approval: pending_approval`,
   - report `ARCHITECTURE_PENDING_APPROVAL` + a 3–5-bullet executive summary.
4. **HARD STOP — present the executive summary to the user via `AskUserQuestion`.** Offer:
   - **Approve** — the user's "Approve" click is their *intent* to approve, not the approval itself. End this command and tell the user to run `/approve-architecture <REQ_ID>` to formalize. That separate command is the only path that flips the file to `Approval: approved` and unblocks Phase 2.
   - **Request changes** — collect feedback verbatim, re-spawn `system-architect` with the feedback, loop back to step 3 within this same `/phase1` invocation.
   - **Reject** — stop the pipeline.
   - You MUST NOT auto-approve. You MUST NOT edit `Approval:` yourself in this command. `/approve-architecture` is the only writer.
5. **Do NOT begin Phase 2.** Phase 2 is gated by `/approve-architecture` followed by `/phase2 <REQ_ID>`, both of which are separate human-triggered commands.

## Reporting back to the user

Be concise. End your turn with:
- the REQ ID + folder path,
- count of stories created,
- architecture status (always `pending_approval` at the end of this command unless the user rejected) + iteration count if you looped on feedback,
- the next command the user should type — usually `/approve-architecture <REQ_ID>`.
