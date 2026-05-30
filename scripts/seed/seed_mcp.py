#!/usr/bin/env python3
"""
Seed mock data into agent-board via MCP over SSE.

Hierarchy per REQ001:
    Project -> Document
            -> User Story -> Task

This script shells out to `curl` to avoid any pip dependency. It:
  1. Spawns `curl -sN <base>/sse` and reads stdout line-by-line in a thread,
     captures sessionId from the `endpoint` event, dispatches `message` events
     back to callers waiting on their JSON-RPC id.
  2. Shells out `curl -s -X POST <base>/message?sessionId=...` for each
     tools/call request, then blocks until the matching response shows up
     on the SSE stream.
  3. Creates 3 projects, each with several documents (rich markdown + mermaid
     to exercise REQ004 rendering), user stories with mixed statuses, and tasks.

Run:
    python3 scripts/seed/seed_mcp.py
    MCP_BASE_URL=http://localhost:8081 python3 scripts/seed/seed_mcp.py
"""
import json
import os
import queue
import re
import shutil
import subprocess
import sys
import threading
from typing import Any

MCP_BASE_URL = os.environ.get("MCP_BASE_URL", "http://localhost:8081")
CALL_TIMEOUT = 30
HANDSHAKE_TIMEOUT = 10


class MCPClient:
    """Minimal MCP SSE client built on `curl`. Keeps the SSE socket open,
    dispatches JSON-RPC responses to callers waiting on their id."""

    def __init__(self, base_url: str):
        self.base_url = base_url.rstrip("/")
        self.session_id: str | None = None
        self._next_id = 0
        self._pending: dict[int, queue.Queue] = {}
        self._proc: subprocess.Popen | None = None
        self._sse_thread: threading.Thread | None = None
        self._stop = threading.Event()

    def connect(self) -> str:
        if shutil.which("curl") is None:
            raise RuntimeError("curl is required but not on PATH")

        ready = threading.Event()
        captured_id: list[str] = []

        def reader():
            assert self._proc is not None
            event_name = ""
            try:
                for raw in self._proc.stdout:  # type: ignore[union-attr]
                    if self._stop.is_set():
                        return
                    line = raw.rstrip("\n").rstrip("\r")
                    if line == "":
                        event_name = ""
                        continue
                    if line.startswith("event:"):
                        event_name = line[len("event:"):].strip()
                        continue
                    if not line.startswith("data:"):
                        continue
                    data = line[len("data:"):].strip()
                    if event_name == "endpoint":
                        m = re.search(r"sessionId=([a-zA-Z0-9-]+)", data)
                        if m and not captured_id:
                            captured_id.append(m.group(1))
                            ready.set()
                        continue
                    if event_name == "message":
                        try:
                            msg = json.loads(data)
                        except json.JSONDecodeError:
                            continue
                        rpc_id = msg.get("id")
                        if rpc_id is None:
                            continue
                        q = self._pending.get(rpc_id)
                        if q is not None:
                            q.put(msg)
            except Exception as e:
                if not self._stop.is_set():
                    print(f"[sse-reader] error: {e}", file=sys.stderr)

        self._proc = subprocess.Popen(
            ["curl", "-sN", f"{self.base_url}/sse"],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            bufsize=1,
        )
        self._sse_thread = threading.Thread(target=reader, daemon=True)
        self._sse_thread.start()

        if not ready.wait(HANDSHAKE_TIMEOUT):
            raise RuntimeError(f"timed out waiting for sessionId from {self.base_url}/sse")
        self.session_id = captured_id[0]
        return self.session_id

    def close(self):
        self._stop.set()
        if self._proc is not None:
            try:
                self._proc.terminate()
                self._proc.wait(timeout=2)
            except Exception:
                try:
                    self._proc.kill()
                except Exception:
                    pass

    def call_tool(self, name: str, arguments: dict[str, Any]) -> Any:
        if self.session_id is None:
            raise RuntimeError("not connected")
        self._next_id += 1
        rpc_id = self._next_id
        q: queue.Queue = queue.Queue(maxsize=1)
        self._pending[rpc_id] = q
        payload = {
            "jsonrpc": "2.0",
            "id": rpc_id,
            "method": "tools/call",
            "params": {"name": name, "arguments": arguments},
        }
        url = f"{self.base_url}/message?sessionId={self.session_id}"
        try:
            post = subprocess.run(
                [
                    "curl", "-s", "-S", "-o", "/dev/null", "-w", "%{http_code}",
                    "-X", "POST",
                    "-H", "Content-Type: application/json",
                    "-d", json.dumps(payload),
                    url,
                ],
                check=False,
                capture_output=True,
                text=True,
                timeout=CALL_TIMEOUT,
            )
            if post.returncode != 0:
                raise RuntimeError(f"curl POST failed (exit {post.returncode}): {post.stderr.strip()}")
            status = post.stdout.strip()
            if not status.startswith("2"):
                raise RuntimeError(f"POST returned HTTP {status} for tool {name}")
            try:
                msg = q.get(timeout=CALL_TIMEOUT)
            except queue.Empty as e:
                raise RuntimeError(f"timed out waiting for response to {name} (id={rpc_id})") from e
        finally:
            self._pending.pop(rpc_id, None)

        if "error" in msg:
            raise RuntimeError(f"tool {name} JSON-RPC error: {msg['error']}")
        result = msg.get("result", {})
        if result.get("isError"):
            content = result.get("content", [])
            text = content[0].get("text", "") if content else ""
            raise RuntimeError(f"tool {name} returned isError: {text}")
        content = result.get("content", [])
        if not content:
            return {}
        text = content[0].get("text", "")
        try:
            return json.loads(text)
        except json.JSONDecodeError:
            return {"_raw": text}


# ---------------------------------------------------------------------------
# Mock dataset
# ---------------------------------------------------------------------------

MERMAID_FLOWCHART = """```mermaid
flowchart LR
    Client[Browser] -->|HTTPS| Edge[Edge / CDN]
    Edge --> Web[Next.js CSR]
    Web -->|fetch| API[Echo API]
    API --> DB[(PostgreSQL)]
    API --> MCP[MCP over SSE]
```"""

MERMAID_SEQUENCE = """```mermaid
sequenceDiagram
    participant U as User
    participant FE as Next.js (CSR)
    participant BE as Echo API
    participant DB as Postgres

    U->>FE: open /projects/:id
    FE->>BE: GET /api/v1/projects/:id
    BE->>DB: SELECT ... FROM projects WHERE id=$1
    DB-->>BE: row
    BE-->>FE: 200 {id,name,description,...}
    FE-->>U: rendered header + tab switcher
```"""

PROJECTS: list[dict[str, Any]] = [
    {
        "name": "Vibe Commerce Platform",
        "description": "Headless commerce platform — basket, checkout, fulfillment microservices fronted by a Next.js storefront.",
        "documents": [
            {
                "title": "Architecture overview",
                "content": (
                    "# Architecture overview\n\n"
                    "The platform is a set of Go microservices behind an Edge gateway, "
                    "consumed by a **CSR-only** Next.js storefront. State lives in Postgres; "
                    "no shared in-memory cache between services.\n\n"
                    "## Topology\n\n"
                    f"{MERMAID_FLOWCHART}\n\n"
                    "## Why CSR-only\n\n"
                    "- Predictable deploy story (static export → CDN)\n"
                    "- No SSR data-fetching coupling between FE and BE\n"
                    "- Easier to reason about MSW mocking in tests\n\n"
                    "## Trade-offs\n\n"
                    "| Concern | CSR (current) | SSR alternative |\n"
                    "|---|---|---|\n"
                    "| TTFB | fast | slow (depends on BE) |\n"
                    "| SEO | needs prerender | native |\n"
                    "| Hosting | static CDN | Node runtime |\n"
                ),
            },
            {
                "title": "Add-to-basket request flow",
                "content": (
                    "# Add-to-basket request flow\n\n"
                    f"{MERMAID_SEQUENCE}\n\n"
                    "## Idempotency\n\n"
                    "Clients pass `Idempotency-Key` on POST. The basket service keys "
                    "the row by `(user_id, idempotency_key)`; replays return the original 201."
                ),
            },
            {
                "title": "Local dev runbook",
                "content": (
                    "# Local dev runbook\n\n"
                    "## Prereqs\n\n"
                    "- Go 1.22+\n"
                    "- Node 20+\n"
                    "- Docker (for Postgres)\n\n"
                    "## Bring up the stack\n\n"
                    "```bash\n"
                    "docker compose up -d postgres\n"
                    "cd services/agent-board && go run ./cmd/api-server\n"
                    "cd web && npm run dev\n"
                    "```\n\n"
                    "## Run tests\n\n"
                    "```bash\n"
                    "cd services/agent-board && go test ./...\n"
                    "cd web && npm test -- --watchAll=false --forceExit\n"
                    "robot --include US001 tests/e2e/REQ004_project_detail_page/\n"
                    "```"
                ),
            },
            {
                "title": "API conventions",
                "content": (
                    "# API conventions\n\n"
                    "- All endpoints under `/api/v1`.\n"
                    "- Bare-object responses for single resources (no envelope).\n"
                    "- Collection endpoints use `{\"<plural>\":[…]}` envelopes (never `null`, never bare arrays).\n"
                    "- Errors use `{\"code\":\"CODE\",\"message\":\"…\"}` with the right HTTP status.\n\n"
                    "## Status codes\n\n"
                    "| Code | When |\n"
                    "|---|---|\n"
                    "| 200 | OK with body |\n"
                    "| 201 | Created |\n"
                    "| 204 | OK no body |\n"
                    "| 400 | malformed request |\n"
                    "| 404 | resource not found |\n"
                    "| 409 | conflict (e.g. unique constraint) |\n"
                    "| 500 | server fault, retryable |\n"
                ),
            },
        ],
        "stories": [
            {
                "title": "Browse the catalog from the storefront",
                "description": "As a shopper I want to browse the product catalog so I can decide what to buy.",
                "status": "done",
                "tasks": [
                    {"title": "BE: GET /api/v1/products list endpoint", "description": "Paginated list with name + price.", "status": "completed"},
                    {"title": "FE: ProductGrid component", "description": "RTL tests + MSW-backed integration.", "status": "completed"},
                    {"title": "E2E: visitor lands on /products", "description": "Robot suite covering navigation.", "status": "completed"},
                ],
            },
            {
                "title": "Add item to basket",
                "description": "As a shopper I want to add a product to my basket so I can review and check out.",
                "status": "in_development",
                "tasks": [
                    {"title": "BE: POST /api/v1/baskets/me/items", "description": "Idempotent insert keyed on Idempotency-Key.", "status": "in_review"},
                    {"title": "FE: Add-to-basket button + toast", "description": "Optimistic UI + rollback on error.", "status": "in_progress"},
                    {"title": "FE: useBasket hook", "description": "Wraps fetch + AbortController + retry.", "status": "pending"},
                ],
            },
            {
                "title": "Persist basket across sessions",
                "description": "As a shopper I want my basket to survive logout so I don't lose what I picked.",
                "status": "draft",
                "tasks": [
                    {"title": "BE: basket-by-user-id repo", "description": "DB-backed instead of memory.", "status": "pending"},
                ],
            },
        ],
    },
    {
        "name": "Customer Portal Redesign",
        "description": "Modernize the legacy AngularJS portal — convert to Next.js Pages Router CSR with a real component library.",
        "documents": [
            {
                "title": "Migration plan",
                "content": (
                    "# Migration plan\n\n"
                    "## Phase 1 — strangle the shell\n\n"
                    "Wrap the legacy app in a Next.js shell so we can mount new pages alongside the old ones.\n\n"
                    "## Phase 2 — convert flows page-by-page\n\n"
                    "- [x] Login\n- [x] Dashboard\n- [ ] Account settings\n- [ ] Billing\n- [ ] Support tickets\n\n"
                    "## Phase 3 — retire AngularJS\n\n"
                    "Delete the legacy bundle, drop the iframe, ship the new design system as the only UI.\n"
                ),
            },
            {
                "title": "Design system tokens",
                "content": (
                    "# Design system tokens\n\n"
                    "```ts\n"
                    "export const tokens = {\n"
                    "  color: {\n"
                    "    primary: '#2C5BFF',\n"
                    "    surface: '#FFFFFF',\n"
                    "    text:    '#0B1320',\n"
                    "  },\n"
                    "  space: [0, 4, 8, 12, 16, 24, 32, 48, 64],\n"
                    "  radius: { sm: 4, md: 8, lg: 12 },\n"
                    "};\n"
                    "```"
                ),
            },
            {
                "title": "Accessibility checklist",
                "content": (
                    "# Accessibility checklist\n\n"
                    "- Every interactive element is keyboard-reachable.\n"
                    "- Focus ring visible (`focus-visible:ring-2`).\n"
                    "- Form fields have associated `<label>` (or `aria-label`).\n"
                    "- Modal dialogs trap focus + restore it on close.\n"
                    "- Live regions announce async updates.\n"
                    "- Color contrast ≥ 4.5:1 for text, ≥ 3:1 for UI components."
                ),
            },
        ],
        "stories": [
            {
                "title": "Login page parity with legacy",
                "description": "New login mirrors legacy behavior (validation, error copy, remember-me).",
                "status": "done",
                "tasks": [
                    {"title": "FE: <LoginForm/> + tests", "description": "Validation + submit + a11y.", "status": "completed"},
                    {"title": "FE: session token handling", "description": "Set cookie via API; redirect to /.", "status": "completed"},
                ],
            },
            {
                "title": "Dashboard summary cards",
                "description": "Four cards: account, billing, support, recent activity.",
                "status": "in_signoff",
                "tasks": [
                    {"title": "FE: <SummaryCard/> primitive", "description": "Generic card; subclassed per metric.", "status": "completed"},
                    {"title": "FE: useAccountSummary hook", "description": "AbortController on unmount.", "status": "completed"},
                    {"title": "FE: empty/loading/error states", "description": "Match design system.", "status": "in_review"},
                ],
            },
            {
                "title": "Account settings page",
                "description": "Profile, password, notification prefs.",
                "status": "in_development",
                "tasks": [
                    {"title": "BE: PATCH /api/v1/users/me", "description": "Partial update + audit log.", "status": "in_progress"},
                    {"title": "FE: <ProfileForm/>", "description": "Inline validation.", "status": "pending"},
                    {"title": "FE: <PasswordChangeForm/>", "description": "Confirm new + current.", "status": "pending"},
                ],
            },
        ],
    },
    {
        "name": "Internal Tooling — Agent Board",
        "description": "Eat our own dog food — a Project / Document / User Story / Task board used by the AI engineering team itself.",
        "documents": [
            {
                "title": "Why this exists",
                "content": (
                    "# Why this exists\n\n"
                    "Our multi-agent pipeline produces a lot of markdown — requirement docs, "
                    "user stories, architecture diagrams, test specs, sign-off logs. Reading "
                    "them through `cat` in a terminal works but is hostile to humans. This tool "
                    "gives the team a browser surface to navigate the same artifacts.\n\n"
                    "## In scope\n\n"
                    "- Project → Document browsing\n- Markdown + mermaid rendering\n- URL-deeplink to a specific doc\n\n"
                    "## Out of scope (intentionally)\n\n"
                    "- Editing (tools/CLI remain the source of truth)\n- Auth (LAN-only)\n- Real-time collab\n"
                ),
            },
            {
                "title": "Data model diagram",
                "content": (
                    "# Data model\n\n"
                    "```mermaid\n"
                    "erDiagram\n"
                    "    PROJECT ||--o{ DOCUMENT : has\n"
                    "    PROJECT ||--o{ USER_STORY : has\n"
                    "    USER_STORY ||--o{ TASK : has\n"
                    "    PROJECT {\n"
                    "        uuid id PK\n"
                    "        string name\n"
                    "        string description\n"
                    "    }\n"
                    "    DOCUMENT {\n"
                    "        uuid id PK\n"
                    "        uuid project_id FK\n"
                    "        string title\n"
                    "        text content\n"
                    "    }\n"
                    "    USER_STORY {\n"
                    "        uuid id PK\n"
                    "        uuid project_id FK\n"
                    "        string title\n"
                    "        string status\n"
                    "    }\n"
                    "    TASK {\n"
                    "        uuid id PK\n"
                    "        uuid user_story_id FK\n"
                    "        string title\n"
                    "        string status\n"
                    "    }\n"
                    "```\n"
                ),
            },
            {
                "title": "REQ004 retrospective findings",
                "content": (
                    "# REQ004 retrospective\n\n"
                    "Cross-cutting findings from the project detail page work:\n\n"
                    "1. **Coverage gap** — repo error branches thin (pre-existing). Filed as REQ005/US005.\n"
                    "2. **`context.Background()`** without timeouts in `cmd/api-server/main.go`. Filed as REQ005/US004.\n"
                    "3. **FE hook inconsistency** — only `useDocument` aborts; `useProject` / `useProjectDocuments` don't. Filed as REQ005/US006.\n"
                    "4. **Quality gate script bugs** — `printf --` swallowed failures; FE gate hung on MSW open handle.\n"
                    "5. **Spec defects survived sign-off** — robot import path + keyword arity slipped past tech-lead because no `--dryrun` was in the gate.\n\n"
                    "All five items now have stories under REQ005."
                ),
            },
            {
                "title": "How a new REQ is born",
                "content": (
                    "# How a new REQ is born\n\n"
                    "```mermaid\n"
                    "stateDiagram-v2\n"
                    "    [*] --> phase1: /phase1 <description>\n"
                    "    phase1 --> pending_approval: architect drafts\n"
                    "    pending_approval --> approved: /approve-architecture\n"
                    "    approved --> phase2: /phase2\n"
                    "    phase2 --> phase3: tech-lead + tester ready\n"
                    "    phase3 --> done: po-ba signs off\n"
                    "    done --> [*]\n"
                    "```\n"
                ),
            },
            {
                "title": "Empty-content edge case (you should see no body below)",
                "content": "",
            },
        ],
        "stories": [
            {
                "title": "Dashboard lists all projects",
                "description": "REQ002 baseline — landing on / shows clickable project cards.",
                "status": "done",
                "tasks": [
                    {"title": "BE: GET /api/v1/projects", "description": "List endpoint with envelope.", "status": "completed"},
                    {"title": "FE: <ProjectCard/> + <ProjectList/>", "description": "Card grid + loading/error.", "status": "completed"},
                ],
            },
            {
                "title": "Project detail page with tabs",
                "description": "REQ004/US001 — click card, land on /projects/:id with Documents + User Stories tabs.",
                "status": "done",
                "tasks": [
                    {"title": "BE: GET /api/v1/projects/:id", "description": "Bare-object response.", "status": "completed"},
                    {"title": "FE: detail page + tab switcher", "description": "URL persists active tab.", "status": "completed"},
                    {"title": "FE: ProjectCard becomes Link", "description": "Card wraps Next Link to /projects/:id.", "status": "completed"},
                ],
            },
            {
                "title": "Documents tab — list + select + previewer",
                "description": "REQ004/US002 — sidebar lists docs, clicking one renders in the previewer.",
                "status": "done",
                "tasks": [
                    {"title": "BE: GET /api/v1/projects/:id/documents", "description": "Ordered by updatedAt DESC.", "status": "completed"},
                    {"title": "BE: GET /api/v1/documents/:id", "description": "Bare-object incl. content.", "status": "completed"},
                    {"title": "FE: DocumentsTab + Sidebar + Previewer", "description": "AbortController race-safe.", "status": "completed"},
                ],
            },
            {
                "title": "Markdown + mermaid rendering",
                "description": "REQ004/US003 — GFM, syntax highlighting, mermaid SVG, XSS sanitized.",
                "status": "done",
                "tasks": [
                    {"title": "FE: MarkdownRenderer + error boundary", "description": "react-markdown + rehype-sanitize + rehype-highlight.", "status": "completed"},
                    {"title": "FE: MermaidDiagram", "description": "Lazy import + cleanup on unmount.", "status": "completed"},
                ],
            },
            {
                "title": "Quality hardening retrospective",
                "description": "REQ005 — close the gaps the REQ004 audit surfaced.",
                "status": "in_development",
                "tasks": [
                    {"title": "Fix printf -- bug in run-gate.sh", "description": "Group A must-fix.", "status": "pending"},
                    {"title": "Add --forceExit to FE gate", "description": "Group A must-fix.", "status": "pending"},
                    {"title": "Soft-warn gosec/govulncheck if missing", "description": "Group A must-fix.", "status": "pending"},
                    {"title": "context.Background audit", "description": "Replace with WithTimeout where needed.", "status": "pending"},
                    {"title": "Backfill 14 repo error-branch tests", "description": "Coverage push to ≥95% per file.", "status": "pending"},
                ],
            },
        ],
    },
]


STORY_INITIAL_STATUS = "draft"
TASK_INITIAL_STATUS = "pending"

# REQ003 enforces step-by-step transitions. Define the path from initial state
# to each terminal/intermediate state we want to seed into.
STORY_TRANSITION_PATH: dict[str, list[str]] = {
    "draft":             [],  # no-op
    "in_development":    ["in_development"],
    "in_signoff":        ["in_development", "in_signoff"],
    "done":              ["in_development", "in_signoff", "done"],
    "changes_requested": ["in_development", "in_signoff", "changes_requested"],
}

TASK_TRANSITION_PATH: dict[str, list[str]] = {
    "pending":           [],  # no-op
    "in_progress":       ["in_progress"],
    "in_review":         ["in_progress", "in_review"],
    "completed":         ["in_progress", "in_review", "completed"],
    "changes_requested": ["in_progress", "in_review", "changes_requested"],
}


def reset_all(client: MCPClient) -> tuple[int, int, int, int]:
    """List every existing project and delete it (and its children via list+delete)."""
    p_del = d_del = s_del = t_del = 0
    projects = client.call_tool("list_projects", {}).get("projects", []) or []
    for p in projects:
        pid = p["id"]
        try:
            stories = client.call_tool("list_user_stories", {"projectId": pid}).get("userStories", []) or []
        except Exception:
            stories = []
        for s in stories:
            sid = s["id"]
            try:
                tasks = client.call_tool("list_tasks", {"userStoryId": sid}).get("tasks", []) or []
            except Exception:
                tasks = []
            for t in tasks:
                client.call_tool("delete_task", {"id": t["id"]})
                t_del += 1
            client.call_tool("delete_user_story", {"id": sid})
            s_del += 1
        try:
            docs = client.call_tool("list_documents", {"projectId": pid}).get("documents", []) or []
        except Exception:
            docs = []
        for d in docs:
            client.call_tool("delete_document", {"id": d["id"]})
            d_del += 1
        client.call_tool("delete_project", {"id": pid})
        p_del += 1
    return p_del, d_del, s_del, t_del


def main() -> int:
    reset = "--reset" in sys.argv
    print(f"→ MCP base: {MCP_BASE_URL}    reset={reset}")
    client = MCPClient(MCP_BASE_URL)
    try:
        session_id = client.connect()
    except Exception as e:
        print(f"✗ failed to open SSE session: {e}", file=sys.stderr)
        return 1
    print(f"✓ sessionId: {session_id}")

    if reset:
        print("→ resetting (deleting all existing projects + children)...")
        try:
            p, d, s, t = reset_all(client)
            print(f"✓ reset complete — deleted {p} projects, {d} documents, {s} stories, {t} tasks")
        except Exception as e:
            print(f"✗ reset failed: {e}", file=sys.stderr)
            client.close()
            return 1

    project_count = 0
    doc_count = 0
    story_count = 0
    task_count = 0

    try:
        for proj_spec in PROJECTS:
            proj = client.call_tool(
                "create_project",
                {"name": proj_spec["name"], "description": proj_spec["description"]},
            )
            project_id = proj["id"]
            project_count += 1
            print(f"\n=== Project [{project_id[:8]}…] {proj['name']!r}")

            for doc_spec in proj_spec["documents"]:
                doc = client.call_tool(
                    "create_document",
                    {
                        "projectId": project_id,
                        "title": doc_spec["title"],
                        "content": doc_spec["content"],
                    },
                )
                doc_count += 1
                print(f"   📄 [{doc['id'][:8]}…] {doc['title']!r} ({len(doc_spec['content'])} chars)")

            for story_spec in proj_spec["stories"]:
                # REQ003: stories must be created in the initial state (`draft`).
                # If the spec wants a different status, create then update.
                story = client.call_tool(
                    "create_user_story",
                    {
                        "projectId": project_id,
                        "title": story_spec["title"],
                        "description": story_spec["description"],
                        "status": STORY_INITIAL_STATUS,
                    },
                )
                story_id = story["id"]
                path = STORY_TRANSITION_PATH.get(story_spec["status"], [])
                if path is None:
                    raise RuntimeError(f"unknown story status {story_spec['status']!r}")
                for step in path:
                    story = client.call_tool(
                        "update_user_story",
                        {"id": story_id, "status": step},
                    )
                story_count += 1
                print(f"   📋 [{story_id[:8]}…] {story['title']!r} (status={story['status']})")

                for task_spec in story_spec["tasks"]:
                    # REQ003: tasks must be created in `pending`. Same create-then-update pattern.
                    task = client.call_tool(
                        "create_task",
                        {
                            "userStoryId": story_id,
                            "title": task_spec["title"],
                            "description": task_spec["description"],
                            "status": TASK_INITIAL_STATUS,
                        },
                    )
                    task_id = task["id"]
                    path = TASK_TRANSITION_PATH.get(task_spec["status"], [])
                    if path is None:
                        raise RuntimeError(f"unknown task status {task_spec['status']!r}")
                    for step in path:
                        task = client.call_tool(
                            "update_task",
                            {"id": task_id, "status": step},
                        )
                    task_count += 1
                    print(f"      ✓ {task['title']!r} (status={task['status']})")

        print(
            f"\n✓ Seed complete — {project_count} projects, {doc_count} documents, "
            f"{story_count} user stories, {task_count} tasks"
        )
        return 0
    except Exception as e:
        print(f"\n✗ seeder error after {project_count}p/{doc_count}d/{story_count}s/{task_count}t: {e}", file=sys.stderr)
        return 1
    finally:
        client.close()


if __name__ == "__main__":
    sys.exit(main())
