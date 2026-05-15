#!/usr/bin/env python3
"""
sync-gemini.py — regenerate .gemini/agents and .gemini/commands from .claude/.

Canonical source is .claude/. Edit there; run this script to update Gemini.
Skills and evals are NOT touched (they're symlinked:
  .gemini/skills -> ../.claude/skills
  .gemini/evals  -> ../.claude/evals).

The script normalizes Claude-specific names (tool names in frontmatter,
runner-specific phrases in agent / command bodies) into Gemini-native equivalents.

Usage:
    python3 scripts/sync-gemini.py
"""
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
CLAUDE_AGENTS = ROOT / ".claude" / "agents"
CLAUDE_COMMANDS = ROOT / ".claude" / "commands"
GEMINI_AGENTS = ROOT / ".gemini" / "agents"
GEMINI_COMMANDS = ROOT / ".gemini" / "commands"

GEMINI_MODEL = "gemini-3.1-pro-preview"

FRONTMATTER_RE = re.compile(r"^---\s*\n(.*?)\n---\s*\n", re.DOTALL)

# Claude tool name -> Gemini CLI tool name. Verified against geminicli.com docs.
# Claude tools with no Gemini equivalent are dropped (None).
CLAUDE_TO_GEMINI_TOOL = {
    "Read": "read_file",
    "Write": "write_file",
    "Edit": "replace",
    "Glob": "glob",
    "Grep": "search_file_content",
    "Bash": "run_shell_command",
    "WebFetch": "web_fetch",
    "WebSearch": "google_web_search",
    "AskUserQuestion": None,  # no equivalent — Gemini agents ask via natural language
    "NotebookEdit": None,
    "ExitPlanMode": None,
    "TaskCreate": None,
    "TaskUpdate": None,
    "TaskList": None,
    "TaskGet": None,
    "TaskOutput": None,
    "TaskStop": None,
}

# Substitutions applied to agent and command bodies so they read naturally
# under Gemini CLI. Applied in order; first match wins per occurrence.
# Strategy: drop redundant tool-name references where the surrounding prose
# already conveys the intent; only fall back to a textual replacement when
# the reference stands alone.
BODY_SUBSTITUTIONS = [
    # AskUserQuestion: surrounding text usually already says "ask the user" /
    # "confirm with the user", so the tool-name reference adds nothing.
    (r"\s+via\s+`AskUserQuestion`\s+or\s+similar", ""),
    (r"\s+via\s+`AskUserQuestion`", ""),
    (r"\s+through\s+`AskUserQuestion`", ""),
    (r"`AskUserQuestion`", "asking the user"),
    (r"\bAskUserQuestion\b", "asking the user"),
    # Subagent-spawning language. Claude Code uses the Agent tool with
    # subagent_type; Gemini CLI invokes subagents via auto-routing or @name.
    (r"two `Agent` tool calls", "two subagent invocations"),
    (r"parallel `Agent` tool calls", "parallel subagent invocations"),
    (r"parallel `Agent` calls", "parallel subagent invocations"),
    (r"a single `Agent` tool call", "a single subagent invocation"),
    (r"the `Agent` tool", "the subagent-invocation mechanism"),
    (r"\bAgent tool\b", "subagent-invocation mechanism"),
    (r"`Agent` call", "subagent invocation"),
    (r"`subagent_type`", "subagent name"),
    (r"\bsubagent_type\b", "subagent name"),
    # Worktree isolation. `isolation: "worktree"` is a Claude Code Agent-tool
    # parameter — the harness manages the worktree lifecycle automatically.
    # Gemini CLI has no equivalent, so the orchestrator runs `git worktree add`
    # by hand. More specific patterns first; the generic fallback catches the rest.
    (
        r'MUST include `isolation: "worktree"` so the harness creates a fresh worktree on a new branch off the current HEAD',
        "MUST run inside a fresh git worktree on a new branch off the current HEAD; "
        "the orchestrator creates one per spawn with `git worktree add -b agent/<short-id> "
        ".worktrees/<short-id>` and removes it after the merge",
    ),
    (
        r'MUST include `isolation: "worktree"`',
        "MUST run inside a fresh git worktree managed by the orchestrator "
        "(`git worktree add -b agent/<short-id> .worktrees/<short-id>`)",
    ),
    (
        r'`isolation: "worktree"`',
        "manual git worktree management",
    ),
]


def normalize_body(body: str) -> str:
    for pattern, replacement in BODY_SUBSTITUTIONS:
        body = re.sub(pattern, replacement, body)
    return body


def parse_frontmatter(text: str):
    """Return (fields_dict, body_after_frontmatter)."""
    m = FRONTMATTER_RE.match(text)
    if not m:
        raise ValueError("No YAML frontmatter found")
    raw = m.group(1)
    body = text[m.end():]
    fields = {}
    current_key = None
    for line in raw.splitlines():
        if not line.strip():
            continue
        if line[0] in (" ", "\t") and current_key:
            fields[current_key] += "\n" + line.strip()
            continue
        if ":" in line:
            key, _, value = line.partition(":")
            fields[key.strip()] = value.strip()
            current_key = key.strip()
    return fields, body


def map_tools(claude_tools_field: str) -> list:
    """Convert 'Read, Write, Edit, ...' to Gemini-native names; drop unmapped."""
    if not claude_tools_field:
        return []
    raw_names = [t.strip() for t in claude_tools_field.split(",") if t.strip()]
    mapped = []
    seen = set()
    for name in raw_names:
        if name not in CLAUDE_TO_GEMINI_TOOL:
            # Unknown tool — pass through verbatim so the user can audit.
            if name not in seen:
                mapped.append(name)
                seen.add(name)
            continue
        gemini = CLAUDE_TO_GEMINI_TOOL[name]
        if gemini is None or gemini in seen:
            continue
        mapped.append(gemini)
        seen.add(gemini)
    return mapped


def render_yaml_tools(tools: list) -> str:
    if not tools:
        return ""
    lines = ["tools:"]
    for t in tools:
        lines.append(f"  - {t}")
    return "\n".join(lines) + "\n"


def convert_agent(claude_path: Path, gemini_path: Path):
    text = claude_path.read_text()
    fields, body = parse_frontmatter(text)
    name = fields.get("name", claude_path.stem)
    description = fields.get("description", "")
    tools = map_tools(fields.get("tools", ""))
    tools_yaml = render_yaml_tools(tools)
    new_frontmatter = (
        "---\n"
        f"name: {name}\n"
        f"description: {description}\n"
        f"model: {GEMINI_MODEL}\n"
        f"{tools_yaml}"
        "---\n"
    )
    gemini_path.parent.mkdir(parents=True, exist_ok=True)
    gemini_path.write_text(new_frontmatter + normalize_body(body))
    dropped = [
        t.strip() for t in fields.get("tools", "").split(",")
        if t.strip() and CLAUDE_TO_GEMINI_TOOL.get(t.strip(), False) is None
    ]
    suffix = f" (dropped: {', '.join(dropped)})" if dropped else ""
    print(f"  agent  → {gemini_path.relative_to(ROOT)}{suffix}")


def convert_command(claude_path: Path, gemini_path: Path):
    text = claude_path.read_text()
    fields, body = parse_frontmatter(text)
    description = fields.get("description", "")
    body = body.replace("$ARGUMENTS", "{{args}}")
    body = normalize_body(body)
    # Escape any literal triple-quote runs in the body (rare).
    body = body.replace('"""', '""\\"')
    description_escaped = description.replace("\\", "\\\\").replace('"', '\\"')
    toml = (
        f'description = "{description_escaped}"\n'
        'prompt = """\n'
        f"{body}"
        '"""\n'
    )
    gemini_path.parent.mkdir(parents=True, exist_ok=True)
    gemini_path.write_text(toml)
    print(f"  command → {gemini_path.relative_to(ROOT)}")


def main():
    if not CLAUDE_AGENTS.is_dir() or not CLAUDE_COMMANDS.is_dir():
        print("ERR: .claude/agents or .claude/commands missing", file=sys.stderr)
        sys.exit(1)

    print("Syncing Gemini config from Claude canonical source...")
    print(f"  model = {GEMINI_MODEL}")

    for src in sorted(CLAUDE_AGENTS.glob("*.md")):
        dst = GEMINI_AGENTS / src.name
        convert_agent(src, dst)

    for src in sorted(CLAUDE_COMMANDS.glob("*.md")):
        dst = GEMINI_COMMANDS / (src.stem + ".toml")
        convert_command(src, dst)

    print("Done. .gemini/skills/ and .gemini/evals/ are symlinks — no sync needed.")


if __name__ == "__main__":
    main()
