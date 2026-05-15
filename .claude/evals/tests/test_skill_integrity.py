"""Structural tests for vibe-commerce skills under .claude/skills/.

Mirrors claude-skills/tests/test_skill_integrity.py, scoped to this project.
Recurses to handle plugin-bundled skills (skills/<name>/skills/<name>/SKILL.md).
"""

import glob
import os
import re

import pytest

from conftest import CLAUDE_SKILLS_DIR, REPO_ROOT


def _short(path):
    return os.path.relpath(path, REPO_ROOT)


def _find_all_skill_dirs():
    """Every directory containing a SKILL.md anywhere under .claude/skills/."""
    if not os.path.isdir(CLAUDE_SKILLS_DIR):
        return []
    matches = []
    for root, _dirs, files in os.walk(CLAUDE_SKILLS_DIR):
        if "SKILL.md" in files:
            matches.append(root)
    return sorted(matches)


ALL_SKILL_DIRS = _find_all_skill_dirs()


def test_skills_directory_exists():
    assert os.path.isdir(CLAUDE_SKILLS_DIR), f"Missing {_short(CLAUDE_SKILLS_DIR)}"


def test_at_least_one_skill_found():
    assert ALL_SKILL_DIRS, "No SKILL.md found anywhere under .claude/skills/"


@pytest.mark.parametrize(
    "skill_dir", ALL_SKILL_DIRS, ids=[_short(s) for s in ALL_SKILL_DIRS]
)
def test_has_frontmatter(skill_dir):
    skill_md = os.path.join(skill_dir, "SKILL.md")
    with open(skill_md, "r", encoding="utf-8") as f:
        content = f.read()
    assert content.startswith("---"), f"{_short(skill_md)} missing opening ---"
    closing = content.find("---", 4)
    assert closing > 0, f"{_short(skill_md)} has unclosed frontmatter"


@pytest.mark.parametrize(
    "skill_dir", ALL_SKILL_DIRS, ids=[_short(s) for s in ALL_SKILL_DIRS]
)
def test_frontmatter_has_name_and_description(skill_dir):
    skill_md = os.path.join(skill_dir, "SKILL.md")
    with open(skill_md, "r", encoding="utf-8") as f:
        content = f.read()
    m = re.match(r"^---\s*\n(.*?)\n---\s*\n", content, re.DOTALL)
    assert m, f"{_short(skill_md)} frontmatter not parseable"
    fm = m.group(1)
    assert re.search(r"^name\s*:", fm, re.MULTILINE), (
        f"{_short(skill_md)} frontmatter missing `name:`"
    )
    assert re.search(r"^description\s*:", fm, re.MULTILINE), (
        f"{_short(skill_md)} frontmatter missing `description:`"
    )


@pytest.mark.parametrize(
    "skill_dir", ALL_SKILL_DIRS, ids=[_short(s) for s in ALL_SKILL_DIRS]
)
def test_frontmatter_name_matches_folder(skill_dir):
    skill_md = os.path.join(skill_dir, "SKILL.md")
    with open(skill_md, "r", encoding="utf-8") as f:
        content = f.read()
    m = re.match(r"^---\s*\n(.*?)\n---\s*\n", content, re.DOTALL)
    fm = m.group(1) if m else ""
    name_match = re.search(r'^name\s*:\s*"?([^"\n]+?)"?\s*$', fm, re.MULTILINE)
    assert name_match, f"{_short(skill_md)} couldn't parse name field"
    declared_name = name_match.group(1).strip()
    folder_name = os.path.basename(skill_dir)
    assert declared_name == folder_name, (
        f"{_short(skill_md)} declares name '{declared_name}' "
        f"but lives in folder '{folder_name}'"
    )


@pytest.mark.parametrize(
    "skill_dir", ALL_SKILL_DIRS, ids=[_short(s) for s in ALL_SKILL_DIRS]
)
def test_has_h1(skill_dir):
    skill_md = os.path.join(skill_dir, "SKILL.md")
    with open(skill_md, "r", encoding="utf-8") as f:
        content = f.read()
    body = re.sub(r"^---\s*\n.*?\n---\s*\n", "", content, count=1, flags=re.DOTALL)
    assert re.search(r"^#\s+.+", body, re.MULTILINE), (
        f"{_short(skill_md)} has no H1 heading"
    )


@pytest.mark.parametrize(
    "skill_dir", ALL_SKILL_DIRS, ids=[_short(s) for s in ALL_SKILL_DIRS]
)
def test_skill_md_not_empty(skill_dir):
    skill_md = os.path.join(skill_dir, "SKILL.md")
    size = os.path.getsize(skill_md)
    assert size > 100, f"{_short(skill_md)} suspiciously small ({size} bytes)"


@pytest.mark.parametrize(
    "skill_dir", ALL_SKILL_DIRS, ids=[_short(s) for s in ALL_SKILL_DIRS]
)
def test_references_dir_not_empty_if_present(skill_dir):
    refs = os.path.join(skill_dir, "references")
    if os.path.isdir(refs):
        files = [f for f in os.listdir(refs) if not f.startswith(".")]
        assert files, f"{_short(refs)} exists but is empty"


@pytest.mark.parametrize(
    "skill_dir", ALL_SKILL_DIRS, ids=[_short(s) for s in ALL_SKILL_DIRS]
)
def test_scripts_dir_has_scripts_if_present(skill_dir):
    scripts = os.path.join(skill_dir, "scripts")
    if not os.path.isdir(scripts):
        return
    patterns = ("*.py", "*.mjs", "*.js", "*.ts", "*.sh", "*.ps1")
    found = []
    for pat in patterns:
        found.extend(glob.glob(os.path.join(scripts, pat)))
    assert found, f"{_short(scripts)} exists but has no script files ({patterns})"


def test_no_duplicate_skill_names():
    seen = {}
    for d in ALL_SKILL_DIRS:
        name = os.path.basename(d)
        seen.setdefault(name, []).append(_short(d))
    dupes = {k: v for k, v in seen.items() if len(v) > 2}
    assert not dupes, f"Skill names appear 3+ times: {dupes}"
