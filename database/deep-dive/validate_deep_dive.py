#!/usr/bin/env python3
from __future__ import annotations

import argparse
import csv
import re
import sys
from dataclasses import dataclass
from pathlib import Path


ROOT = Path(__file__).resolve().parent
REPO = ROOT.parents[1]
REGISTRY = ROOT / "du-registry.tsv"
SOURCE_MAP = ROOT / "source-map.md"
SOURCE_MAP_TSV = ROOT / "source-map.tsv"
ACTIVE_PLAN = REPO / "database" / "database-deep-study-plan.md"
ACTIVE_WORK = REPO / "docs" / "works" / "WORK_20260519_DATABASE_DEEP_STUDY_SYSTEM_PLAN.md"

SOURCE_REQUIRED_COLUMNS = {
    "id",
    "local_seed",
    "official_sources",
    "lab_or_observability",
    "preservation_disposition",
    "source_status",
    "notes",
}

ALLOWED_DISPOSITIONS = {
    "preserve",
    "merge",
    "expand",
    "supersede-with-pointer",
    "archive-candidate",
    "sensitive-source-do-not-promote",
    "raw-source-sanitize-first",
    "new-source",
}

PLACEHOLDER_VALUES = {"", "-", "todo", "tbd", "n/a", "na", "none", "null", "미정"}

FORBIDDEN_PROSE_PATTERNS: tuple[tuple[str, str], ...] = (
    ("old repeated verification suffix", r"PASS when observed signal matches the documented state transition"),
    ("old answer verification boilerplate", r"이 답변의 검증은 말로 끝나지 않는다"),
    ("old generic evidence list boilerplate", r"로그, SQL, 상태 row, 테스트 fixture"),
    ("old final-result-only boilerplate", r"최종 결과만 맞고"),
    ("old case separation boilerplate", r"이 케이스를 별도로 두는 이유는"),
    ("old case pass-signal boilerplate", r"이 케이스의 PASS 신호는"),
    ("mvcc observation-focus scaffold", r"관측 초점:"),
    ("scenario numbered scaffold", r"scenario [0-9]+"),
    ("broad operations time-window boilerplate", r"같은 5분 구간 안에서"),
    ("broad operations vague-team boilerplate", r"팀 안에서 '느리다', '안 보인다'"),
    ("beginner-template boilerplate", r"초급자는 보통"),
    ("korean grammar artifact", r"다이다"),
    ("english state-observation scaffold", r"observe local state"),
    ("english next-action scaffold", r"decide next safe action"),
    ("padding confession", r"길이를 늘리기 위해서"),
    ("old repetition self-check sentence", r"이 검증은 단순 복습 문장이 아니라"),
    ("internal registry meta leaked into reader body", r"DB deep-dive registry"),
    ("internal length gate leaked into reader body", r"20,000자 하한"),
    ("worker ownership meta leaked into reader body", r"Worker [A-Z][^\n]{0,40}(소유|범위|tranche|ownership|scope)"),
    ("tranche meta leaked into reader body", r"tranche"),
    ("ownership-scope meta leaked into reader body", r"소유 범위"),
    ("generic expected-evidence suffix", r"expected evidence must match the described state transition"),
    ("generic application evidence boilerplate", r"응답만 그럴듯하고 내부 row, 로그, metric"),
    ("generic application proof boilerplate", r"실제 증거로 남는지 확인하는 방식으로 닫는다"),
    ("internal numeric target leaked into reader body", r"20,000자, trace"),
    ("internal registry-source map leaked into reader body", r"DU registry/source-map"),
    ("reader-facing registry meta leaked into body", r"문서의 registry"),
    ("source-map meta leaked into reader body", r"source-map"),
    ("DU seed meta leaked into reader body", r"DU[0-9]{2} seed"),
    ("seed linkage meta leaked into reader body", r"seed로 연결"),
    ("generic case-input boilerplate", r"케이스에서는 [^\n]{0,80}핵심 입력이다"),
    ("generic case-observation boilerplate", r"이 판단은 [^\n]{0,80}에서 다시 확인되어야 한다"),
    ("generic hand-replay boilerplate", r"손재생 점검의 PASS 조건"),
)


@dataclass(frozen=True)
class DU:
    id: str
    target: str
    section: str
    min_chars: int
    source_requirement: str
    teaching_spine: str
    required_trap: str


@dataclass(frozen=True)
class SourceEntry:
    id: str
    local_seed: str
    official_sources: str
    lab_or_observability: str
    preservation_disposition: str
    source_status: str
    notes: str


def load_registry() -> list[DU]:
    with REGISTRY.open(encoding="utf-8", newline="") as f:
        rows = list(csv.DictReader(f, delimiter="\t"))
    units: list[DU] = []
    seen: set[str] = set()
    for row in rows:
        du_id = row["id"].strip()
        if du_id in seen:
            raise ValueError(f"duplicate DU id: {du_id}")
        seen.add(du_id)
        units.append(
            DU(
                id=du_id,
                target=row["target"].strip(),
                section=row["section"].strip(),
                min_chars=int(row["min_chars"].strip()),
                source_requirement=row["source_requirement"].strip(),
                teaching_spine=row["teaching_spine"].strip(),
                required_trap=row["required_trap"].strip(),
            )
        )
    expected = [f"DU{i:02d}" for i in range(1, 57)]
    actual = [unit.id for unit in units]
    if actual != expected:
        raise ValueError(f"DU registry must be exactly DU01-DU56 in order, got {actual[:3]}...{actual[-3:]}")
    return units


def is_placeholder(value: str) -> bool:
    return value.strip().lower() in PLACEHOLDER_VALUES


def has_official_source(value: str) -> bool:
    return "https://" in value or "http://" in value or "RFC " in value or "ISO " in value


def has_lab_or_observability(value: str) -> bool:
    tokens = [
        "labs/",
        "EXPLAIN",
        "SHOW",
        "SELECT",
        "pg_",
        "performance_schema",
        "INFORMATION_SCHEMA",
        "check:",
        "test:",
        "metric:",
        "log:",
    ]
    return any(token in value for token in tokens)


def load_source_entries(args: argparse.Namespace, failures: list[str]) -> dict[str, SourceEntry]:
    if not SOURCE_MAP_TSV.exists():
        if not args.allow_missing_source_map:
            failures.append("source-map.tsv is missing")
        return {}

    with SOURCE_MAP_TSV.open(encoding="utf-8", newline="") as f:
        reader = csv.DictReader(f, delimiter="\t")
        if reader.fieldnames is None:
            failures.append("source-map.tsv has no header")
            return {}
        missing_columns = SOURCE_REQUIRED_COLUMNS.difference(reader.fieldnames)
        if missing_columns:
            failures.append(f"source-map.tsv missing columns: {sorted(missing_columns)}")
            return {}
        rows = list(reader)

    entries: dict[str, SourceEntry] = {}
    for row in rows:
        du_id = row["id"].strip()
        if du_id in entries:
            failures.append(f"source-map.tsv duplicate DU id: {du_id}")
            continue
        entry = SourceEntry(
            id=du_id,
            local_seed=row["local_seed"].strip(),
            official_sources=row["official_sources"].strip(),
            lab_or_observability=row["lab_or_observability"].strip(),
            preservation_disposition=row["preservation_disposition"].strip(),
            source_status=row["source_status"].strip(),
            notes=row["notes"].strip(),
        )
        entries[du_id] = entry

        if is_placeholder(entry.local_seed):
            failures.append(f"{du_id}: source-map local_seed is empty or placeholder")
        if is_placeholder(entry.official_sources) or not has_official_source(entry.official_sources):
            failures.append(f"{du_id}: source-map official_sources lacks an official URL/RFC/ISO reference")
        if is_placeholder(entry.lab_or_observability) or not has_lab_or_observability(entry.lab_or_observability):
            failures.append(f"{du_id}: source-map lab_or_observability lacks a lab path or concrete observation command")
        if entry.preservation_disposition not in ALLOWED_DISPOSITIONS:
            failures.append(f"{du_id}: invalid preservation_disposition {entry.preservation_disposition!r}")
        if entry.source_status != "verified" and not args.allow_planned_sources:
            failures.append(f"{du_id}: source_status must be verified for whole-complete validation")

    return entries


def split_sections(text: str) -> dict[str, str]:
    lines = text.splitlines()
    starts: list[tuple[int, str]] = []
    for idx, line in enumerate(lines):
        if line.startswith("## "):
            starts.append((idx, line[3:].strip()))
    if not starts:
        return {"<file-body>": "\n".join(lines[1:] if lines and lines[0].startswith("# ") else lines).strip()}
    sections: dict[str, str] = {}
    for pos, (start, title) in enumerate(starts):
        end = starts[pos + 1][0] if pos + 1 < len(starts) else len(lines)
        sections[title] = "\n".join(lines[start + 1 : end]).strip()
    return sections


def has_ascii_or_worked_trace(body: str) -> bool:
    fence = re.search(r"```(?:text|sql|mermaid|bash)?\n[\s\S]*?```", body)
    arrows = "->" in body or "=>" in body or "-->" in body
    table = bool(re.search(r"^\|.+\|$", body, re.MULTILINE))
    box = "+---" in body or "----+" in body or "|   " in body
    return bool(fence and (arrows or table or box or "SELECT" in body or "BEGIN" in body))


def has_practical_failure_trap(body: str) -> bool:
    return any(token in body for token in ["함정", "장애", "주의", "실패", "운영에서", "실무에서", "위험"])


def has_history_or_origin(body: str) -> bool:
    return any(token in body for token in ["배경", "등장", "역사", "왜 생겼", "문제가 생겼", "필요해졌"])


def has_observability_or_verification(body: str) -> bool:
    return any(token in body for token in ["검증", "관측", "확인", "재현", "EXPLAIN", "SHOW ", "SELECT ", "pg_", "performance_schema"])


def has_natural_flow_markers(body: str) -> bool:
    paragraphs = [p.strip() for p in body.split("\n\n") if p.strip()]
    if len(paragraphs) < 8:
        return False
    return any(token in body for token in ["그래서", "이 때문에", "그 결과", "반대로", "여기서"])


def forbidden_prose_hits(body: str) -> list[str]:
    hits: list[str] = []
    for label, pattern in FORBIDDEN_PROSE_PATTERNS:
        if re.search(pattern, body, re.IGNORECASE):
            hits.append(label)
    return hits


def empty_nested_heading_hits(body: str) -> list[str]:
    lines = body.splitlines()
    hits: list[str] = []
    for idx, line in enumerate(lines):
        if not re.match(r"^#{3,6} ", line):
            continue
        next_idx = idx + 1
        while next_idx < len(lines) and not lines[next_idx].strip():
            next_idx += 1
        if next_idx >= len(lines) or re.match(r"^#{1,6} ", lines[next_idx]):
            hits.append(line.strip())
    return hits


def fence_count_ok(text: str) -> bool:
    return text.count("```") % 2 == 0


def duplicate_long_paragraphs(paths: list[Path]) -> list[str]:
    seen: dict[str, Path] = {}
    duplicates: list[str] = []
    for path in paths:
        text = path.read_text(encoding="utf-8")
        in_fence = False
        for raw in text.splitlines():
            if raw.startswith("```"):
                in_fence = not in_fence
                continue
            if in_fence:
                continue
            line = raw.strip()
            if len(line) < 180:
                continue
            key = re.sub(r"\s+", " ", line)
            if key in seen:
                duplicates.append(f"{path}: duplicates long paragraph also in {seen[key]}")
            else:
                seen[key] = path
    return duplicates


def validate(args: argparse.Namespace) -> int:
    failures: list[str] = []
    units = load_registry()
    source_map = SOURCE_MAP.read_text(encoding="utf-8") if SOURCE_MAP.exists() else ""
    source_entries = load_source_entries(args, failures)
    if not source_map and not args.allow_missing_source_map:
        failures.append("source-map.md is missing")

    if "15,000" in ACTIVE_PLAN.read_text(encoding="utf-8") or "15000" in ACTIVE_PLAN.read_text(encoding="utf-8"):
        failures.append("active plan still contains the old length requirement")
    if ACTIVE_WORK.exists():
        work_text = ACTIVE_WORK.read_text(encoding="utf-8")
        if "15,000" in work_text or "15000" in work_text:
            failures.append("active WORK still contains the old length requirement")

    checked_paths: set[Path] = set()
    for unit in units:
        path = ROOT / unit.target
        checked_paths.add(path)
        if not path.exists():
            failures.append(f"{unit.id}: missing target file {unit.target}")
            continue
        text = path.read_text(encoding="utf-8")
        if not fence_count_ok(text):
            failures.append(f"{unit.id}: unbalanced fenced code blocks in {unit.target}")
        sections = split_sections(text)
        body = sections.get(unit.section)
        if body is None:
            failures.append(f"{unit.id}: missing ## section {unit.section!r} in {unit.target}")
            continue
        body_len = len(body)
        if body_len < unit.min_chars:
            failures.append(f"{unit.id}: body length {body_len} < {unit.min_chars}")
        if not has_ascii_or_worked_trace(body):
            failures.append(f"{unit.id}: missing ASCII/code/table worked trace")
        if not has_practical_failure_trap(body):
            failures.append(f"{unit.id}: missing senior practical failure trap")
        if not has_history_or_origin(body):
            failures.append(f"{unit.id}: missing background/history/origin explanation")
        if not has_observability_or_verification(body):
            failures.append(f"{unit.id}: missing observability or verification path")
        if not has_natural_flow_markers(body):
            failures.append(f"{unit.id}: prose flow markers or paragraph depth too weak")
        for hit in forbidden_prose_hits(body):
            failures.append(f"{unit.id}: forbidden boilerplate/scaffold remains: {hit}")
        for heading in empty_nested_heading_hits(body):
            failures.append(f"{unit.id}: empty nested heading remains: {heading}")
        if source_map and unit.id not in source_map:
            failures.append(f"{unit.id}: missing source-map entry")
        if source_entries and unit.id not in source_entries:
            failures.append(f"{unit.id}: missing source-map.tsv entry")
        if source_entries and unit.id in source_entries:
            entry = source_entries[unit.id]
            if unit.source_requirement not in entry.notes and unit.source_requirement not in entry.local_seed:
                failures.append(f"{unit.id}: source-map.tsv does not reference registry source requirement")

    if checked_paths:
        existing_paths = [path for path in checked_paths if path.exists()]
        failures.extend(duplicate_long_paragraphs(sorted(existing_paths)))

    if failures:
        print("FAIL")
        for failure in failures:
            print(f"- {failure}")
        return 1
    print("PASS")
    print(f"validated_units={len(units)}")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--allow-missing-source-map", action="store_true")
    parser.add_argument("--allow-planned-sources", action="store_true")
    args = parser.parse_args()
    return validate(args)


if __name__ == "__main__":
    sys.exit(main())
