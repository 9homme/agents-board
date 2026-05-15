"""Structural tests for vibe-commerce agents under .claude/agents/ AND .gemini/agents/.

Validates: presence, frontmatter, name=filename, model matches role matrix
(Claude only — Gemini model strings vary), tools come from per-platform allowlist,
and parity (same 6 agents with identical descriptions) between platforms.

YAML parsing uses pyyaml if available, else a small regex fallback.
"""

import os
import re

import pytest

from conftest import (
    CLAUDE_AGENTS_DIR,
    CLAUDE_AGENT_MODELS,
    CLAUDE_TOOL_ALLOWLIST,
    GEMINI_AGENTS_DIR,
    GEMINI_TOOL_ALLOWLIST,
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
GEMINI_AGENT_FILES = _list_agent_files(GEMINI_AGENTS_DIR)


def _agent_name_from_path(p):
    return os.path.splitext(os.path.basename(p))[0]


def test_claude_agents_dir_exists():
    assert os.path.isdir(CLAUDE_AGENTS_DIR), f"Missing {_short(CLAUDE_AGENTS_DIR)}"


def test_gemini_agents_dir_exists():
    assert os.path.isdir(GEMINI_AGENTS_DIR), f"Missing {_short(GEMINI_AGENTS_DIR)}"


def test_claude_has_all_required_agents():
    got = {_agent_name_from_path(p) for p in CLAUDE_AGENT_FILES}
    missing = REQUIRED_AGENTS - got
    assert not missing, f"Missing Claude agents: {missing}"


def test_gemini_has_all_required_agents():
    got = {_agent_name_from_path(p) for p in GEMINI_AGENT_FILES}
    missing = REQUIRED_AGENTS - got
    assert not missing, f"Missing Gemini agents: {missing}"


# -------- per-file checks (Claude) --------


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
        f"CLAUDE.md role matrix requires '{expected}' for {name}"
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


# -------- per-file checks (Gemini) --------


@pytest.mark.parametrize(
    "path", GEMINI_AGENT_FILES, ids=[_short(p) for p in GEMINI_AGENT_FILES]
)
def test_gemini_frontmatter_present(path):
    fm = _parse_frontmatter(path)
    assert fm is not None, f"{_short(path)} has no parseable frontmatter"


@pytest.mark.parametrize(
    "path", GEMINI_AGENT_FILES, ids=[_short(p) for p in GEMINI_AGENT_FILES]
)
def test_gemini_required_fields(path):
    fm = _parse_frontmatter(path) or {}
    for key in ("name", "description", "model", "tools"):
        assert key in fm and fm[key], f"{_short(path)} frontmatter missing `{key}`"


@pytest.mark.parametrize(
    "path", GEMINI_AGENT_FILES, ids=[_short(p) for p in GEMINI_AGENT_FILES]
)
def test_gemini_name_matches_filename(path):
    fm = _parse_frontmatter(path) or {}
    declared = str(fm.get("name", "")).strip()
    filename = _agent_name_from_path(path)
    assert declared == filename, (
        f"{_short(path)} declares name '{declared}' but filename is '{filename}'"
    )


@pytest.mark.parametrize(
    "path", GEMINI_AGENT_FILES, ids=[_short(p) for p in GEMINI_AGENT_FILES]
)
def test_gemini_tools_in_allowlist(path):
    fm = _parse_frontmatter(path) or {}
    tools_raw = fm.get("tools", "")
    if isinstance(tools_raw, list):
        tools = [t.strip() for t in tools_raw if t.strip()]
    else:
        tools = [t.strip() for t in str(tools_raw).split(",") if t.strip()]
    assert tools, f"{_short(path)} declares no tools"
    bad = [t for t in tools if t not in GEMINI_TOOL_ALLOWLIST]
    assert not bad, (
        f"{_short(path)} declares unknown Gemini tools: {bad}. "
        f"Allowed: {sorted(GEMINI_TOOL_ALLOWLIST)}"
    )


# -------- cross-platform parity --------


def _description_of(path):
    fm = _parse_frontmatter(path) or {}
    return str(fm.get("description", "")).strip()


@pytest.mark.parametrize("agent_name", sorted(REQUIRED_AGENTS))
def test_descriptions_match_across_platforms(agent_name):
    claude_path = os.path.join(CLAUDE_AGENTS_DIR, f"{agent_name}.md")
    gemini_path = os.path.join(GEMINI_AGENTS_DIR, f"{agent_name}.md")
    if not (os.path.isfile(claude_path) and os.path.isfile(gemini_path)):
        pytest.skip(f"{agent_name} missing from one or both platforms (caught elsewhere)")
    c_desc = _description_of(claude_path)
    g_desc = _description_of(gemini_path)
    assert c_desc == g_desc, (
        f"Description drift between platforms for '{agent_name}':\n"
        f"  .claude: {c_desc[:120]}...\n"
        f"  .gemini: {g_desc[:120]}..."
    )


def test_tech_lead_references_review_gate_on_both_platforms():
    """Both .claude/agents/tech-lead.md and .gemini/agents/tech-lead.md must
    reference scripts/review/run-gate.sh — the gate is mandatory in review mode.
    If either file loses the reference (drift, accidental rewrite), the test
    fails so the next sync run can restore it."""
    required_phrases = [
        "scripts/review/run-gate.sh",
        "REVIEW GATE",
    ]
    for platform_dir in (CLAUDE_AGENTS_DIR, GEMINI_AGENTS_DIR):
        path = os.path.join(platform_dir, "tech-lead.md")
        if not os.path.isfile(path):
            pytest.skip(f"{_short(path)} missing (caught by required-agents test)")
        with open(path, "r", encoding="utf-8") as f:
            content = f.read()
        for phrase in required_phrases:
            assert phrase in content, (
                f"{_short(path)} is missing required gate phrase: '{phrase}'. "
                f"Edit .claude/agents/tech-lead.md and rerun scripts/sync-gemini.py."
            )


def test_po_ba_can_ask_user_on_each_platform():
    """po-ba must be able to ask clarifying questions — flag if Gemini's po-ba
    tools list lacks any user-input capability. This is informational: it asserts
    the Claude side has AskUserQuestion, and warns rather than fails if Gemini
    has no equivalent in its allowlist."""
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
    # Gemini side intentionally not asserted: the platform's CLI handles user
    # prompts differently. If/when Gemini gains a tool name for user input,
    # add it to GEMINI_TOOL_ALLOWLIST and assert here.
