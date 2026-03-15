#!/usr/bin/env python3
"""Generate skills and agents tables inside AGENTS.md."""

from __future__ import annotations

import argparse
from pathlib import Path
import re
import sys

SKILLS_TABLE_START = "<!-- SKILLS_TABLE:START -->"
SKILLS_TABLE_END = "<!-- SKILLS_TABLE:END -->"
SKILLS_PLACEHOLDER = "Context | Read this file"

AGENTS_TABLE_START = "<!-- AGENTS_TABLE:START -->"
AGENTS_TABLE_END = "<!-- AGENTS_TABLE:END -->"
AGENTS_SECTION_TITLE = "# Agents (Auto-load based on context)"
AGENTS_SECTION_INTRO = "Use these subagents for specialized, read-only analysis workflows."
AGENTS_INSERT_BEFORE = "\n# How to use skills"
ACTIVE_SKILLS_FILE = "ACTIVE_SKILLS.txt"
ACTIVE_AGENTS_FILE = "ACTIVE_AGENTS.txt"


def parse_frontmatter(text: str) -> dict[str, str]:
    if not text.startswith("---\n"):
        return {}

    end = text.find("\n---\n", 4)
    if end == -1:
        return {}

    block = text[4:end]
    meta: dict[str, str] = {}
    for raw_line in block.splitlines():
        line = raw_line.rstrip()
        if not line or line.startswith((" ", "\t")) or ":" not in line:
            continue
        key, value = line.split(":", 1)
        meta[key.strip()] = value.strip()
    return meta


def escape_cell(value: str) -> str:
    return re.sub(r"\|", r"\\|", value.strip())


def build_table(headers: list[str], rows: list[list[str]]) -> str:
    lines = [
        "| " + " | ".join(headers) + " |",
        "| " + " | ".join(["---"] * len(headers)) + " |",
    ]
    for row in rows:
        lines.append("| " + " | ".join(row) + " |")
    return "\n".join(lines)


def load_active_entries(path: Path) -> set[str] | None:
    if not path.is_file():
        return None

    active: set[str] = set()
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        active.add(line)
    return active


def replace_marker_block(text: str, start: str, end: str, inner: str) -> str | None:
    if start not in text or end not in text:
        return None

    block = f"{start}\n{inner}\n{end}"
    pattern = re.compile(rf"{re.escape(start)}.*?{re.escape(end)}", re.DOTALL)
    return pattern.sub(block, text, count=1)


def collect_skills_rows(skills_dir: Path, repo_root: Path) -> list[list[str]]:
    active_entries = load_active_entries(skills_dir / ACTIVE_SKILLS_FILE)
    rows: list[list[str]] = []
    for skill_file in sorted(skills_dir.glob("*/SKILL.md")):
        skill_dir_name = skill_file.parent.name
        if active_entries is not None and skill_dir_name not in active_entries:
            continue

        meta = parse_frontmatter(skill_file.read_text(encoding="utf-8"))
        context = meta.get("name") or skill_file.parent.name
        description = meta.get("description", "")
        path = skill_file.relative_to(repo_root).as_posix()
        rows.append(
            [
                f"`{escape_cell(context)}`",
                f"`{escape_cell(path)}`",
                escape_cell(description),
            ]
        )

    rows.sort(key=lambda row: row[0].lower())

    if active_entries is not None:
        existing = {p.parent.name for p in skills_dir.glob("*/SKILL.md")}
        missing = sorted(active_entries - existing)
        for item in missing:
            print(
                f"warning: active skill '{item}' not found under skills/",
                file=sys.stderr,
            )
    return rows


def collect_agents_rows(agents_dir: Path, repo_root: Path) -> list[list[str]]:
    active_entries = load_active_entries(agents_dir / ACTIVE_AGENTS_FILE)
    rows: list[list[str]] = []
    for agent_file in sorted(agents_dir.glob("*.md")):
        agent_name_from_file = agent_file.stem
        if active_entries is not None and agent_name_from_file not in active_entries:
            continue

        meta = parse_frontmatter(agent_file.read_text(encoding="utf-8"))
        agent_name = meta.get("name") or agent_file.stem
        description = meta.get("description", "")
        path = agent_file.relative_to(repo_root).as_posix()
        rows.append(
            [
                f"`{escape_cell(agent_name)}`",
                f"`{escape_cell(path)}`",
                escape_cell(description),
            ]
        )

    rows.sort(key=lambda row: row[0].lower())

    if active_entries is not None:
        existing = {p.stem for p in agents_dir.glob("*.md")}
        missing = sorted(active_entries - existing)
        for item in missing:
            print(
                f"warning: active agent '{item}' not found under agents/",
                file=sys.stderr,
            )
    return rows


def inject_skills_table(text: str, table: str) -> str:
    replaced = replace_marker_block(text, SKILLS_TABLE_START, SKILLS_TABLE_END, table)
    if replaced is not None:
        return replaced

    if SKILLS_PLACEHOLDER in text:
        block = f"{SKILLS_TABLE_START}\n{table}\n{SKILLS_TABLE_END}"
        return text.replace(SKILLS_PLACEHOLDER, block, 1)

    raise ValueError(
        f"Could not find skills markers or placeholder ({SKILLS_PLACEHOLDER!r}) in AGENTS.md."
    )


def inject_agents_table(text: str, table: str) -> str:
    replaced = replace_marker_block(text, AGENTS_TABLE_START, AGENTS_TABLE_END, table)
    if replaced is not None:
        return replaced

    section = "\n".join(
        [
            AGENTS_SECTION_TITLE,
            "",
            AGENTS_SECTION_INTRO,
            "",
            AGENTS_TABLE_START,
            table,
            AGENTS_TABLE_END,
            "",
        ]
    )

    anchor_index = text.find(AGENTS_INSERT_BEFORE)
    if anchor_index == -1:
        raise ValueError(
            f"Could not find insertion anchor ({AGENTS_INSERT_BEFORE.strip()!r}) in AGENTS.md."
        )

    return text[:anchor_index] + "\n" + section + text[anchor_index:]


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Generate skills and agents tables in AGENTS.md."
    )
    parser.add_argument(
        "--repo-root",
        default=".",
        help="Repository root that contains AGENTS.md, skills/, and agents/.",
    )
    args = parser.parse_args()

    repo_root = Path(args.repo_root).resolve()
    skills_dir = repo_root / "skills"
    agents_dir = repo_root / "agents"
    agents_md = repo_root / "AGENTS.md"

    if not skills_dir.is_dir():
        print(f"skills directory not found: {skills_dir}", file=sys.stderr)
        return 1
    if not agents_dir.is_dir():
        print(f"agents directory not found: {agents_dir}", file=sys.stderr)
        return 1
    if not agents_md.is_file():
        print(f"AGENTS.md not found: {agents_md}", file=sys.stderr)
        return 1

    skills_rows = collect_skills_rows(skills_dir, repo_root)
    agents_rows = collect_agents_rows(agents_dir, repo_root)

    skills_table = build_table(
        headers=["Context", "Read this file", "Description"],
        rows=skills_rows,
    )
    agents_table = build_table(
        headers=["Agent", "Read this file", "Description"],
        rows=agents_rows,
    )

    original = agents_md.read_text(encoding="utf-8")
    with_skills = inject_skills_table(original, skills_table)
    with_agents = inject_agents_table(with_skills, agents_table)
    agents_md.write_text(with_agents, encoding="utf-8")

    print(
        f"Updated {agents_md.relative_to(repo_root)} with "
        f"{len(skills_rows)} skills and {len(agents_rows)} agents."
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
