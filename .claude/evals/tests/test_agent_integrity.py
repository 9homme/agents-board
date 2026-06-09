"""Structural tests for a-team agents under .claude/agents/.

Validates: presence, frontmatter, name=filename, model matches role matrix,
tools come from the allowlist.

YAML parsing uses pyyaml if available, else a small regex fallback.
"""

import os
import re

import pytest

from conftest import (
    CLAUDE_AGENTS_DIR,
    CLAUDE_AGENT_MODELS,
    CLAUDE_TOOL_ALLOWLIST,
    REPO_ROOT,
    REQUIRED_AGENTS,
)


try:
    import yaml
    _HAS_YAML = True
except ImportError:
    _HAS_YAML = False


def _short(path):
    return os.path.relpath(path, REPO_ROOT)


def _list_agent_files(agents_dir):
    if not os.path.isdir(agents_dir):
        return []
    return sorted(
        os.path.join(agents_dir, f)
        for f in os.listdir(agents_dir)
        if f.endswith(".md") and not f.startswith(".")
    )


def _parse_frontmatter(path):
    with open(path, "r", encoding="utf-8") as f:
        content = f.read()
    m = re.match(r"^---\s*\n(.*?)\n---\s*\n", content, re.DOTALL)
    if not m:
        return None
    raw = m.group(1)
    if _HAS_YAML:
        try:
            return yaml.safe_load(raw) or {}
        except yaml.YAMLError:
            return None
    # Regex fallback: handles `key: value` and `tools: a, b` or block list.
    out = {}
    current_list_key = None
    for line in raw.splitlines():
        if not line.strip():
            current_list_key = None
            continue
        if current_list_key and line.lstrip().startswith("- "):
            out.setdefault(current_list_key, []).append(line.lstrip()[2:].strip())
            continue
        m2 = re.match(r"^(\w[\w-]*)\s*:\s*(.*)$", line)
        if m2:
            key, val = m2.group(1), m2.group(2).strip()
            if val == "":
                current_list_key = key
            else:
                current_list_key = None
                out[key] = val
    return out


CLAUDE_AGENT_FILES = _list_agent_files(CLAUDE_AGENTS_DIR)


def _agent_name_from_path(p):
    return os.path.splitext(os.path.basename(p))[0]


# -------- directory and presence checks --------


def test_claude_agents_dir_exists():
    assert os.path.isdir(CLAUDE_AGENTS_DIR), f"Missing {_short(CLAUDE_AGENTS_DIR)}"


def test_claude_has_all_required_agents():
    got = {_agent_name_from_path(p) for p in CLAUDE_AGENT_FILES}
    missing = REQUIRED_AGENTS - got
    assert not missing, f"Missing Claude agents: {missing}"


# -------- per-file checks --------


@pytest.mark.parametrize(
    "path", CLAUDE_AGENT_FILES, ids=[_short(p) for p in CLAUDE_AGENT_FILES]
)
def test_claude_frontmatter_present(path):
    fm = _parse_frontmatter(path)
    assert fm is not None, f"{_short(path)} has no parseable frontmatter"


@pytest.mark.parametrize(
    "path", CLAUDE_AGENT_FILES, ids=[_short(p) for p in CLAUDE_AGENT_FILES]
)
def test_claude_required_fields(path):
    fm = _parse_frontmatter(path) or {}
    for key in ("name", "description", "model", "tools"):
        assert key in fm and fm[key], f"{_short(path)} frontmatter missing `{key}`"


@pytest.mark.parametrize(
    "path", CLAUDE_AGENT_FILES, ids=[_short(p) for p in CLAUDE_AGENT_FILES]
)
def test_claude_name_matches_filename(path):
    fm = _parse_frontmatter(path) or {}
    declared = str(fm.get("name", "")).strip()
    filename = _agent_name_from_path(path)
    assert declared == filename, (
        f"{_short(path)} declares name '{declared}' but filename is '{filename}'"
    )


@pytest.mark.parametrize(
    "path", CLAUDE_AGENT_FILES, ids=[_short(p) for p in CLAUDE_AGENT_FILES]
)
def test_claude_model_matches_role_matrix(path):
    name = _agent_name_from_path(path)
    if name not in CLAUDE_AGENT_MODELS:
        pytest.skip(f"Unknown agent '{name}' — not in role matrix")
    fm = _parse_frontmatter(path) or {}
    model = str(fm.get("model", "")).strip()
    expected = CLAUDE_AGENT_MODELS[name]
    assert model == expected, (
        f"{_short(path)} declares model '{model}', "
        f"role matrix requires '{expected}' for {name}"
    )


@pytest.mark.parametrize(
    "path", CLAUDE_AGENT_FILES, ids=[_short(p) for p in CLAUDE_AGENT_FILES]
)
def test_claude_tools_in_allowlist(path):
    fm = _parse_frontmatter(path) or {}
    tools_raw = fm.get("tools", "")
    if isinstance(tools_raw, list):
        tools = [t.strip() for t in tools_raw if t.strip()]
    else:
        tools = [t.strip() for t in str(tools_raw).split(",") if t.strip()]
    assert tools, f"{_short(path)} declares no tools"
    bad = [t for t in tools if t not in CLAUDE_TOOL_ALLOWLIST]
    assert not bad, (
        f"{_short(path)} declares unknown Claude tools: {bad}. "
        f"Allowed: {sorted(CLAUDE_TOOL_ALLOWLIST)}"
    )


# -------- structural invariants --------


def test_worktree_isolation_required_in_phase3_command():
    """The phase3 orchestrator command must enforce worktree isolation.
    Claude side uses the Agent tool's `isolation: "worktree"`. If this goes
    missing, parallel agents will race on the working tree."""
    claude_phase3 = os.path.join(REPO_ROOT, ".claude", "commands", "phase3.md")

    with open(claude_phase3, "r", encoding="utf-8") as f:
        claude_text = f.read()
    assert 'isolation: "worktree"' in claude_text, (
        ".claude/commands/phase3.md must reference `isolation: \"worktree\"` "
        "(the Claude Code Agent-tool parameter for worktree isolation)."
    )
    assert "Files touched" in claude_text, (
        ".claude/commands/phase3.md must reference the `Files touched` field "
        "(used by the orchestrator to avoid co-picking overlapping tasks)."
    )


def test_task_template_has_files_touched_section():
    """The shared task template must include a `## Files touched` section so
    every task declares its file scope — the orchestrator depends on this to
    avoid co-picking overlapping tasks. tech-lead-planner authors tasks from
    this template."""
    path = os.path.join(REPO_ROOT, ".claude", "refs", "task-template.md")
    if not os.path.isfile(path):
        pytest.skip(f"{_short(path)} missing (caught elsewhere)")
    with open(path, "r", encoding="utf-8") as f:
        content = f.read()
    assert "## Files touched" in content, (
        f"{_short(path)} task template must include a `## Files touched` section."
    )
    assert "## Architecture extract" in content, (
        f"{_short(path)} task template must include a `## Architecture extract` "
        "section so devs implement without opening architecture.md."
    )


def test_planner_references_task_template():
    """tech-lead-planner must point at the shared task template."""
    path = os.path.join(CLAUDE_AGENTS_DIR, "tech-lead-planner.md")
    if not os.path.isfile(path):
        pytest.skip(f"{_short(path)} missing (caught by required-agents test)")
    with open(path, "r", encoding="utf-8") as f:
        content = f.read()
    assert ".claude/refs/task-template.md" in content, (
        f"{_short(path)} must reference .claude/refs/task-template.md."
    )


def test_devs_have_worktree_awareness():
    """be-dev and fe-dev prompts must mention worktree isolation and commit
    instructions. If a dev doesn't commit, the harness cleans up the worktree
    and the work is lost."""
    for agent_name in ("be-dev", "fe-dev"):
        path = os.path.join(CLAUDE_AGENTS_DIR, f"{agent_name}.md")
        if not os.path.isfile(path):
            pytest.skip(f"{_short(path)} missing (caught elsewhere)")
        with open(path, "r", encoding="utf-8") as f:
            content = f.read()
        assert "Worktree isolation" in content or "worktree isolation" in content, (
            f"{_short(path)} must explain worktree isolation to the dev."
        )
        assert "git commit" in content or "Commit on the worktree branch" in content, (
            f"{_short(path)} must instruct the dev to commit before exiting."
        )


def test_tech_lead_reviewer_references_review_gate():
    """tech-lead-reviewer.md must reference scripts/review/run-gate.sh — the
    REQ-level quality gate (Mode 2) is mandatory."""
    required_phrases = [
        "scripts/review/run-gate.sh",
        "REVIEW GATE",
    ]
    path = os.path.join(CLAUDE_AGENTS_DIR, "tech-lead-reviewer.md")
    if not os.path.isfile(path):
        pytest.skip(f"{_short(path)} missing (caught by required-agents test)")
    with open(path, "r", encoding="utf-8") as f:
        content = f.read()
    for phrase in required_phrases:
        assert phrase in content, (
            f"{_short(path)} is missing required gate phrase: '{phrase}'."
        )


def test_po_ba_can_ask_user():
    """po-ba must be able to ask clarifying questions via AskUserQuestion."""
    claude_path = os.path.join(CLAUDE_AGENTS_DIR, "po-ba.md")
    if os.path.isfile(claude_path):
        fm = _parse_frontmatter(claude_path) or {}
        tools_raw = fm.get("tools", "")
        tools = (
            [t.strip() for t in tools_raw] if isinstance(tools_raw, list)
            else [t.strip() for t in str(tools_raw).split(",") if t.strip()]
        )
        assert "AskUserQuestion" in tools, (
            ".claude/agents/po-ba.md must declare AskUserQuestion — "
            "po-ba's job is to ask clarifying questions before writing stories."
        )
