# Orchestrator findings

## Verdict

`REWORK -> PARTIAL APPLY -> REVIEWED ALIGN for this tranche`.

정식 경로는 `interviews/database-deep-dive/`다. `database/deep-dive/`는 이전 실행에서 만들어진 장문 초안과 coverage 자료로 보존하되, 정식 인터뷰 학습 산출물로 그대로 승격하지 않는다.

## Multi-agent synthesis

- Source Architect: source와 generated draft의 경계를 먼저 나누지 않으면 잘못된 위치 문제를 반복한다. `source-boundary.tsv`, `claim-audit.tsv`, `composition-audit.tsv`가 필요하다.
- Composition Critic: 작은 claim이 각각 맞아도 반복, 순서, DBMS 경계, log 소비자 경계가 흐리면 큰 설명은 틀린다. 첫 파일럿은 WAL/redo/undo/crash recovery/PITR처럼 반례가 강한 주제가 적합하다.
- Protocol Sentinel: 이미 push된 `database/deep-dive` commit은 되돌리거나 rewriting하지 않는다. path-limited corrective commit으로 정식 경로와 noncanonical 표시를 남기고, whole-complete를 주장하지 않는다.

## Frozen checklist

- [x] `interviews/database-deep-dive/`를 canonical 위치로 만든다.
- [x] `database/deep-dive/`를 이전 생성 초안/coverage 자료로 표시한다.
- [x] source 문장 claim audit와 composition audit의 파일 구조를 둔다.
- [x] 파일럿 주제를 하나 골라 실제 본문에 적용한다.
- [x] 구조 validator를 둔다.
- [ ] 전체 DB 면접 심화 코퍼스를 전부 작성한다.
- [ ] 전체 planned topic의 claim/composition audit를 닫는다.
- [ ] 모든 문서에 대해 critic review와 직접 재생 경로를 닫는다.

## Claim cards

### C-001

- claim: 정식 산출물 위치는 `interviews/database-deep-dive/`여야 한다.
- reason: `interviews` 프로젝트의 목적은 면접 질문에 짧게 답하고 깊게 설명할 수 있는 정식 답변 자산을 만드는 것이다.
- evidence: `interviews/PROJECT_INTENT.md`, `interviews/USECASE.md`, 사용자 피드백.
- attack: `study/database`에 이미 초안이 있으므로 그대로 두는 편이 빠르다는 반론.
- response: `ACCEPT_REPAIR`. 기존 초안은 보존하되 canonical 권한은 낮춘다.
- status: APPLY.

### C-002

- claim: sentence-level truth audit만으로는 충분하지 않고 composition-level audit가 필요하다.
- reason: WAL, redo, undo, PITR 같은 개념은 각각의 문장이 맞아도 조합 순서와 DBMS 경계가 틀리면 잘못된 mental model을 만든다.
- evidence: 사용자 피드백, old generated draft의 반복/경계 흐림, `composition-audit.tsv`.
- attack: claim audit가 충분히 세밀하면 composition audit가 중복일 수 있다는 반론.
- response: `REBUT`. claim audit는 작은 단위의 사실 여부를 보지만, section thesis와 bridge claim의 전체 효과를 판정하지 못한다.
- status: APPLY.

### C-003

- claim: 현재 작업은 corrective structural plus pilot closure이며 whole-complete가 아니다.
- reason: planned topic 전체가 작성되지 않았고, 모든 source 문장의 claim audit도 닫히지 않았다.
- evidence: `topic-registry.tsv`, frozen checklist.
- attack: 첫 파일럿과 validator가 PASS하면 완료로 볼 수 있지 않냐는 반론.
- response: `DOWNGRADE`. validator PASS는 구조적 필요조건이고 파일럿 완료는 전체 완료가 아니다.
- status: APPLY as PARTIAL.

## Review Repair Log

### Source Architect finding

- finding: generated draft rows were incorrectly allowed to appear as `T1 Direct Evidence` for domain mechanism claims.
- repair: `claim-audit.tsv` now has `source_class`; generated draft domain claims are downgraded to `T2 Strong Inference` or used only as composition/process direct evidence. Official source rows were added for PostgreSQL WAL, InnoDB redo, and MySQL binary-log PITR.
- validator: `validate_interview_database_deep_dive.py` now fails if generated draft material is used as `T1` domain evidence.

### Source line trace finding

- finding: one PITR row cited the wrong old-draft line span.
- repair: LOG-C05 now points to the matching `backup, restore, PITR` section line, and local raw spans are checked against cited line ranges.

### Protocol Sentinel finding

- finding: old `database/database-deep-study-plan.md` still contained active-looking canonical path and whole-complete language.
- repair: stale path and closure sections now explicitly say they are historical and invalid for the current `interviews/database-deep-dive/` reconstruction.

### Validation result

- partial structural validation: `python3 interviews/database-deep-dive/tools/validate_interview_database_deep_dive.py --allow-planned` passes.
- whole validation: the same command without `--allow-planned` fails because planned topics remain. This is the intended fail-closed result.
