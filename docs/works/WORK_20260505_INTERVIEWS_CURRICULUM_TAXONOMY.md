# WORK_20260505_INTERVIEWS_CURRICULUM_TAXONOMY

## 0. Meta

- 작업 제목: 인터뷰 원재료 10대 대주제 curriculum 재배치
- WORK 파일 경로: `docs/works/WORK_20260505_INTERVIEWS_CURRICULUM_TAXONOMY.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `execute | refactor_docs`
- 작업 깊이: `full`
- 원문 사용자 요청: "데이터베이스와 저장소에 검색과 NoSQL를 포함시키고 싶습니다. ... 문서를 만들고, 대분류, 중분류, 소분류하고 원문 내용들을 분석해서 각 위치에 갖다 놔 주세요."
- 대상 경로 / 자산: `interviews/intervie*.md`, `interviews/curriculum/`, `interviews/tools/build_interview_curriculum.py`
- 현재 상태: `COMPLETE`
- 완료 게이트: `PASS`
- finish: `verify+commit`

## 1. Request Normalization

- goal: 원문 interview source를 10개 내외의 최종 학습 대분류로 다시 배치한다.
- refs:
  - `interviews/PROJECT_INTENT.md`
  - `interviews/USECASE.md`
  - `interviews/tools/split_interview_sources.py`
  - previous migration commit `7fc479d`
- scope:
  - source files: `interviews/interview_questions.md`, `interviews/interview_questions2.md`, `interviews/interview_questions3.md`, `interviews/interview_questions4.md`, `interviews/interview_s4.md`, `interviews/interviews.md`, `interviews/interviews2.md`
  - output: `interviews/curriculum/`
- mode: `execute`
- run_mode: `normal`
- finish: `verify+commit`
- must_keep:
  - `검색/NoSQL`은 `데이터베이스, 저장소, 검색/NoSQL` 대분류 안에 둔다.
  - 대주제 문서는 10개로 유지한다.
  - 각 문서 안에서 중분류와 소분류를 둔다.
  - 원문 source chunk는 의미를 바꾸지 않고 배치한다.
  - source span, SHA-256, duplicate alias를 남겨 원문 추적성을 유지한다.
- extra_checks:
  - 7개 source file 전체가 curriculum manifest에 잡혀야 한다.
  - 241개 source chunk와 133개 unique chunk가 모두 배치되어야 한다.
  - fallback 분류가 남으면 안 된다.

## 2. Root-First Framing

- 근본 문제: 16개 이상의 topic staging 문서는 원문 보존에는 유용하지만, 사람이 인터뷰 준비 경로로 따라가기에는 문서 수가 많다.
- 왜 중요한가: 인터뷰 준비 자산은 빠르게 찾아 짧게 답하고, 이어서 깊게 내려갈 수 있어야 한다.
- 작업 목표: 10개 대분류를 canonical curriculum으로 만들고, 중분류와 소분류 아래에 원문 chunk를 배치한다.
- 성공 정의: 모든 source chunk가 10개 대분류 또는 source-context appendix 중 하나에 배치되고, manifest 검증이 PASS한다.
- PARTIAL 조건: curriculum 문서는 생성됐지만 coverage 검증이나 commit이 닫히지 않은 경우.
- BLOCKED 조건: source 파일 접근, 생성, 검증, commit 중 하나가 불가능한 경우.

## 3. Project Overlay

- 로컬 AGENTS 경로: `AGENTS.md`
- 활성화한 프로젝트 규칙:
  - `interviews`는 경력 기술 인터뷰 준비 subproject다.
  - 원재료는 정식 답변 자산으로 바로 승격하지 않고, 질문 cluster와 개념 cluster를 분리한다.
  - 좋은 답변은 `짧은 직답 -> 깊은 메커니즘 -> 예시 -> 꼬리 질문 -> 검증/근거`로 승격되어야 한다.
- 전역 규칙과 충돌: 없음.

## 4. Frozen Checklist

- [x] 대주제 문서를 10개로 고정한다.
- [x] `검색/NoSQL`은 `데이터베이스, 저장소, 검색/NoSQL` 안에 포함한다.
- [x] 모든 원문 unique chunk를 대분류/중분류/소분류 위치에 배치한다.
- [x] 기술 대주제에 직접 들어가지 않는 질문 은행과 시나리오는 별도 source-context appendix에 둔다.
- [x] source span, SHA-256, duplicate alias를 manifest와 문서에 남긴다.
- [x] fallback 분류가 0개인지 확인한다.
- [x] coverage 검증으로 source chunk 241개, unique chunk 133개가 모두 닫히는지 확인한다.
- [x] 이번 작업 파일만 commit한다.

## 5. Evidence / Decision Ledger

- D-01: final curriculum은 10개 대주제 문서로 둔다.
  - support tier: `T1 Direct Evidence`
  - evidence: 사용자 요청의 "대주제 10개 내외", "문서가 또 너무 많으면 안 좋다".
  - counterexample: staging topic file을 최종 학습 구조로 삼으면 문서 탐색 비용이 다시 커진다.
  - admission lane: `APPLY`
- D-02: `검색/NoSQL`은 DB 문서 안으로 합친다.
  - support tier: `T1 Direct Evidence`
  - evidence: 사용자 요청의 "`데이터베이스와 저장소`에 `검색과 NoSQL`를 포함".
  - counterexample: 별도 검색 문서를 두면 저장소 선택, 인덱싱, 샤딩, 조회 성능 판단이 분리된다.
  - admission lane: `APPLY`
- D-03: 질문 은행과 면접 시나리오는 10개 기술 대주제 밖의 appendix로 둔다.
  - support tier: `T2 Strong Inference`
  - evidence: 해당 chunk는 기술 본문이라기보다 질문 목록, 이력서 기반 시나리오, source front matter다.
  - counterexample: 억지로 기술 대주제 하나에 넣으면 source context가 기술 설명처럼 오해될 수 있다.
  - admission lane: `APPLY`
- D-04: curriculum 문서에서는 heading depth만 조정한다.
  - support tier: `T2 Strong Inference`
  - evidence: 대분류/중분류/소분류 계층 안에 원문 heading을 그대로 넣으면 Markdown 구조가 깨진다.
  - counterexample: 문장 자체를 rewrite하면 아직 딥 리라이트 전 단계에서 원문 의미가 변할 수 있다.
  - admission lane: `APPLY`

## 6. Scope / Impact

- created:
  - `interviews/tools/build_interview_curriculum.py`
  - `interviews/curriculum/README.md`
  - `interviews/curriculum/_question-index.md`
  - `interviews/curriculum/_curriculum_manifest.json`
  - `interviews/curriculum/*.md`
- updated:
  - `interviews/README.md`
- replaced as primary navigation:
  - `interviews/topics/` staging docs are superseded by `interviews/curriculum/`.
- excluded:
  - deep rewrite of each answer
  - official-source fact checking for each topic
  - deletion or rewriting of raw `intervie*.md` source files

## 7. Critique + Repair

- Challenge: broad question-bank chunks do not belong cleanly to one technical 대주제.
  - Repair: keep them in `_source-context-and-question-bank.md` and use `_question-index.md` as the later promotion bridge.
- Challenge: keyword overlap can misplace chunks, such as `Deadlock` matching DB `Lock` or `BLOCKED` matching `lock`.
  - Repair: high-risk terms were handled with earlier specific rules before generic DB/Spring/network rules.
- Challenge: keeping both `topics/` and `curriculum/` can leave too many visible docs.
  - Repair: `curriculum/` becomes the primary structure, and the prior topic staging docs are removed from the active file tree.

## 8. Verification Plan

- `python3 interviews/tools/build_interview_curriculum.py`
- manifest verification:
  - source files = `7`
  - source chunks = `241`
  - unique chunks = `133`
  - major count = `10`
  - fallback placements = `0`
  - every unique SHA appears in exactly one curriculum chunk marker
  - deduplicated source aliases match the source chunks
- `python3 -m py_compile interviews/tools/build_interview_curriculum.py`
- path-limited `git diff --check`

## 9. Execution Log

- generated:
  - `interviews/curriculum/README.md`
  - `interviews/curriculum/_question-index.md`
  - `interviews/curriculum/_source-context-and-question-bank.md`
  - `interviews/curriculum/_curriculum_manifest.json`
  - `interviews/curriculum/01-language-runtime.md`
  - `interviews/curriculum/02-concurrency-async-io.md`
  - `interviews/curriculum/03-os-kernel-computer-architecture.md`
  - `interviews/curriculum/04-network-web-protocols.md`
  - `interviews/curriculum/05-security-cryptography.md`
  - `interviews/curriculum/06-database-storage-search-nosql.md`
  - `interviews/curriculum/07-messaging-event-driven.md`
  - `interviews/curriculum/08-distributed-systems-architecture.md`
  - `interviews/curriculum/09-spring-backend-frameworks.md`
  - `interviews/curriculum/10-problem-solving-code-quality.md`
- generated helper:
  - `interviews/tools/build_interview_curriculum.py`
- updated:
  - `interviews/README.md` now points to `curriculum/` as the main learning entry.
- removed active staging docs:
  - `interviews/topics/`
- final counts:
  - source chunks: `241`
  - unique chunks: `133`
  - major documents: `10`
  - deduplicated groups: `108`
- major counts:
  - `01-language-runtime.md`: `9`
  - `02-concurrency-async-io.md`: `38`
  - `03-os-kernel-computer-architecture.md`: `10`
  - `04-network-web-protocols.md`: `8`
  - `05-security-cryptography.md`: `6`
  - `06-database-storage-search-nosql.md`: `17`
  - `07-messaging-event-driven.md`: `5`
  - `08-distributed-systems-architecture.md`: `5`
  - `09-spring-backend-frameworks.md`: `22`
  - `10-problem-solving-code-quality.md`: `5`
  - `_source-context-and-question-bank.md`: `8`
- repair during classification:
  - `Race Condition`, `Deadlock`, `epoll`, `File Descriptor`, `infix/inline`, `Spring JDBC`, wait/notify scenario chunks were checked and moved away from misleading broad keyword matches.

## 10. Verification Log

- `python3 interviews/tools/build_interview_curriculum.py`: `PASS`
- manifest verification script:
  - source files = `7 PASS`
  - source chunks = `241 PASS`
  - unique chunks = `133 PASS`
  - major count = `10 PASS`
  - deduplicated groups = `108 PASS`
  - source file SHA-256 hashes match manifest: `PASS`
  - every source chunk has a representative mapping: `PASS`
  - every unique SHA appears in exactly one curriculum marker: `PASS`
  - fallback placements = `0 PASS`
  - old `interviews/topics/` directory removed: `PASS`
- `python3 -m py_compile interviews/tools/build_interview_curriculum.py interviews/tools/split_interview_sources.py`: `PASS`
- `git diff --check -- interviews/README.md interviews/tools/build_interview_curriculum.py docs/works/WORK_20260505_INTERVIEWS_CURRICULUM_TAXONOMY.md interviews/curriculum`: `PASS`

## 11. Final Audit

- requested closure scope: 10개 내외 대주제 curriculum을 만들고, 그 안에 중분류/소분류를 두며, 원문 내용을 각 위치에 배치한다.
- achieved closure scope: 10개 기술 대주제와 1개 source-context appendix를 만들고, 133개 unique source chunk를 모두 배치했다.
- user-specific decision: `검색/NoSQL`은 독립 대분류가 아니라 `06. 데이터베이스, 저장소, 검색/NoSQL`에 포함했다.
- source preservation: raw `intervie*.md` 파일은 삭제하거나 변경하지 않았다. curriculum 문서에서는 heading depth만 조정했고, source span과 SHA-256을 남겼다.
- downstream impact: 이후 deep rewrite는 `interviews/curriculum/`을 기본 진입점으로 삼는다. raw source와 manifest가 남아 있어 배치 오류가 발견되면 재생성 또는 수동 이동으로 되돌릴 수 있다.
- remaining open tranche disclosure:
  - 이번 work는 `curriculum taxonomy + source placement` tranche만 닫는다.
  - 다음 tranche는 각 소주제를 실제 면접 답변 자산으로 승격하는 deep rewrite다.
- whole-request objective status: `COMPLETE_FOR_THIS_CURRICULUM_PLACEMENT_TRANCHE`, `PARTIAL_FOR_LARGER_DEEP_REWRITE_OBJECTIVE`.
