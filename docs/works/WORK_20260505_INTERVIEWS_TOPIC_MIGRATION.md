# WORK_20260505_INTERVIEWS_TOPIC_MIGRATION

## 0. Meta

- 작업 제목: interview 원재료 주제별 원문 보존 이동
- WORK 파일 경로: `docs/works/WORK_20260505_INTERVIEWS_TOPIC_MIGRATION.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `execute | refactor_docs`
- 작업 깊이: `full`
- 원문 사용자 요청: "`intervie*.md` 라고 되어 있는 것들의 내용을 분석하고 분류하여 주제별로 문서를 만들고, 우선은 각 내용을 해치지 않으면서 새로운 파일들을 만들어서 주제별로 이동시켜야 합니다. 가능한가요?"
- 대상 경로 / 자산: `interviews/intervie*.md`, `interviews/topics/`, `interviews/tools/split_interview_sources.py`
- 현재 상태: `COMPLETE`
- 완료 게이트: `PASS`
- finish: `verify+commit`

## 1. Request Normalization

- goal: `intervie*.md` 원재료를 손상 없이 주제별 새 문서로 이동할 수 있는 기반을 만든다.
- refs: `interviews/PROJECT_INTENT.md`, `interviews/USECASE.md`, `interviews/README.md`
- scope: `interviews/interview_questions.md`, `interviews/interview_questions2.md`, `interviews/interview_questions3.md`, `interviews/interview_questions4.md`, `interviews/interview_s4.md`, `interviews/interviews.md`, `interviews/interviews2.md`
- mode: `execute`
- run_mode: `normal`
- finish: `verify+commit`
- must_keep:
  - 원문 내용은 재작성하지 않고 source chunk로 그대로 보존한다.
  - exact duplicate는 본문을 한 번만 두되 source alias를 manifest에 남긴다.
  - 원본 파일은 이번 단계에서 삭제하지 않는다.
  - 딥 리라이트는 다음 단계로 남기고, 이번 단계는 보존형 주제별 staging을 닫는다.
- extra_checks:
  - 7개 source file 전체가 manifest에 잡혀야 한다.
  - source chunk count와 unique chunk count가 검증 가능해야 한다.
  - topic file에는 source span과 SHA-256이 남아야 한다.

## 2. Root-First Framing

- 근본 문제: 현재 큰 interview raw 파일들은 주제별 딥 학습 문서로 승격하기에는 너무 크고 섞여 있다.
- 왜 중요한가: 딥 리라이트 전에 내용을 잃거나 왜곡하면 이후 설명 품질을 끌어올릴 수 없다.
- 작업 목표: 원문을 해치지 않는 topic staging area와 재생성 가능한 split script를 만든다.
- 성공 정의: 모든 source chunk가 manifest에 있고, unique chunk가 topic files에 포함되며, verification script로 coverage가 확인된다.
- PARTIAL 조건: topic file은 생성됐지만 coverage 검증이나 commit이 닫히지 않은 경우.
- BLOCKED 조건: source file 접근, 파일 생성, 검증, 또는 commit이 불가능한 경우.

## 3. Project Overlay

- 로컬 AGENTS 경로: `AGENTS.md`
- 활성화한 프로젝트 규칙:
  - `interviews/interview_questions*.md` 같은 대형 수집 문서는 source reservoir로 본다.
  - 정식 문서로 바로 다듬기보다 질문 cluster와 개념 cluster를 분리해 promotion한다.
  - 현재 평균 품질이 아니라 최고 수준의 학습 문서로 끌어올리는 것이 목표다.
- `interviews` overlay:
  - 짧은 직답과 깊은 메커니즘 설명을 동시에 지원해야 한다.
  - 작은 기술 단위가 시스템 아키텍처 설명으로 조립되어야 한다.
- 전역 규칙과 충돌: 없음.

## 4. Frozen Checklist

- [x] `intervie*.md` 7개 파일을 전부 inventory에 포함한다.
- [x] 원문 section/chunk 경계를 source span으로 기록한다.
- [x] exact duplicate는 제거하되 duplicate source alias를 남긴다.
- [x] 새 topic files는 source chunk 본문을 재작성하지 않는다.
- [x] topic index와 migration manifest를 생성한다.
- [x] verification으로 source chunk coverage와 topic file inclusion을 확인한다.
- [x] 기존 원본 파일을 삭제하거나 직접 변경하지 않는다.
- [x] 이번 작업 파일만 commit한다.

## 5. Evidence / Decision Ledger

- D-01: 이번 단계는 딥 리라이트가 아니라 원문 보존형 주제별 staging으로 닫는다.
  - support tier: `T1 Direct Evidence`
  - evidence: 사용자 요청의 "우선은 각 내용을 해치지 않으면서" 문구와 `interviews/USECASE.md`의 승격 대기소 모델.
  - counterexample: 바로 딥 리라이트하면 원문 의미를 잃거나 출처 추적이 끊길 수 있다.
  - admission lane: `APPLY`
- D-02: 원본 파일 삭제 없이 topic files를 생성한다.
  - support tier: `T2 Strong Inference`
  - evidence: 원문 보존과 dirty worktree 상태를 함께 고려하면 copy-first가 가장 되돌리기 쉽다.
  - counterexample: 실제 move/delete는 원본과 새 파일 간 coverage 검증이 충분히 누적된 뒤에 해야 안전하다.
  - admission lane: `APPLY`
- D-03: exact duplicate chunk는 본문 중복 대신 alias로 보존한다.
  - support tier: `T1 Direct Evidence`
  - evidence: `interview_questions2.md`와 `interviews2.md`는 동일 SHA-256 파일이다.
  - counterexample: 중복 본문을 두 번 붙이면 읽기와 후속 리라이트 비용만 늘어난다.
  - admission lane: `APPLY`

## 6. Scope / Impact

- source files:
  - `interview_questions.md`
  - `interview_questions2.md`
  - `interview_questions3.md`
  - `interview_questions4.md`
  - `interview_s4.md`
  - `interviews.md`
  - `interviews2.md`
- observed line count: 28,692 total source lines.
- together-moving surfaces:
  - topic docs
  - generator script
  - migration manifest
  - WORK ledger
- excluded:
  - source raw file deletion
  - official-source fact checking
  - deep rewritten answer docs

## 7. Critique + Repair

- Challenge: 자동 분류가 완벽하지 않으면 잘못된 topic file에 들어갈 수 있다.
  - Repair: 이번 단계는 source-preserving staging이므로 source span과 manifest를 남기고, 잘못 분류된 chunk는 후속 단계에서 손실 없이 이동 가능하게 한다.
- Challenge: duplicate를 제거하면 특정 파일의 내용이 빠졌다고 오해할 수 있다.
  - Repair: `_migration_manifest.json`의 모든 source chunk와 duplicate source alias에 각 원본 span을 남긴다.
- Challenge: "이동"이라는 말 때문에 원본 삭제까지 수행할 수 있다.
  - Repair: 이번 단계는 원본 삭제 없는 copy-first 이동으로 제한한다. 삭제는 coverage 검증과 사용 확인 이후 별도 작업으로 둔다.

## 8. Verification Plan

- `python3 interviews/tools/split_interview_sources.py`
- manifest 검증:
  - source files = 7
  - source chunks = parsed source chunk count
  - unique chunks = deduplicated unique chunk count
  - every representative chunk hash appears in exactly one topic file
  - every source chunk has representative ref
- hygiene:
  - `git diff --check -- interviews/tools/split_interview_sources.py interviews/topics docs/works/WORK_20260505_INTERVIEWS_TOPIC_MIGRATION.md`

## 9. Execution Log

- inventory:
  - source file count: `7`
  - source line count: `28,692`
  - exact duplicate file hash: `interview_questions2.md` and `interviews2.md`
- generated assets:
  - `interviews/tools/split_interview_sources.py`
  - `interviews/topics/README.md`
  - `interviews/topics/_migration_manifest.json`
  - `interviews/topics/*.md` topic staging files
- generator result:
  - source chunks: `241`
  - unique chunks: `133`
  - deduplicated groups: `108`
- topic counts:
  - `00-source-front-matter-and-question-banks`: `9`
  - `algorithms-and-data-structures`: `1`
  - `concurrency-async-io`: `34`
  - `containers-and-devops`: `1`
  - `database-and-storage`: `8`
  - `distributed-systems-and-architecture`: `4`
  - `functional-programming`: `1`
  - `jvm-java-kotlin-runtime`: `11`
  - `language-runtimes-go-node-php`: `3`
  - `messaging-and-streaming`: `4`
  - `network-and-web-protocols`: `8`
  - `os-kernel-computer-architecture`: `9`
  - `search-and-nosql`: `6`
  - `security-and-cryptography`: `8`
  - `spring-and-frameworks`: `24`
  - `testing-and-code-design`: `2`
- repair during verification:
  - first generated pass left stale `miscellaneous.md` from an earlier classifier result.
  - generator now clears its known generated outputs before writing new topic files.
  - final representative chunk count in `miscellaneous`: `0`.

## 10. Verification Log

- `python3 interviews/tools/split_interview_sources.py`: `PASS`
- manifest verification script:
  - source files match source list: `PASS`
  - source chunk count: `241 PASS`
  - unique chunk count: `133 PASS`
  - current source file SHA-256 hashes match manifest: `PASS`
  - every source span maps to a representative chunk: `PASS`
  - every unique representative SHA-256 marker appears in exactly one topic file: `PASS`
  - deduplicated source aliases match manifest: `PASS`
- `python3 -m py_compile interviews/tools/split_interview_sources.py`: `PASS`
- `git diff --check -- interviews/tools/split_interview_sources.py docs/works/WORK_20260505_INTERVIEWS_TOPIC_MIGRATION.md`: `PASS`

## 11. Final Audit

- requested closure scope: `intervie*.md` 원재료를 해치지 않고 주제별 새 파일로 이동할 수 있는 첫 보존 단계.
- achieved closure scope: 7개 source file, 241개 source chunk, 133개 unique chunk를 topic staging files와 manifest로 보존했다.
- source preservation: 원본 source file은 삭제하거나 직접 수정하지 않았다.
- topic classification status: 모든 대표 chunk가 explicit topic file에 들어갔고 `miscellaneous` 대표 chunk는 남기지 않았다.
- downstream impact: 새 topic files는 다음 딥 리라이트의 입력이다. 되돌림은 generated topic files와 manifest를 제거하거나 generator를 재실행하는 방식으로 가능하다.
- remaining open tranche disclosure:
  - 이번 work는 `source-preserving topic migration` tranche만 닫는다.
  - 다음 tranche는 각 topic file을 기준으로 공식 문서, 표준, 예제, 실패 모드, 검증 경로를 붙인 deep study interview answer 문서로 승격하는 작업이다.
- whole-request objective status: `PARTIAL_FOR_LARGER_INTERVIEW_REWRITE`, `COMPLETE_FOR_THIS_MIGRATION_TRANCHE`.
