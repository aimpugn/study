"""Count what the end of each capture read interval is, for the transaction runs and the Spring runs.

Usage: python3 -I reads.py <screen html> [<screen html> ...]
       (open_source_internals/workbench/design/concept-5-tx.html ... concept-9-spring.html)

The decoder (decode/scenario_check.go 395-433 captureReads) gives each part of a capture the interval from the
recvfrom of its query to the next query the observation connection received, or to the start of the next capture
when no query follows, or to its own start in the last capture. So the end is an upper bound chosen by that rule,
not an event of the read. For every screen this prints, per kind of end, how many read intervals end there:
"next-capture" (end == the next capture's t), "same" (end == start, which must be the last capture), and
"query" (some other time, which must be before the next capture's t). The last line of each screen says whether
every end fits one of the three kinds; the exit code is 1 if any does not.
"""
import collections
import json
import re
import sys


def load(path):
    """Return the RUNS object of a screen file (run key -> decoded record)."""
    text = open(path, encoding="utf-8").read()
    m = re.search(r"(?m)^\s*const RUNS = (.*);\s*$", text)
    return json.loads(m.group(1))


def classify(runs):
    """Return a Counter of end kinds and the number of ends that fit no kind.

    Args:
        runs: the RUNS object of one screen.

    Returns:
        (Counter of "next-capture" / "same" / "query", number of ends that break the rule).

    """
    kinds, bad = collections.Counter(), 0
    for r in runs.values():
        caps = sorted(r["captures"], key=lambda c: c["t"])
        for k, c in enumerate(caps):
            nxt = caps[k + 1]["t"] if k + 1 < len(caps) else None
            for a, b in c["reads"].values():
                if b == a:
                    kinds["same"] += 1
                    bad += nxt is not None
                elif nxt is not None and b == nxt:
                    kinds["next-capture"] += 1
                else:
                    kinds["query"] += 1
                    bad += nxt is not None and b > nxt
    return kinds, bad


def main(paths):
    """Print the end kinds for every screen and return the exit code (1 if any end breaks the rule)."""
    failed = 0
    for path in paths:
        kinds, bad = classify(load(path))
        total = sum(kinds.values())
        print("%s\t읽기 구간 %d: 다음 관찰의 시작 %d, 끝이 시작과 같음(마지막 관찰) %d, 그 밖(다음 질의를 받은 때) %d\t규칙 밖 %d"
              % (path.rsplit("/", 1)[-1], total, kinds["next-capture"], kinds["same"], kinds["query"], bad))
        failed += bad
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
