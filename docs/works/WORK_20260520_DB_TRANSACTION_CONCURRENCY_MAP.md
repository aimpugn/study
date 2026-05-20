# WORK_20260520_DB_TRANSACTION_CONCURRENCY_MAP

## 0. Meta

- 작업 제목: DB transaction propagation / isolation / lock / deadlock / race condition 지도 보강
- WORK 파일 경로: `docs/works/WORK_20260520_DB_TRANSACTION_CONCURRENCY_MAP.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `explain | refactor_docs`
- 작업 깊이: `full`
- 원문 사용자 요청: `propagation, read committed, pdirty read, phantom read, repeatable read 등과 락과 데드락과 경쟁 조건 등을 제대로 한번 정리하고 싶습니다.`
- 대상 경로 / 자산:
  - `interviews/database-deep-dive/08-isolation-lock-deadlock.md`
  - `interviews/database-deep-dive/README.md`
  - `interviews/database-deep-dive/topic-registry.tsv`
  - `interviews/database-deep-dive/audit/*.tsv`
- 시작 일시: 2026-05-20 22:25:53 KST
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `test+commit`

## 1. Request Normalization

- goal: 면접/학습 문서에서 transaction propagation, isolation level, dirty read, phantom read, repeatable read, lock, deadlock, race condition을 하나의 역할 지도 안에서 정리한다.
- scope: 기존 canonical DB deep-dive 08번 문서를 보강하고, README/registry/audit evidence를 같은 의미로 동기화한다.
- mode: repo asset edit
- must_keep: 기존 DB deep-dive 구조와 validator 계약을 유지한다. 기존 dirty worktree의 다른 사용자 변경은 건드리지 않는다.
- extra_checks: 공식 Spring/PostgreSQL/MySQL 문서로 하중을 지탱하는 claim을 확인한다.

## 2. Root-First Framing

- 근본 문제: 기존 08번 문서는 isolation/lock/deadlock은 강하지만, 사용자가 함께 언급한 propagation과 race condition이 같은 지도 안에서 분리되지 않아 `REQUIRES_NEW`를 격리 수준처럼 이해하거나 race condition을 phantom read와 같은 말로 보는 오해가 남을 수 있었다.
- 작업 목표: propagation은 호출 경계 정책, isolation은 관측 계약, lock은 실행 수단, deadlock은 wait graph cycle, race condition은 실행 순서 의존 문제라는 축을 초반에 고정한다.
- 성공 정의: 본문, replay, 꼬리 질문, 함정 질문, registry, claim/evidence audit가 같은 의미를 말하고 validator가 PASS한다.
- PARTIAL 조건: 본문은 보강됐지만 audit/registry 또는 validator가 불일치한다.
- BLOCKED 조건: 공식 문서 확인이 불가능하거나 기존 validator 계약을 깨뜨리는 경우.

## 3. Reader & Internalization Contract

- 주 독자: DB/Spring transaction 질문을 면접에서 짧게 답하고 꼬리 질문에서 깊게 내려가야 하는 백엔드 개발자.
- teach-back 목표: 독자는 `전파는 transaction 경계`, `격리는 관측 계약`, `lock은 자원 충돌 조정`, `deadlock은 순환 대기`, `race condition은 실행 순서 의존 문제`라고 구분해 말할 수 있어야 한다.
- learner gap: dirty/non-repeatable/phantom 표와 Spring propagation enum을 같은 표면에 놓고 보면서, 둘이 같은 문제를 푸는 설정이라고 오해할 수 있다.
- first brick: 좌석 예약 예제에 `REQUIRED`, `REQUIRES_NEW`, `NESTED` 호출 trace를 붙여 commit 경계와 관측 계약을 분리했다.
- misconception repair:
  - `REQUIRES_NEW`가 phantom을 막는다는 오해
  - `READ COMMITTED`면 deadlock이 없다는 오해
  - race condition과 phantom read가 같은 말이라는 오해
- 품질 기준 exemplar: `database-deep-dive` corpus의 기존 08번 구조와 `study-explanation`의 state trace 규칙.

## 4. Project Overlay

- 적용한 로컬 AGENTS 경로:
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - `/Users/rody/VscodeProjects/study/interviews/PROJECT_INTENT.md`
  - `/Users/rody/VscodeProjects/study/interviews/USECASE.md`
- 활성화한 프로젝트 규칙:
  - 정식 답변 자산은 짧은 직답과 깊은 메커니즘을 함께 가져야 한다.
  - root `interviews/`는 curated learning surface이고 `source/`는 raw reservoir다.
  - 기존 평균이 아니라 exemplar 수준으로 끌어올린다.
- multi-agent fallback: 현재 사용자가 명시적으로 sub-agent를 요청하지 않았으므로 single-agent + separated audit lanes로 처리했다.

## 5. Evidence Ledger

- E-01: Spring 공식 Transaction Propagation 문서
  - 자료: `https://docs.spring.io/spring-framework/reference/data-access/transaction/declarative/tx-propagation.html`
  - 닫은 주장: `REQUIRED`는 기존 물리 transaction 참여, `REQUIRES_NEW`는 독립 물리 transaction, `NESTED`는 savepoint 성격이라는 경계 claim.
- E-02: Spring 공식 `@Transactional` 문서
  - 자료: `https://docs.spring.io/spring-framework/reference/data-access/transaction/declarative/annotations.html`
  - 닫은 주장: 기본 propagation/isolation 속성과 isolation/readOnly/timeout 적용 경계.
- E-03: PostgreSQL transaction isolation 문서
  - 자료: `https://www.postgresql.org/docs/18/transaction-iso.html`
  - 닫은 주장: PostgreSQL에서 `READ COMMITTED`, `REPEATABLE READ`, `SERIALIZABLE`, dirty/non-repeatable/phantom/serialization anomaly의 제품별 경계.
- E-04: MySQL 8.4 InnoDB transaction isolation / locking / deadlock 문서
  - 자료: `https://dev.mysql.com/doc/refman/8.4/en/innodb-transaction-isolation-levels.html`, `https://dev.mysql.com/doc/refman/8.4/en/innodb-locking.html`, `https://dev.mysql.com/doc/refman/8.4/en/innodb-deadlocks.html`
  - 닫은 주장: InnoDB default `REPEATABLE READ`, locking read에서 gap/next-key lock, deadlock 처리와 retry 필요성.

## 6. Fixed Checklist

- [x] `pdirty read`는 표준 용어가 아니라 `dirty read` 맥락으로 정리한다.
- [x] propagation과 isolation을 같은 축으로 섞지 않는다.
- [x] `READ COMMITTED`, `REPEATABLE READ`, phantom/dirty read를 제품별 차이와 함께 설명한다.
- [x] lock/deadlock/race condition을 각각 다른 실패/관측 축으로 분리한다.
- [x] Spring propagation replay를 추가한다.
- [x] README, registry, claim/evidence/composition audit를 동기화한다.
- [x] validator와 path-limited diff check를 통과한다.
- [x] unrelated dirty worktree는 건드리지 않는다.

## 7. Critic Review

- Definition review: 단순 용어집을 만드는 요청이 아니라, existing canonical 08번 문서를 사용자가 든 개념까지 포괄하도록 보강하는 작업으로 정의했다. 새 파일을 만들면 validator registry 계약과 canonical 위치가 분산되므로 기존 08번 보강이 더 맞다.
- Criteria review: 본문만 바꾸면 audit/registry drift가 생긴다. 따라서 README, topic registry, claim/evidence/composition audit를 함께 고정했다.
- Checklist review: propagation, isolation, dirty/phantom/repeatable, lock/deadlock/race condition을 모두 포함했고, `pdirty read`는 dirty read로 해석했다.
- Result review: validator PASS, diff-check PASS, 공식 source 링크와 audit ref 추가를 확인했다.

## 8. Verification Log

- `python3 interviews/database-deep-dive/tools/validate_interview_database_deep_dive.py`
  - 결과: `PASS: interviews database deep-dive structural validation`
- `git diff --check -- interviews/database-deep-dive/08-isolation-lock-deadlock.md interviews/database-deep-dive/README.md interviews/database-deep-dive/topic-registry.tsv interviews/database-deep-dive/audit/claim-audit.tsv interviews/database-deep-dive/audit/composition-audit.tsv interviews/database-deep-dive/audit/evidence-refs.tsv docs/works/WORK_20260520_DB_TRANSACTION_CONCURRENCY_MAP.md`
  - 결과: PASS

## 9. Final Closure

- requested closure scope: 사용자가 언급한 transaction propagation, isolation level, dirty/phantom/repeatable read, lock, deadlock, race condition을 정식 인터뷰 학습 문서로 정리한다.
- achieved closure scope: canonical 08번 문서와 관련 index/audit 표면까지 동기화했다.
- registry source: `interviews/database-deep-dive/topic-registry.tsv`
- work stage-tranche id: `DBI-LOCK-01-propagation-race-extension`
- remaining open tranche disclosure: 없음.
- whole-request completion verdict: `WHOLE_COMPLETE`
