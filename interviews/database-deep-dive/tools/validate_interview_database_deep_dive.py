#!/usr/bin/env python3
import argparse
import csv
import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[3]
CANONICAL = ROOT / "interviews/database-deep-dive"
AUDIT = CANONICAL / "audit"
REGISTRY = CANONICAL / "topic-registry.tsv"
FORBIDDEN_READER_PATTERNS = [
    re.compile(r"\bDU\d{2}\b"),
    re.compile(r"source-map"),
    re.compile(r"\bregistry\b", re.IGNORECASE),
    re.compile(r"20,000자\s*하한"),
]
REQUIRED_SECTIONS = [
    "## 2-5분 개요",
    "## 먼저 잡아야 할 작은 모델",
    "## 깊은 메커니즘",
    "## DBMS별 경계",
    "## 직접 재생해 보기",
    "## 면접 꼬리 질문",
    "## 함정 질문",
    "## 더 깊게 볼 자료",
]


def read_tsv(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as f:
        return list(csv.DictReader(f, delimiter="\t"))


def fail(errors: list[str], message: str) -> None:
    errors.append(message)


def validate_tsv(path: Path, required_headers: list[str], errors: list[str]) -> list[dict[str, str]]:
    if not path.exists():
        fail(errors, f"missing TSV: {path.relative_to(ROOT)}")
        return []
    rows = read_tsv(path)
    with path.open(encoding="utf-8", newline="") as f:
        reader = csv.reader(f, delimiter="\t")
        headers = next(reader, [])
    missing = [h for h in required_headers if h not in headers]
    if missing:
        fail(errors, f"{path.relative_to(ROOT)} missing headers: {', '.join(missing)}")
    return rows


def split_refs(value: str) -> list[str]:
    return [item.strip() for item in value.split(";") if item.strip()]


def get_boundary_map(rows: list[dict[str, str]]) -> dict[str, str]:
    return {row["path"]: row["classification"] for row in rows}


def source_span_text(path: Path, start: str, end: str) -> str | None:
    if not start.isdigit() or not end.isdigit():
        return None
    lines = path.read_text(encoding="utf-8", errors="replace").splitlines()
    start_i = max(int(start) - 1, 0)
    end_i = min(int(end), len(lines))
    return "\n".join(lines[start_i:end_i])


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--allow-planned", action="store_true")
    args = parser.parse_args()

    errors: list[str] = []
    if not CANONICAL.exists():
        fail(errors, "canonical root does not exist")

    boundary_rows = validate_tsv(
        AUDIT / "source-boundary.tsv",
        ["path", "classification", "note"],
        errors,
    )
    claim_rows = validate_tsv(
        AUDIT / "claim-audit.tsv",
        [
            "claim_id",
            "source_path",
            "source_class",
            "claim_text",
            "support_tier",
            "disposition",
            "target_section",
            "verification_ref",
        ],
        errors,
    )
    composition_rows = validate_tsv(
        AUDIT / "composition-audit.tsv",
        [
            "section_id",
            "target_path",
            "section_title",
            "section_thesis",
            "supporting_claim_ids",
            "final_disposition",
            "repair_action",
        ],
        errors,
    )
    evidence_rows = validate_tsv(
        AUDIT / "evidence-refs.tsv",
        ["ref_id", "source_type", "url_or_path", "scope", "note"],
        errors,
    )

    if boundary_rows:
        draft_rows = [r for r in boundary_rows if r["path"].startswith("database/deep-dive/")]
        if not draft_rows:
            fail(errors, "source-boundary.tsv does not classify database/deep-dive")
        bad_draft = [
            r["path"]
            for r in draft_rows
            if r["classification"] != "generated-draft-noncanonical"
        ]
        if bad_draft:
            fail(errors, "database/deep-dive rows must be generated-draft-noncanonical")

    boundary_map = get_boundary_map(boundary_rows)
    evidence_ids = {row["ref_id"] for row in evidence_rows}
    claim_ids = {row["claim_id"] for row in claim_rows}

    for row in claim_rows:
        source_path = row["source_path"]
        source_class = row["source_class"]
        support_tier = row["support_tier"]
        claim_type = row["claim_type"]
        if source_path in boundary_map and source_class != boundary_map[source_path]:
            fail(errors, f"{row['claim_id']} source_class mismatch for {source_path}")
        if (
            source_class == "generated-draft-noncanonical"
            and support_tier == "T1 Direct Evidence"
            and claim_type not in {"composition", "process"}
        ):
            fail(errors, f"{row['claim_id']} uses generated draft as T1 domain evidence")
        for ref in split_refs(row.get("verification_ref", "")):
            if ref not in evidence_ids:
                fail(errors, f"{row['claim_id']} has unresolved verification_ref: {ref}")
        if source_path.startswith("http"):
            continue
        local_source = ROOT / source_path
        if local_source.exists():
            raw = row.get("raw_sentence_or_span", "")
            if not raw.startswith(("trace block:", "repeated ", "official summary:", "summary:")):
                span = source_span_text(local_source, row.get("start_line", ""), row.get("end_line", ""))
                if span is not None and raw not in span:
                    fail(errors, f"{row['claim_id']} raw span is not found in cited source lines")

    for row in composition_rows:
        refs = (
            split_refs(row.get("supporting_claim_ids", ""))
            + split_refs(row.get("bridge_claim_ids", ""))
            + split_refs(row.get("rejected_or_narrowed_claim_ids", ""))
        )
        for ref in refs:
            if ref not in claim_ids:
                fail(errors, f"{row['section_id']} references missing claim_id: {ref}")

    registry_rows = validate_tsv(
        REGISTRY,
        ["topic_id", "status", "target", "title", "audit_status"],
        errors,
    )
    if registry_rows and not args.allow_planned:
        planned = [r["topic_id"] for r in registry_rows if r["status"] == "planned"]
        if planned:
            fail(errors, f"planned topics remain without --allow-planned: {', '.join(planned)}")

    completed = [r for r in registry_rows if r.get("status") in {"pilot-complete", "complete"}]
    for row in completed:
        target = CANONICAL / row["target"]
        if not target.exists():
            fail(errors, f"completed topic missing target: {row['target']}")
            continue
        text = target.read_text(encoding="utf-8")
        for section in REQUIRED_SECTIONS:
            if section not in text:
                fail(errors, f"{row['target']} missing required section: {section}")
        for pattern in FORBIDDEN_READER_PATTERNS:
            if pattern.search(text):
                fail(errors, f"{row['target']} contains reader-facing generation meta: {pattern.pattern}")
        if len(text) < 20000:
            fail(errors, f"{row['target']} is shorter than pilot deep-dive floor: {len(text)} chars")

    if errors:
        for error in errors:
            print(f"FAIL: {error}", file=sys.stderr)
        return 1
    print("PASS: interviews database deep-dive structural validation")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
