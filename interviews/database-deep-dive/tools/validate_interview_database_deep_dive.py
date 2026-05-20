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
FORBIDDEN_GENERIC_READER_PHRASES = [
    "이 설명을 다른 주제에 옮길 때도 같은 질문을 씁니다",
    "면접에서는 이 주제를 처음부터 내부 구조 전체로 펼치면",
    "이 주제에서 가장 흔한 실수는 이름이 같은 것을",
    "실무에서는 개념의 이름보다 관측 가능한 신호가 중요합니다",
    "PASS 신호는 관측값이 설명한 경계와 맞는 것입니다",
    "답할 때는 먼저 한 문장으로 결론을 말하고",
    "이 답은 부분적으로 그럴듯할 수 있지만",
    "여기서 조심할 점은 공통 용어를 구현 세부까지 같은 것으로 착각하지 않는 것입니다",
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
ALLOWED_COMPLETE_STATUSES = {"complete"}
ALLOWED_IN_PROGRESS_STATUSES = {"planned"}


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


def expand_source_unit(value: str) -> set[str]:
    if "*" not in value:
        return {value}
    expanded: set[str] = set()
    for path in ROOT.glob(value):
        if path.is_dir():
            expanded.update(p.relative_to(ROOT).as_posix() for p in path.rglob("*") if p.is_file())
        elif path.exists():
            expanded.add(path.relative_to(ROOT).as_posix())
    return expanded or {value}


def markdown_h3_titles(text: str) -> list[str]:
    return [line.strip() for line in text.splitlines() if line.startswith("### ")]


def repeated_long_paragraphs(text: str) -> list[str]:
    paragraphs = [
        re.sub(r"\s+", " ", paragraph.strip())
        for paragraph in text.split("\n\n")
        if len(paragraph.strip()) >= 120
    ]
    return sorted({paragraph for paragraph in paragraphs if paragraphs.count(paragraph) > 1})


def cross_doc_links(text: str) -> list[str]:
    return re.findall(r"\]\(((?!#)(?:\d{2}-)[^)#]+\.md)(?:#[^)]+)?\)", text)


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
        non_complete = [
            f"{r['topic_id']}={r['status']}"
            for r in registry_rows
            if r.get("status") not in ALLOWED_COMPLETE_STATUSES
        ]
        if non_complete:
            fail(errors, f"non-complete topic statuses remain: {', '.join(non_complete)}")

    for row in registry_rows:
        status = row.get("status", "")
        if status not in ALLOWED_COMPLETE_STATUSES | ALLOWED_IN_PROGRESS_STATUSES:
            fail(errors, f"{row.get('topic_id', '<unknown>')} has unsupported status: {status}")

    composition_targets = {row.get("target_path", "") for row in composition_rows}
    claim_target_refs = {row.get("target_section", "").split("#", 1)[0] for row in claim_rows}

    completed = [r for r in registry_rows if r.get("status") in ALLOWED_COMPLETE_STATUSES]
    completed_targets = {r["target"] for r in completed}
    cross_doc_paragraphs: dict[str, list[str]] = {}
    for index, row in enumerate(completed, start=1):
        expected_prefix = f"{index:02d}-"
        if not row["target"].startswith(expected_prefix):
            fail(errors, f"{row['topic_id']} target must start with reading-order prefix {expected_prefix}: {row['target']}")

    sensitive_sources = {
        path
        for path, classification in boundary_map.items()
        if classification == "sensitive-source-do-not-promote"
    }
    for row in completed:
        source_units = split_refs(row.get("source_units", ""))
        expanded_units: set[str] = set()
        for source_unit in source_units:
            expanded_units.update(expand_source_unit(source_unit))
        sensitive_used = sorted(expanded_units & sensitive_sources)
        if sensitive_used:
            fail(
                errors,
                f"{row['topic_id']} lists sensitive source_units: {', '.join(sensitive_used)}",
            )

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
        for phrase in FORBIDDEN_GENERIC_READER_PHRASES:
            if phrase in text:
                fail(errors, f"{row['target']} contains generic repeated reader phrase: {phrase}")
        h3_titles = markdown_h3_titles(text)
        duplicate_h3 = sorted({title for title in h3_titles if h3_titles.count(title) > 1})
        if duplicate_h3:
            fail(errors, f"{row['target']} contains duplicate H3 headings: {', '.join(duplicate_h3)}")
        repeated_paragraphs = repeated_long_paragraphs(text)
        if repeated_paragraphs:
            preview = repeated_paragraphs[0][:90]
            fail(errors, f"{row['target']} contains repeated long paragraph: {preview}")
        outbound_links = [link for link in cross_doc_links(text) if link != row["target"]]
        if len(set(outbound_links)) < 2:
            fail(errors, f"{row['target']} must link to at least two related deep-dive docs")
        for link in outbound_links:
            if link not in completed_targets:
                fail(errors, f"{row['target']} links to unknown deep-dive doc: {link}")
        for paragraph in [
            re.sub(r"\s+", " ", paragraph.strip())
            for paragraph in text.split("\n\n")
            if len(paragraph.strip()) >= 160 and not paragraph.lstrip().startswith("- [")
        ]:
            cross_doc_paragraphs.setdefault(paragraph, []).append(row["target"])
        if len(text) < 20000:
            fail(errors, f"{row['target']} is shorter than pilot deep-dive floor: {len(text)} chars")
        if row["target"] not in composition_targets:
            fail(errors, f"{row['target']} has no composition-audit row")
        if row["target"] not in claim_target_refs and row["target"] != "README.md":
            fail(errors, f"{row['target']} has no claim-audit target reference")
        is_search_document_topic = row["target"].endswith("search-document-nosql-engine.md")
        if "https://www.postgresql.org/docs/current/" not in text and "https://dev.mysql.com/doc/refman/8.4/" not in text and not is_search_document_topic:
            fail(errors, f"{row['target']} lacks a primary PostgreSQL/MySQL source link")
        if is_search_document_topic and "https://www.elastic.co/" not in text and "https://firebase.google.com/" not in text:
            fail(errors, f"{row['target']} lacks a primary search/document-store source link")

    canonical_docs = {
        path.name
        for path in CANONICAL.glob("*.md")
        if path.name not in {"README.md", "validation.md"}
    }
    registry_docs = {row["target"] for row in registry_rows}
    extra_docs = sorted(canonical_docs - registry_docs)
    if extra_docs:
        fail(errors, f"canonical docs not listed in topic-registry.tsv: {', '.join(extra_docs)}")

    repeated_cross_doc = {
        paragraph: targets
        for paragraph, targets in cross_doc_paragraphs.items()
        if len(set(targets)) > 1
    }
    if repeated_cross_doc:
        paragraph, targets = next(iter(repeated_cross_doc.items()))
        fail(errors, f"repeated long paragraph across docs ({', '.join(sorted(set(targets)))}): {paragraph[:90]}")

    risk_counts: dict[str, int] = {}
    for row in composition_rows:
        risk = row.get("large_unit_risk", "")
        risk_counts[risk] = risk_counts.get(risk, 0) + 1
    repeated_risks = [risk for risk, count in risk_counts.items() if risk and count > 8]
    if repeated_risks:
        fail(errors, "composition-audit.tsv repeats the same large_unit_risk across multiple topics")

    if errors:
        for error in errors:
            print(f"FAIL: {error}", file=sys.stderr)
        return 1
    print("PASS: interviews database deep-dive structural validation")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
