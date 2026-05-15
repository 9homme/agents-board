"""Shared helpers for eval structural tests.

REPO_ROOT resolves up from .claude/evals/tests/ to the project root, so the tests
work whether invoked from the repo root or from inside the evals folder.
"""

import os

REPO_ROOT = os.path.abspath(
    os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", "..")
)

CLAUDE_AGENTS_DIR = os.path.join(REPO_ROOT, ".claude", "agents")
GEMINI_AGENTS_DIR = os.path.join(REPO_ROOT, ".gemini", "agents")
CLAUDE_SKILLS_DIR = os.path.join(REPO_ROOT, ".claude", "skills")

REQUIRED_AGENTS = {"po-ba", "system-architect", "tech-lead", "tester", "be-dev", "fe-dev"}

# Per CLAUDE.md role matrix.
CLAUDE_AGENT_MODELS = {
    "po-ba": "opus",
    "system-architect": "opus",
    "tech-lead": "opus",
    "tester": "sonnet",
    "be-dev": "sonnet",
    "fe-dev": "sonnet",
}

CLAUDE_TOOL_ALLOWLIST = {
    "Read", "Write", "Edit", "Glob", "Grep", "Bash",
    "AskUserQuestion", "WebFetch", "WebSearch",
}

GEMINI_TOOL_ALLOWLIST = {
    "read_file", "write_file", "replace", "glob",
    "search_file_content", "run_shell_command", "google_web_search",
}
