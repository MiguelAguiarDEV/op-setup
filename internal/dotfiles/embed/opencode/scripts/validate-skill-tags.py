#!/usr/bin/env python3
"""Validate/report skill folder naming tags.

Policy:
- tagged skills use <tag>-<name>
- accepted tags: op, sec, fe, qa
- untagged/general skills are allowed unless --strict-untagged is enabled
"""

from __future__ import annotations

from pathlib import Path
import argparse
import sys

ALLOWED_TAGS = {"op", "sec", "fe", "qa"}
SKILLS_DIR = Path(__file__).resolve().parents[1] / "skills"


def main() -> int:
    parser = argparse.ArgumentParser(description="Validate/report skill naming tags")
    parser.add_argument(
        "--strict-untagged",
        action="store_true",
        help="Fail when skills without approved tag prefix are found.",
    )
    args = parser.parse_args()

    if not SKILLS_DIR.is_dir():
        print(f"skills directory not found: {SKILLS_DIR}", file=sys.stderr)
        return 1

    tagged = []
    unprefixed = []
    checked = 0

    for skill_file in sorted(SKILLS_DIR.glob('*/SKILL.md')):
        checked += 1
        folder = skill_file.parent.name
        prefix, sep, _rest = folder.partition("-")
        if not sep or prefix not in ALLOWED_TAGS:
            unprefixed.append(folder)
            continue
        tagged.append(folder)

    print(f"Checked {checked} skills")
    print(f"Tagged ({len(tagged)}):")
    for name in tagged:
        print(f"  - {name}")

    if unprefixed:
        print(f"Untagged/general ({len(unprefixed)}):")
        for name in unprefixed:
            print(f"  - {name}")

    if args.strict_untagged and unprefixed:
        print(
            "Error: untagged skills found with --strict-untagged. "
            "Rename or explicitly accept them.",
            file=sys.stderr,
        )
        return 1

    print("Skill tag validation passed")
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
