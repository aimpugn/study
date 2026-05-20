#!/usr/bin/env python3
from pathlib import Path


ROOT = Path(__file__).resolve().parents[3]
OUT = ROOT / "interviews/database-deep-dive/audit/source-boundary.tsv"


def classify(path: Path) -> tuple[str, str]:
    rel = path.relative_to(ROOT).as_posix()
    name = path.name
    suffix = path.suffix.lower()

    if name == ".DS_Store":
        return "system-noise", "macOS metadata; content excluded"
    if "/tools/esdump/auth.ini" in rel:
        return "sensitive-source-do-not-promote", "possible credential material; quote only after redaction"
    if rel.startswith("database/deep-dive/"):
        return "generated-draft-noncanonical", "previous generated DB deep-dive output; coverage/draft only"
    if rel == "database/database-deep-study-plan.md":
        return "historical-plan", "prior plan; source for process history, not domain truth"
    if suffix in {".jar", ".png", ".jpg", ".jpeg", ".gif", ".pdf"}:
        return "binary-lab-dependency", "binary artifact; inspect license/content before promotion"
    if rel.startswith("interviews/source/"):
        return "raw-interview-source", "raw reservoir for interview promotion"
    if rel == "interviews/database-storage-search-nosql.md":
        return "curated-placement-source", "large existing batch; source reservoir, not final style exemplar"
    if rel.startswith("database/"):
        return "database-source", "existing study/database source"
    return "other", "not part of current DB interview source boundary"


def main() -> int:
    paths: list[Path] = []
    for base in [ROOT / "database", ROOT / "interviews/source"]:
        if base.exists():
            paths.extend(p for p in base.rglob("*") if p.is_file())
    curated = ROOT / "interviews/database-storage-search-nosql.md"
    if curated.exists():
        paths.append(curated)

    rows = ["path\tclassification\tnote"]
    for path in sorted(set(paths)):
        classification, note = classify(path)
        rows.append(f"{path.relative_to(ROOT).as_posix()}\t{classification}\t{note}")

    OUT.write_text("\n".join(rows) + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
