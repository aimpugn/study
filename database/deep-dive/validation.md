# DB deep-dive 검증 계약

이 디렉터리의 본문은 `database/deep-dive/du-registry.tsv`에 등록된 DU를 모두 닫아야 complete입니다.
사용자의 현재 기준은 각 DU 최소 20,000자입니다.
20,000자는 “이만큼만 채우면 된다”가 아니라, 이보다 짧으면 깊은 설명으로 인정하지 않는 하한선입니다.

검증 명령은 다음입니다.

```bash
python3 database/deep-dive/validate_deep_dive.py
```

이 명령은 실패하면 non-zero로 종료해야 합니다.
작성 중 중간 상태를 확인할 때도 이 명령의 FAIL을 “아직 남은 작업 목록”으로 읽고, whole complete를 선언할 때는 반드시 PASS여야 합니다.

## 통과 조건

- `DU01`부터 `DU56`까지 registry 순서가 유지된다.
- 각 DU의 target file과 `##` section이 존재한다.
- 각 DU section body가 20,000자 이상이다.
- 각 DU에는 실제 값을 따라가는 ASCII/code/table trace가 있다.
- 각 DU에는 실무 장애 함정, 등장 배경, 관측/검증 경로, 자연스러운 문단 흐름을 확인할 수 있는 본문 신호가 있다.
- `source-map.tsv`에 각 DU의 local seed, official source, lab 또는 observability path, preservation disposition, source status가 parseable하게 연결된다.
- whole-complete 시점의 `source-map.tsv`는 모든 DU의 `source_status`가 `verified`여야 한다.
- `source-map.md`에는 사람이 읽을 수 있는 보존 원칙과 DU coverage가 남아 있다.
- fenced code block 균형이 맞다.
- 긴 문단의 중복이 없다.
- 본문에 이전 scaffold, 반복 검산 문장, 내부 제작 메타(`Worker`, `tranche`, registry 길이 기준 등)가 남아 있지 않다.
- active plan과 WORK에 이전 길이 기준이 남아 있지 않다.

## 검증의 한계

스크립트는 문서가 정말 좋은 글인지 완전히 판정하지 못합니다.
예를 들어 문단 사이 논리가 물 흐르듯 자연스러운지, 비유가 실제 이해를 돕는지, 시니어의 실전 감각이 살아 있는지는 기계적으로만 판단하기 어렵습니다.
그래서 validator PASS는 필요조건이고, 충분조건은 아닙니다.
각 DU는 validator PASS 뒤에도 `review-kernel`, `humanize-korean`, `study-explanation` 관점의 critic review를 통과해야 합니다.

작성 중 출처 후보만 먼저 고정한 상태를 확인하려면 다음처럼 실행할 수 있습니다.

```bash
python3 database/deep-dive/validate_deep_dive.py --allow-planned-sources
```

이 옵션은 중간 점검용입니다.
최종 whole-complete 검증에서는 옵션 없이 실행해야 합니다.
