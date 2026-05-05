# WORK_20260505_INTERVIEWS_UNNUMBERED_CURRICULUM_FILES

## 0. Meta

- 작업 제목: interviews curriculum 파일명과 제목의 숫자 접두사 제거
- WORK 파일 경로: `docs/works/WORK_20260505_INTERVIEWS_UNNUMBERED_CURRICULUM_FILES.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `execute | refactor_docs`
- 작업 깊이: `standard`
- 원문 사용자 요청: "네 진행해 주세요."
- 직전 맥락: curriculum 문서의 앞 번호가 꼭 필요한지 검토한 뒤, 파일명과 제목에서 숫자 접두사를 제거하는 방향으로 진행하기로 했다.
- 대상 경로 / 자산: `interviews/`, `interviews/tools/build_interview_curriculum.py`, `interviews/_curriculum_manifest.json`, `interviews/_question-index.md`
- 현재 상태: `COMPLETE`
- 완료 게이트: `PASS`
- finish: `verify+commit`

## 1. Request Normalization

- goal: root curriculum 문서가 사람이 읽는 대주제 이름 그대로 보이게 하고, 정렬용 번호를 파일명과 제목에서 제거한다.
- refs:
  - `interviews/PROJECT_INTENT.md`
  - `interviews/USECASE.md`
  - `docs/works/WORK_20260505_INTERVIEWS_ROOT_CURRICULUM_SOURCE_LAYOUT.md`
- scope:
  - 10개 curriculum 파일명을 unnumbered slug로 재생성한다.
  - 대주제 제목에서 `01.` 같은 번호를 제거한다.
  - README, question index, manifest가 새 파일명과 제목을 가리키게 한다.
  - 생성기가 레거시 numbered 파일을 정리하도록 한다.
- run_mode: `normal`
- must_keep:
  - 대주제 10개 체계는 유지한다.
  - source reservoir는 `interviews/source/`에 둔다.
  - source chunk count `241`, unique chunk count `133`, major count `10`은 유지한다.
  - unrelated dirty files는 건드리지 않는다.

## 2. Root-First Framing

- 근본 문제: `01-...`, `01.` 표기는 정렬에는 편하지만 문서의 의미를 설명하지 않는다. 이 프로젝트에서는 번호 순서보다 대주제 이름 자체가 학습 표면의 기본 단서가 되어야 한다.
- 작업 목표: 내부 생성 순서는 유지하되, 사람이 보는 파일명과 제목에서는 숫자를 제거한다.
- 성공 정의: root에는 unnumbered 10개 curriculum 문서만 남고, 생성기/README/index/manifest가 모두 같은 이름 체계를 사용한다.
- PARTIAL 조건: 파일명은 바뀌었지만 생성기나 manifest가 numbered 이름을 계속 쓰는 경우.
- BLOCKED 조건: 재생성 또는 검증이 실패해 source chunk 보존 여부를 확인할 수 없는 경우.

## 3. Frozen Checklist

- [x] 10개 root curriculum 파일을 numbered 이름에서 unnumbered 이름으로 전환한다.
- [x] 각 curriculum 문서의 H1 제목에서 `NN.` 접두사를 제거한다.
- [x] `build_interview_curriculum.py`의 `MAJORS` 파일명과 제목을 unnumbered 기준으로 수정한다.
- [x] 생성기가 이전 numbered 파일을 cleanup하도록 legacy list를 둔다.
- [x] README, `_question-index.md`, `_curriculum_manifest.json`을 재생성해 새 이름과 링크를 반영한다.
- [x] `interviews/source/` raw source reservoir는 유지한다.
- [x] source chunk `241`, unique chunk `133`, major count `10`이 유지되는지 검증한다.
- [x] 각 unique SHA marker가 curriculum 또는 appendix에 정확히 한 번 존재하는지 검증한다.
- [x] fallback placement가 없는지 검증한다.
- [x] 이번 작업 파일만 commit한다.

## 4. Decisions

- D-01: 숫자 접두사는 사람이 읽는 curriculum 파일명과 제목에서 제거한다.
  - support tier: `T1 Direct Evidence`
  - evidence: 사용자 질문 "각 문서 앞에 넘버링 있는데, 이게 굳이 필요한 건지?"와 이어진 진행 요청.
  - admission lane: `APPLY`
- D-02: 내부 순서는 `MAJORS` 리스트 순서로 유지한다.
  - support tier: `T2 Strong Inference`
  - evidence: README의 대분류 표시와 manifest generation은 리스트 순서를 쓰므로, 파일명 번호 없이도 안정적인 출력 순서를 만들 수 있다.
  - admission lane: `APPLY`
- D-03: 레거시 numbered 파일 cleanup을 생성기에 포함한다.
  - support tier: `T2 Strong Inference`
  - evidence: 생성기를 다시 실행했을 때 예전 numbered 파일이 남으면 root 학습 표면이 중복된다.
  - admission lane: `APPLY`

## 5. Execution Log

- updated `interviews/tools/build_interview_curriculum.py`
  - `MAJORS.file_name`: `01-language-runtime.md` -> `language-runtime.md` 등 10개 변경
  - `MAJORS.title`: `01. 언어와 런타임` -> `언어와 런타임` 등 10개 변경
  - `LEGACY_MAJOR_FILES` 추가
  - `clear_generated_outputs()`가 numbered legacy files도 삭제하도록 보강
- ran `python3 interviews/tools/build_interview_curriculum.py`
- regenerated:
  - `interviews/language-runtime.md`
  - `interviews/concurrency-async-io.md`
  - `interviews/os-kernel-computer-architecture.md`
  - `interviews/network-web-protocols.md`
  - `interviews/security-cryptography.md`
  - `interviews/database-storage-search-nosql.md`
  - `interviews/messaging-event-driven.md`
  - `interviews/distributed-systems-architecture.md`
  - `interviews/spring-backend-frameworks.md`
  - `interviews/problem-solving-code-quality.md`
  - `interviews/README.md`
  - `interviews/_question-index.md`
  - `interviews/_curriculum_manifest.json`

## 6. Verification Log

- `python3 interviews/tools/build_interview_curriculum.py`: `PASS`
  - source chunks: `241`
  - unique chunks: `133`
  - major documents: `10`
- manifest/path verification: `PASS`
  - legacy numbered files removed
  - unnumbered major files exist
  - old `interviews/curriculum/` directory absent
  - `interviews/source/` raw source files: `7`
  - manifest major files have no `NN-` prefix
  - manifest major titles have no `NN.` prefix
  - every source ref starts with `source/`
  - no fallback placement `추가 분류 필요 항목`
  - each unique SHA marker appears once
- `python3 -m py_compile interviews/tools/build_interview_curriculum.py interviews/tools/split_interview_sources.py`: `PASS`
- `git diff --check -- interviews docs/works/WORK_20260505_INTERVIEWS_UNNUMBERED_CURRICULUM_FILES.md`: `PASS`

## 7. Final Audit

- requested closure scope: remove unnecessary numbering from the interview curriculum document surface.
- achieved closure scope: file names, H1 titles, README links, question index, and manifest now use unnumbered names.
- source preservation: raw source files remain under `interviews/source/`.
- regeneration path: `python3 interviews/tools/build_interview_curriculum.py` reproduces the unnumbered root curriculum.
- unrelated open work: existing unrelated dirty files outside `interviews` remain untouched.
- whole-request objective status: `COMPLETE_FOR_THIS_NUMBERING_REMOVAL_TRANCHE`.
