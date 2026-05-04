# WORK_20260505_INTERVIEWS_PROJECT_INTENT

## 0. Meta

- 작업 제목: interviews 프로젝트 의도 고정
- WORK 파일 경로: `docs/works/WORK_20260505_INTERVIEWS_PROJECT_INTENT.md`
- 저장소 / 모듈: `/Users/rody/VscodeProjects/study` / `interviews`
- 작업 유형: `execute`
- 작업 깊이: `standard`
- 원문 사용자 요청: "이 프로젝트는 경력 기술 인터뷰를 준비하기 위한 프로젝트입니다. 그게 프로젝트 의도입니다."
- 대상 경로 / 자산: `interviews/README.md`, `interviews/PROJECT_INTENT.md`, `interviews/USECASE.md`
- finish: `verify+commit`
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`

## 1. Request Normalization

- goal: `interviews` 디렉터리의 프로젝트 의도를 경력 기술 인터뷰 준비로 고정한다.
- scope: `interviews` 서브프로젝트의 사실 문서와 진입점 문서.
- mode: `execute`
- run_mode: `normal`
- must_keep:
  - 루트 `study` 저장소의 학습 자산 목적과 충돌하지 않는다.
  - 질문 수집 문서를 정식 exemplar로 오인하지 않고 source reservoir로 둔다.
  - "작은 기술 단위의 깊은 이해"와 "짧은 시간 안의 설명"을 동시에 보호한다.
- extra_checks:
  - 개발 언어, OS, 커널, 네트워크, DB, 알고리즘, 암호학 등 넓은 주제 범위를 임의로 축소하지 않는다.
  - 시스템 아키텍처를 작은 기술 단위 조립의 결과로 설명한다는 의도를 보존한다.

## 2. Instruction Stack / Project Overlay

- 전역 AGENTS: `~/.codex/AGENTS.md` 확인. 프로젝트 목적/시나리오/용어는 `PROJECT_INTENT.md`, `USECASE.md`, `TERMINOLOGY.md`로 라우팅한다는 규칙을 적용했다.
- 전역 WORK template: `~/.codex/AGENTS_WORK_TEMPLATE.md` 확인. repo-root template이 있으면 우선 사용한다는 규칙을 확인했다.
- 로컬 AGENTS: `/Users/rody/VscodeProjects/study/AGENTS.md` 확인. `interviews`의 큰 질문 모음은 원재료이며, 질문 cluster를 정식 문서로 승격한다는 규칙을 적용했다.
- 로컬 WORK template: `/Users/rody/VscodeProjects/study/AGENTS_WORK_TEMPLATE.md` 확인. 이번 작업은 표준 깊이로 축약 ledger를 남긴다.
- 프로젝트 사실 문서: `/Users/rody/VscodeProjects/study/PROJECT_INTENT.md`, `/Users/rody/VscodeProjects/study/USECASE.md` 확인. 루트 목적은 전체 학습 저장소 기준으로 유지하고, `interviews`에는 하위 프로젝트 사실 문서를 추가한다.

## 3. Success / Failure Criteria

### Frozen Checklist

- [x] `interviews`의 목적이 경력 기술 인터뷰 준비로 명확히 고정된다.
- [x] 답변 기준이 "짧은 직답 + 깊은 메커니즘 설명"으로 드러난다.
- [x] 작은 기술 단위의 이해가 시스템 아키텍처 설명으로 조립된다는 의도가 보존된다.
- [x] 주제 범위가 개발 관련 기술 전체로 열려 있고 특정 기술 목록으로 축소되지 않는다.
- [x] 기존 질문 모음 파일을 원재료로 다루고 정식 기준선으로 승격하지 않는다.
- [x] 루트 `study` 목적/USECASE와 충돌하지 않는다.
- [x] 문서 링크와 Markdown 구조가 깨지지 않는다.

### PARTIAL 조건

- 문서는 추가되었지만 사용자 의도 중 하나가 약화되었거나 검증이 닫히지 않으면 `PARTIAL`이다.

### BLOCKED 조건

- 파일 접근, 쓰기 권한, 또는 Git 작업이 막혀 변경을 신뢰 가능하게 남길 수 없으면 `BLOCKED`다.

## 4. Scope / Impact

- 조사한 경로:
  - `interviews/*.md`
  - `AGENTS.md`
  - `PROJECT_INTENT.md`
  - `USECASE.md`
  - `README.md`
- 발견:
  - `interviews`에는 약 2.8만 줄의 질문/답변 수집 문서가 있지만 별도 `README.md`, `PROJECT_INTENT.md`, `USECASE.md`는 없었다.
  - 루트 `AGENTS.md`는 `interviews/interview_questions*.md`를 큰 원재료로 보고, 질문 cluster를 정식 문서로 승격하라고 이미 지시한다.
  - 루트 `PROJECT_INTENT.md`와 `USECASE.md`는 전체 학습 저장소 목적을 다루므로, 인터뷰 준비라는 하위 프로젝트 의도는 `interviews` 안에 두는 편이 더 낮고 정확한 계층이다.
- 제외한 표면:
  - 기존 질문 모음 본문 수정은 이번 작업의 목적이 아니므로 제외했다.
  - 별도 `TERMINOLOGY.md`는 새 용어 충돌이 아직 없으므로 만들지 않았다.

## 5. Decision / Claim Ledger

- D-01: 루트 `PROJECT_INTENT.md`를 바꾸지 않고 `interviews/PROJECT_INTENT.md`를 만든다.
  - support tier: `T1 Direct Evidence`
  - evidence: 루트 `PROJECT_INTENT.md`는 전체 study 저장소 목적을 설명하고, 현재 요청은 `interviews` 디렉터리의 목적을 특정한다.
  - counterexample check: 루트 목적을 바꾸면 다른 학습 자산까지 인터뷰 준비로 오해될 수 있다.
  - admission lane: `APPLY`
- D-02: `interviews/USECASE.md`도 함께 만든다.
  - support tier: `T2 Strong Inference`
  - evidence: 사용자는 "질문이 들어오면", "짧은 시간 안에", "작은 단위를 조립하여 아키텍처"라는 사용 장면을 함께 설명했다.
  - counterexample check: `PROJECT_INTENT.md` 하나에 모두 넣으면 목적과 사용 장면이 섞여 이후 읽기 기준이 흐려질 수 있다.
  - admission lane: `APPLY`
- D-03: 기존 질문 모음 파일은 수정하지 않는다.
  - support tier: `T1 Direct Evidence`
  - evidence: 요청은 프로젝트 의도 고정이며, 루트 AGENTS는 큰 질문 모음을 원재료로 보라고 한다.
  - counterexample check: 본문을 바로 재구성하면 의도 고정 작업이 대형 콘텐츠 정리 작업으로 확장된다.
  - admission lane: `APPLY`

## 6. Critique + Repair

- Critic finding: `interviews/PROJECT_INTENT.md`만 만들면 대표 사용 장면이 목적 문서에 과밀하게 섞일 수 있다.
- Repair: 사용 장면은 `interviews/USECASE.md`에 분리하고, `PROJECT_INTENT.md`는 목적, 비목표, 판단 기준에 집중했다.
- Critic finding: 주제 범위를 "기술 인터뷰"로만 적으면 알고리즘, 암호학, 커널 같은 넓은 범위가 약해질 수 있다.
- Repair: 개발 관련 기술 전체를 명시하고, 넓다는 말이 얕게 훑는다는 뜻이 아니라고 고정했다.
- Critic finding: 기존 수집 파일을 무시하면 현재 자산과 새 목적 문서가 분리될 수 있다.
- Repair: `README.md`, `PROJECT_INTENT.md`, `USECASE.md`에서 기존 큰 질문 모음을 원재료와 승격 대기소로 설명했다.

## 7. Verification Plan

- Markdown 파일이 실제로 존재해야 한다.
- `rg`로 핵심 문구와 링크 대상이 확인되어야 한다.
- `git diff --check`가 whitespace 오류 없이 통과해야 한다.
- 최종 diff에서 이번 작업 범위 밖 기존 질문 모음 본문이 수정되지 않아야 한다.

## 8. Final Audit

- 최종 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- 체크리스트 재판정: 전 항목 PASS.
- whole-request objective: `interviews` 프로젝트 의도를 로컬 사실 문서로 고정.
- achieved closure scope: `interviews/README.md`, `interviews/PROJECT_INTENT.md`, `interviews/USECASE.md`, `docs/works/WORK_20260505_INTERVIEWS_PROJECT_INTENT.md`.
- remaining open inside requested whole: 없음.
- unrelated open work disclosure: 기존 study 저장소에는 이번 작업 전부터 다수의 수정/미추적 파일이 남아 있으며, 이번 작업에서는 건드리지 않았다.
- next immediate target: 없음.
- verification:
  - `rg -n "경력 기술 인터뷰|짧은 직답|깊은 메커니즘|작은 기술 단위|시스템 아키텍처|개발 관련 기술 전체|원재료|승격" interviews/README.md interviews/PROJECT_INTENT.md interviews/USECASE.md`
  - `git diff --check -- interviews/README.md interviews/PROJECT_INTENT.md interviews/USECASE.md docs/works/WORK_20260505_INTERVIEWS_PROJECT_INTENT.md`
- final audit note: 루트 `study`의 일반 학습 저장소 목적은 그대로 두고, 인터뷰 준비라는 하위 목적만 `interviews` 안에 고정했다.
