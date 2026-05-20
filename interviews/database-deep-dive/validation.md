# Database interview deep-dive validation

이 검증은 정식 인터뷰 DB 심화 문서가 최소 구조와 금지 패턴을 지키는지 확인하는 보조 장치다.
사용자가 지적한 것처럼, 작은 claim이 맞는지 보는 일은 필요조건이고, 큰 설명이 맞는지 보는 일은 별도 조건이다.
따라서 이 validator가 PASS해도 전체 품질이 자동으로 complete가 되지는 않는다.

## 실행

```bash
python3 interviews/database-deep-dive/tools/validate_interview_database_deep_dive.py --allow-planned
```

`--allow-planned`는 현재처럼 파일럿과 구조를 검증하는 중간 상태에서만 쓴다.
전체 코퍼스 완료를 주장하려면 옵션 없이 실행해야 하며, 그때는 planned topic이 남아 있으면 실패해야 한다.

## 통과 조건

- canonical root가 `interviews/database-deep-dive`에 존재한다.
- source boundary, claim audit, composition audit가 parseable TSV로 존재한다.
- `database/deep-dive`는 source가 아니라 generated draft/noncanonical 자료로 분류된다.
- generated draft/noncanonical 자료는 문서 구성, 반복, coverage 증거로는 쓸 수 있지만, domain truth를 지탱하는 `T1 Direct Evidence`로는 쓸 수 없다.
- claim audit의 `verification_ref`는 `audit/evidence-refs.tsv`에서 해소되어야 한다.
- composition audit가 참조하는 claim id는 `audit/claim-audit.tsv`에 존재해야 한다.
- reader-facing 본문에 `DU12`, `source-map`, `registry`, `20,000자 하한` 같은 이전 제작 메타가 새지 않는다.
- completed topic은 필수 section을 가진다.
- completed topic은 `2-5분 개요`, 깊은 메커니즘, 직접 재생 경로, 꼬리 질문, 함정 질문, 더 깊게 볼 자료를 포함한다.
- completed topic은 claim audit와 composition audit에 연결된다.

## 한계

이 검증은 문서의 구조, 금지 패턴, audit 연결을 확인한다.
문장이 실제로 자연스러운지, 작은 claim이 큰 설명 안에서 오해 없이 엮였는지, 면접 답변으로 말했을 때 설득력이 있는지는 review-kernel, study-explanation, humanize-korean 관점의 사람이 읽는 검수로 다시 닫아야 한다.
