# Product Vision — agents-board

## What it is

agents-board is a **control tower for the virtual engineering team** (po-ba, system-architect, tech-lead-planner, tech-lead-reviewer, tester, be-dev, fe-dev). It is not just a task tracker. It is the live operational interface through which a human runs, monitors, and intervenes in the full multi-agent engineering pipeline.

Every project on the board maps to a **real local directory** on disk (`projects.path` — NOT NULL, validated). That path is the physical anchor that lets the control tower know where each team's work actually lives.

## Core capabilities (current + planned)

### Board (current)
- Projects, Requirements, User Stories, Tasks, Documents — full hierarchy
- Task status state machine enforced server-side
- MCP tools for agent interactions (create, update, read entities)
- HTTP REST API for web UI reads

### Control tower (planned, shapes all architectural decisions)

| Capability | What it means |
|---|---|
| **Live agent activity** | See which subagent is running, what task it owns, how long it's been running |
| **Response stream** | Read the live output of any running agent command in real time |
| **Phase commands from UI** | Trigger `/phase1`, `/phase2`, `/phase3` from the board — no terminal required |
| **Permission gates** | When an agent requests a tool call (Bash, Write, Edit, etc.), human can Allow or Deny from the board |
| **Control mode** | Toggle Auto (agents proceed without asking) vs Manual (every tool call requires explicit approval) |
| **Team chat** | Send messages to a specific agent or broadcast to the whole team during a live run |
| **Live tracking** | See per-task progress, which agent holds it, last status change, current review pass count |

## Architectural principles that flow from this vision

**Full REST hierarchy** (`/projects/:pid/requirements/:rid/user-stories/:usid/tasks/:tid`)  
The control tower navigates by project → requirement → story → task. Flat or partial routes break the navigation model.

**`projects.path` is NOT NULL everywhere**  
Without a real disk path, the control tower cannot locate the working tree, cannot show diffs, cannot link live agent output to the right project. Every project must link to a real local directory.

**Requirement creation is MCP-only**  
Humans trigger phases; agents create the artefacts. po-ba creates the requirement DB record via `create_requirement` at the end of Phase 1. The web UI reads requirements — it does not create them.

**MCP = agent interface; HTTP = human interface**  
MCP tools carry the write surface for agents. HTTP endpoints carry the read surface for the web UI. This separation keeps the control tower's view consistent and prevents agents from bypassing the workflow gates.

**`blocked_review_gate` is a first-class task status**  
The control tower must distinguish "code needs fixing" (changes_requested) from "the gate/tooling is broken" (blocked_review_gate). Without this distinction, a broken CI environment looks like a failing implementation.

**`docs/system_architecture.md` is the single source of truth**  
system-architect owns it. Updated at every architecture approval. Agents read it instead of scanning source code. This keeps context windows small and decisions consistent as the system grows.

## What this means for future REQs

Before designing any new feature, ask:
1. Does it help the human see what agents are doing?
2. Does it let the human intervene (approve, deny, redirect)?
3. Does it shorten the feedback loop between human intent and agent action?

If yes → high priority. If it's purely internal plumbing with no control-tower surface → defer or fold into a larger REQ.
