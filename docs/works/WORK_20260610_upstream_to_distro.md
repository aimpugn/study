# WORK_20260610_UPSTREAM_TO_DISTRO

> 템플릿: [`AGENTS_WORK_TEMPLATE.md`](../../AGENTS_WORK_TEMPLATE.md). 작은 항목은 축약하되 근거·반박·검증·최종감사 루프는 유지.

## 0. Meta

- 작업 제목: 업스트림 → 배포판 릴리스 파이프라인 학습 문서 신설 및 관련 문서 동기화
- 작업 유형: `explain + execute`
- 작업 깊이: `full` (개념 문서 신설 트리거)
- 원문 사용자 요청: "이런 정책적인 것도 히스토리가 있고 이유가 있고 맥락이 있을 텐데, 이것들도 잘 정리해서 지식화 하고 싶음. 어디에 어떻게 잘 정리하는 게 좋을지?"
- 대상 경로: `linux/upstream_to_distro.md`(신설), `linux/linux_distributions.md`, `linux/ubuntu/ubuntu_versions.md`, `linux/ubuntu/ubuntu_apt.md`, `terminology/upstream.md`, `terminology/distribution.md`
- 시작/종료: 2026-06-10
- 현재 상태: `COMPLETE`
- finish: `verify + commit`

## 1. Request Normalization

- goal: nginx 버전 차이 → 커널 개발 모델 → stable 백포트 안전장치 → 실증까지 이어진 대화의 지식을, 시간이 지나도 재검증 가능한 학습 문서로 영속화.
- scope: 릴리스 모델·백포트 파이프라인 전반(커널 중심 + nginx 일반화). 우분투 아카이브 내부 구조 심화, SRU 절차 상세는 비범위(필요 시 후속 상세 문서).
- must_keep: 대화에서 확보한 1차 관측(2026-06-10 fetch 결과)을 출처로 보존, 시점 민감 정보에 관측 날짜 표기.

### 1.2 Non-Goals

- Debian 패키징 실습(`debian/` 디렉토리 구조), Launchpad 업로드 절차 상세 — 이번 질문의 줄기가 아님.

## 2. Root-First Framing

- 근본 문제: "배포판 저장소는 왜 latest가 아닌가"라는 표면 질문 아래에, 단일 개발 라인(SSOT) + 선별 백포트 채널이라는 생태계 구조와 그 역사·안전장치가 있다. 이 구조를 모르면 버전 스캐너 오탐 해석, 외부 저장소 추가 판단, 커널 업그레이드 전략을 매번 틀리게 추론한다.
- 성공 정의: 독자가 (1) 대화의 5개 질문에 자기 말로 답하고, (2) 실물 커밋 메타데이터를 읽을 수 있고, (3) PASS/FAIL 명령으로 스스로 재검증할 수 있는 문서.

## 3. Reader & Internalization Contract (요약)

- teach-back 목표: "기여는 위로만, 코드는 아래로만. 동결은 방치가 아니라 백포트와 한 쌍."
- 특히 막아야 하는 오해: 버전 낮음=취약 / cherry-pick이 히스토리를 끌고 옴 / 배포판 트리가 기여 목적지 / upstream 호환 보장=배포판 반영.
- 핵심 대조쌍: cherry-pick vs merge, "호환된다" vs "변하지 않는다", mainline vs stable, GA vs HWE.
- primary exemplar: `git/git_rebase.md` — 참고 원리: 목차, 코드블록 인라인 주석 해부, 실제 명령 중심. 따라 하지 않을 trait: 누적형 Q&A 나열로 주제 흐름이 약해지는 구조, 정의 중복.
- secondary 기준: `ai/authoring/LEARNING_DOC_GUIDE.md` 부록 A 스켈레톤 + 부록 B hard gate.

## 4. Depth Decision

- `full`. 개념 문서 신설 + 여러 자산 동기화 + 역사·근거 밀도가 핵심인 주제(가이드 B.0 기준 "큰 주제": 역사 존재, 구현 다수, 비교 대상 다수, 다계층).

## 6~7. Topic Analysis & Critique

- Challenge 1 — "대화 요약문이 되어 버릴 위험": 대화 순서가 아니라 학습 전개(역사→지형→실물→방어선→배포판→비교)로 재구성. 사용자의 실제 오해 5개는 "혼동하기 쉬운 것들" 절로 승격해 보존.
- Challenge 2 — "시점 민감 정보의 부패": 버전·EOL·주기를 "2026-06-10 관측"으로 라벨링하고, 재관측 절차를 검증 경로 6번으로 문서 안에 내장.
- Challenge 3 — "기존 문서와의 역할 충돌": `linux_distributions.md`(Docker 태그 중심), `ubuntu_versions.md`(코드명 메모)는 이번 질문에 답하지 않음을 확인 → 신설이 맞고, 양쪽에 역할 분리 링크만 추가.

## 8. Scope Expansion & Impact Sync

- 확장 키워드: upstream, distribution, 배포판, stable, LTS, backport, cherry-pick, SRU, HWE, apt, 저장소, 버전.
- 함께 점검한 자산: `linux/` 전체 목록, `linux/ubuntu/*`, `terminology/upstream.md`, `terminology/distribution.md`, `knowledge/cards`(판단 카드 아님 → 비대상), 루트 `README.md`(개별 문서 인덱스 없음 → 비대상).
- 동기화 실행: 관련 문서 5곳에 교차 링크. `ubuntu_versions.md`의 낡은 추측(Lunar/Noble을 미래 버전으로 기술)과 잘못된 지원 기간 서술(전 릴리스 5년)을 정정 — 한쪽만 바꾸면 새 문서와 기존 문서가 모순되는 드리프트가 생기는 지점.

## 9. Evidence Ledger

- E-01 stable 규칙(원문 인용: mainline 선행, 100줄, 버그 자격, 선행 커밋 문법, 48h 리뷰)
    - 근거 유형: official doc — kernel.org stable-kernel-rules (2026-06-10 fetch)
- E-02 리뷰·실패 통지·릴리스의 현행 가동
    - 근거 유형: 1차 관측 — stable 리스트 미러(spinics)에서 `[PATCH 6.12 000/307] 6.12.93-rc1 review`, `FAILED: patch ... 6.12-stable`, `Linux 7.0.12` 제목 확인 (2026-06-10)
- E-03 stable 커밋 메타데이터 실물
    - 근거 유형: 1차 관측 — linux-6.12.y `b7b72e88...` 커밋에서 `commit 13031fb6... upstream`, `Fixes: 8201d1028caa`, `Cc: stable` 확인 (kernel.googlesource 미러)
- E-04 의존성의 명시 입장 실물
    - 근거 유형: 1차 관측 — `1285e83...` 리팩터링 커밋의 `Stable-dep-of: 3b041514cb6e` 태그
- E-05 현행 LTS·EOL·릴리스 주기
    - 근거 유형: official doc — kernel.org releases 페이지 (mainline 9~10주, stable 주 1회, LTS 표)
- E-06 우분투 SRU 주기·26.04 커널 베이스
    - 근거 유형: official doc — kernel.ubuntu.com (4~5주 SRU, 2주 보안 주기, 26.04=7.0 기반)
- E-07 역사 연표(1991~2023)
    - 근거 유형: 널리 확인된 공개 기록. AUTOSEL 도입 연도만 `추정` 표시로 강등. 2023 LTS 단축 방침은 발표 사실 + "실제 운영은 연장" 관측을 분리 서술.

## 10. Evidence Critique + Repair

- lore.kernel.org 원본이 봇 차단(Anubis) → 미러(spinics)·git 미러(googlesource)로 대체 확보. 미러 의존 한계는 출처 절에 미러임을 명시해 보완.
- 25.04 커널 버전 등 확신 낮은 개별 값은 본문에서 제외(jammy/noble/26.04 관측만 사용).
- 우분투 도달 단계는 이 커밋에 대해 미관측 → 본문에서 "절차상 흐름"으로 분리 표기.

## 11~12. Design & Critique

- 채택 구조: 직답 → 역사(타임라인 15항목 + 설계 사고 5질문) → 지형(다이어그램+용어) → 실물 해부(같은 수정을 7시점 trace) → 방어선 5겹 → 의존성 4갈래 → 배포판 절반 → 8축×6채널 비교표 → 오해 5쌍 → 실패 모드 → 검증 6경로 → 출처.
- 대안 검토: (a) 커널/우분투 2개 문서 분리 — 대화의 핵심이 "두 세계의 연결"이므로 단일 monograph 채택, 분량이 더 커지면 후속 분리(가이드 3.5). (b) knowledge/cards 카드화 — 판단 카드가 아니라 학습 문서이므로 비채택.
- 비유(출판 정오표)는 깨지는 지점(정오표 간 의존성)을 본문에 명시.

## 13/15. Plan (실행 파일)

1. `linux/upstream_to_distro.md` 신설 (본문)
2. 관련 문서 5곳 교차 링크 + `ubuntu_versions.md` 사실 정정
3. WORK 로그(본 파일)
4. 링크 경로 검증 → 커밋

## 17. Frozen Checklist

- C-01 (사용자) 대화의 지식이 문서화됨 — PASS 기준: 대화의 5개 질문이 개요에 명시되고 본문에서 각각 닫힘.
- C-02 (AI) LEARNING_DOC_GUIDE hard gate — PASS 기준: 원자료 줄 단위 해부 / 실제 관측값 / 1차 출처 칸 / 다축 비교 / 역사 연도·인물 모두 존재.
- C-03 (AI) 시점 민감 정보 라벨 — PASS 기준: 버전·EOL·주기 값에 "2026-06-10 관측" 표기 + 재관측 경로 존재.
- C-04 (AI) 영향 범위 동기화 — PASS 기준: 관련 5문서 링크 + 기존 문서의 모순 사실 정정.
- C-05 (AI) repo closure — PASS 기준: WORK 로그 존재 + 커밋 완료.

## 18. Execution Log

- 조사: 저장소 구조, AGENTS.md, LEARNING_DOC_GUIDE.md, WORK 템플릿, exemplar(git_rebase.md), 기존 linux/ubuntu/terminology 문서, .markdownlint.json.
- 수정: 계획 4개 항목 그대로 실행. 버린 접근 — 대화 Q&A 원문 순서 보존(요약문화 위험으로 폐기), 2문서 분리(단일 연결 서사 우선으로 폐기).

## 19. Verification

- 실행한 검증: (1) 본문 인용값(커밋 해시·메일 제목·규칙 문구·EOL·SRU 주기)을 당일 fetch 결과와 대조 — 일치. (2) 신설/수정 문서의 상대 링크 대상 존재 확인(Test-Path) — 통과. (3) git 커밋.
- 실행하지 못한 검증과 이유: 본문 "직접 확인해 보기" 1~4번(apt/git 명령)은 리눅스 머신·커널 클론이 필요해 이 Windows 작업 환경에서 미실행. 문서에는 실행하지 않았음을 전제로 PASS/FAIL 기준만 제공(실행한 것처럼 쓰지 않음). markdownlint CLI 미설치로 정적 lint 미실행, `.markdownlint.json` 규칙(4칸 들여쓰기, 줄 길이 무제한)을 수동 준수.

## 20. Explanation Quality Review (판정 요약)

- 직답 초반 배치 / 역사·설계 사고 닫힘 / 실물 artifact 우선 / 같은 대상 다시점 trace / 대조쌍 대칭 / 사실·추론(추정 표기)·미관측 분리 / 검증 경로 PASS·FAIL 명시 — 충족.
- 잔여 한계: 비교표의 RHEL 열은 공개 기록 기반(직접 관측 아님). AUTOSEL 도입 연도 추정. 본문에 동일하게 표시됨.

## 21. Final Audit & Closure

- intent-fit: 사용자의 "어디에·어떻게"에 대해 — 어디에: `linux/upstream_to_distro.md`(허브) + 기존 문서 링크망. 어떻게: 저장소 규범(monograph + hard gate + 관측 라벨)으로. 둘 다 실물로 답함.
- remaining risks: 시점 민감 값의 자연 부패 — 문서 내 재관측 절차로 완화.
- C-01~C-05: 전부 PASS.
- 최종 상태: `COMPLETE` (커밋은 본 작업 커밋에 포함)
