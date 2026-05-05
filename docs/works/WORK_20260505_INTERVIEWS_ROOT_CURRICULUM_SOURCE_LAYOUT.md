# WORK_20260505_INTERVIEWS_ROOT_CURRICULUM_SOURCE_LAYOUT

## 0. Meta

- 작업 제목: interviews root curriculum 배치와 source reservoir 분리
- WORK 파일 경로: `docs/works/WORK_20260505_INTERVIEWS_ROOT_CURRICULUM_SOURCE_LAYOUT.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `execute | refactor_docs`
- 작업 깊이: `standard`
- 원문 사용자 요청: "커리큘럼 문서들을 interviews 바로 아래에 두고, source 문서들을 source 디렉토리로 옮겨 주세요. 즉 기존 문서가 아니라 앞으로 새로운 문서 기반으로 정리하고 학습할 예정입니다."
- 대상 경로 / 자산: `interviews/`, `interviews/source/`, `interviews/tools/build_interview_curriculum.py`, `interviews/tools/split_interview_sources.py`
- 현재 상태: `COMPLETE`
- 완료 게이트: `PASS`
- finish: `verify+commit`

## 1. Request Normalization

- goal: 앞으로의 학습 표면을 `interviews/` root의 새 대주제 문서로 만들고, 기존 원재료 문서는 `interviews/source/`로 분리한다.
- refs:
  - `interviews/PROJECT_INTENT.md`
  - `interviews/USECASE.md`
  - previous curriculum commit `ee8df5e`
- scope:
  - move root-facing curriculum docs out of `interviews/curriculum/`
  - move raw `intervie*.md` source docs into `interviews/source/`
  - update regeneration scripts and README links
- mode: `execute`
- run_mode: `normal`
- finish: `verify+commit`
- must_keep:
  - root curriculum docs are the new primary learning surface.
  - raw source docs remain preserved under `source/`.
  - source refs and manifest must point to `source/...`.
  - no unrelated dirty files are touched.
- extra_checks:
  - `interviews/curriculum/` must be gone.
  - `interviews/source/` must contain the 7 raw source files.
  - source chunk count remains `241`, unique chunk count remains `133`, major count remains `10`.

## 2. Root-First Framing

- 근본 문제: `interviews/curriculum/` 하위에 새 문서가 있으면 사용자가 또 하위 폴더를 찾아야 하고, root에는 기존 source 파일들이 남아 새 학습 표면과 원재료 표면이 섞인다.
- 작업 목표: `interviews/` root를 새 문서 기반 학습 표면으로 만들고, 원재료는 `source/`로 명확히 분리한다.
- 성공 정의: root에 10개 대주제 문서와 index/manifest가 있고, raw sources는 `source/`에 있으며, 재생성/검증이 통과한다.
- PARTIAL 조건: 파일 이동은 됐지만 생성 스크립트나 검증이 새 경로와 맞지 않는 경우.
- BLOCKED 조건: 파일 이동, 검증, commit 중 하나가 불가능한 경우.

## 3. Frozen Checklist

- [x] 10개 curriculum 문서를 `interviews/` 바로 아래로 이동한다.
- [x] `_question-index.md`와 `_curriculum_manifest.json`을 `interviews/` 바로 아래에 둔다.
- [x] raw `intervie*.md` 7개 파일을 `interviews/source/`로 이동한다.
- [x] source-context appendix를 `interviews/source/`로 이동한다.
- [x] `interviews/curriculum/` 디렉터리를 제거한다.
- [x] README와 source README가 새 구조를 설명한다.
- [x] 생성 스크립트가 `source/`를 읽고 root 문서를 다시 생성한다.
- [x] 검증에서 source chunk `241`, unique chunk `133`, major count `10`이 유지된다.
- [x] 이번 작업 파일만 commit한다.

## 4. Decisions

- D-01: root에는 새 학습 문서를 두고, source에는 원재료만 둔다.
  - support tier: `T1 Direct Evidence`
  - evidence: 사용자 요청의 "커리큘럼 문서들을 interviews 바로 아래", "source 문서들을 source 디렉토리", "앞으로 새로운 문서 기반".
  - admission lane: `APPLY`
- D-02: root `README.md`는 새 학습 표면의 index 역할을 겸한다.
  - support tier: `T2 Strong Inference`
  - evidence: root에 별도 `curriculum/README.md`를 둘 수 없고, root 진입점은 하나여야 탐색 비용이 낮다.
  - admission lane: `APPLY`
- D-03: raw source refs는 `source/<file>:line-line` 형식으로 바꾼다.
  - support tier: `T1 Direct Evidence`
  - evidence: source files moved under `interviews/source/`.
  - admission lane: `APPLY`

## 5. Verification Plan

- `python3 interviews/tools/build_interview_curriculum.py`
- manifest verification:
  - raw source paths under `interviews/source/`
  - source files = `7`
  - source chunks = `241`
  - unique chunks = `133`
  - major count = `10`
  - every unique SHA marker appears once in root docs or source-context appendix
  - no `interviews/curriculum/` directory remains
- `python3 -m py_compile interviews/tools/build_interview_curriculum.py interviews/tools/split_interview_sources.py`
- path-limited `git diff --check`

## 6. Execution Log

- moved root learning docs:
  - `interviews/01-language-runtime.md`
  - `interviews/02-concurrency-async-io.md`
  - `interviews/03-os-kernel-computer-architecture.md`
  - `interviews/04-network-web-protocols.md`
  - `interviews/05-security-cryptography.md`
  - `interviews/06-database-storage-search-nosql.md`
  - `interviews/07-messaging-event-driven.md`
  - `interviews/08-distributed-systems-architecture.md`
  - `interviews/09-spring-backend-frameworks.md`
  - `interviews/10-problem-solving-code-quality.md`
- moved root indexes:
  - `interviews/_question-index.md`
  - `interviews/_curriculum_manifest.json`
- moved source reservoir:
  - `interviews/source/interview_questions.md`
  - `interviews/source/interview_questions2.md`
  - `interviews/source/interview_questions3.md`
  - `interviews/source/interview_questions4.md`
  - `interviews/source/interview_s4.md`
  - `interviews/source/interviews.md`
  - `interviews/source/interviews2.md`
  - `interviews/source/_source-context-and-question-bank.md`
- added:
  - `interviews/source/README.md`
- updated:
  - `interviews/README.md`
  - `interviews/tools/build_interview_curriculum.py`
  - `interviews/tools/split_interview_sources.py`
- removed:
  - `interviews/curriculum/`

## 7. Verification Log

- `python3 interviews/tools/build_interview_curriculum.py`: `PASS`
- manifest/path verification:
  - source files under `interviews/source/`: `7 PASS`
  - raw source files left in `interviews/` root: `0 PASS`
  - source chunks: `241 PASS`
  - unique chunks: `133 PASS`
  - major count: `10 PASS`
  - deduplicated groups: `108 PASS`
  - every source ref uses `source/` prefix: `PASS`
  - every unique SHA appears in exactly one root curriculum marker or source-context marker: `PASS`
  - `interviews/curriculum/` removed: `PASS`
- `python3 -m py_compile interviews/tools/build_interview_curriculum.py interviews/tools/split_interview_sources.py`: `PASS`
- `git diff --check -- interviews docs/works/WORK_20260505_INTERVIEWS_ROOT_CURRICULUM_SOURCE_LAYOUT.md`: `PASS`

## 8. Final Audit

- requested closure scope: curriculum docs directly under `interviews/`, source docs under `interviews/source/`, future learning based on new docs.
- achieved closure scope: root learning surface now contains the 10 major documents plus index/manifest, and all raw source files moved to `source/`.
- source preservation: raw source file content was moved, not rewritten.
- regeneration path: `python3 interviews/tools/build_interview_curriculum.py` reads `interviews/source/` and writes the root learning documents.
- unrelated open work: existing unrelated dirty/staged files outside `interviews` remain untouched.
- whole-request objective status: `COMPLETE_FOR_THIS_LAYOUT_REORGANIZATION_TRANCHE`.
