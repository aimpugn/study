#!/usr/bin/env python3
import re
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
HEADING_RE = re.compile(r"^(#{2,3})\s+(.+?)\s*$")
TOC_LINE_RE = re.compile(r"^\s*- \[[^\]]+\]\(#[^)]+\)\s*$")


def slug(title: str) -> str:
    value = title.strip().lower()
    value = re.sub(r"`([^`]*)`", r"\1", value)
    value = re.sub(r"<[^>]+>", "", value)
    value = re.sub(r"[^\w가-힣ㄱ-ㅎㅏ-ㅣ\s-]", "", value)
    value = re.sub(r"\s+", "-", value.strip())
    return value


def headings(lines: list[str]) -> list[tuple[int, str]]:
    result: list[tuple[int, str]] = []
    for line in lines:
        match = HEADING_RE.match(line)
        if match:
            result.append((len(match.group(1)), match.group(2)))
    return result


def toc_lines(lines: list[str]) -> list[str]:
    rendered: list[str] = []
    for level, title in headings(lines):
        indent = "" if level == 2 else "    "
        rendered.append(f"{indent}- [{title}](#{slug(title)})")
    return rendered


def find_existing_toc(lines: list[str]) -> tuple[int, int]:
    first_h2 = next((i for i, line in enumerate(lines) if line.startswith("## ")), None)
    if first_h2 is None:
        raise ValueError("document has no H2 heading")

    candidates = [i for i in range(first_h2) if TOC_LINE_RE.match(lines[i])]
    if not candidates:
        return first_h2, first_h2

    start = candidates[0]
    end = candidates[-1] + 1
    while start > 0 and lines[start - 1].strip() == "":
        start -= 1
    while end < len(lines) and lines[end].strip() == "":
        end += 1
        break
    return start, end


def update_file(path: Path) -> bool:
    original = path.read_text(encoding="utf-8")
    lines = original.splitlines()
    start, end = find_existing_toc(lines)
    replacement = toc_lines(lines)
    new_lines = lines[:start] + replacement + [""] + lines[end:]
    updated = "\n".join(new_lines).rstrip() + "\n"
    if updated == original:
        return False
    path.write_text(updated, encoding="utf-8")
    return True


def main() -> int:
    changed: list[str] = []
    for path in sorted(ROOT.glob("[0-9][0-9]-*.md")):
        if update_file(path):
            changed.append(path.relative_to(ROOT).as_posix())
    if changed:
        print("updated TOC:")
        for path in changed:
            print(f"- {path}")
    else:
        print("TOC already up to date")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
