# Validation

이 디렉터리의 검증은 세 층으로 나뉩니다.

1. 구조 검증

    `python3 interviews/database-deep-dive/tools/validate_interview_database_deep_dive.py`를 실행합니다. 이 명령은 정식 문서가 모두 topic 목록에 들어 있고, planned 상태가 남아 있지 않으며, 각 문서가 필수 섹션과 source 링크, audit 연결, reader-facing 금지 패턴 검사를 통과하는지 확인합니다. 반복 문구, 중복 H3 heading, 긴 문단 반복, 민감 source unit, 반복된 composition risk도 실패로 봅니다.

2. claim 검증

    `audit/claim-audit.tsv`는 문서의 하중을 지탱하는 claim을 source, support tier, 적용 범위, 반례와 함께 기록합니다. 모든 문장을 표에 복제하는 것이 목적이 아니라, 문서 전체 판단을 바꾸는 claim이 근거 없이 본문에 들어가지 않게 하는 것이 목적입니다.

3. composition 검증

    `audit/composition-audit.tsv`는 각 문서의 필수 섹션이 어떤 thesis를 맡는지, 어떤 claim을 지지하는지, 어떤 큰 단위 오류를 막는지 기록합니다. 작은 문장이 맞아도 전체 설명이 틀릴 수 있으므로 이 검증이 필요합니다.

완료 판정은 validator PASS만으로 닫지 않습니다. 전체 문서 수, planned 상태 0개, source boundary 갱신, claim/composition audit 연결, 다중 검수 결과, path-limited commit/push가 함께 닫혀야 합니다.
