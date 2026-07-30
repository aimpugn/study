# WORK_20260730_GIT_MIRROR_MIGRATION_GUIDE

> 템플릿: [`AGENTS_WORK_TEMPLATE.md`](../../AGENTS_WORK_TEMPLATE.md). 반복 필드는 축약하되 근거, 반박, 검증, 최종 감사는 유지한다.

## 0. Meta

- 작업 제목: `git clone --mirror` 결과를 새 원격 저장소에 안전하게 반영하는 가이드
- 작업 유형: `research + analysis + explain + execute`
- 작업 깊이: `full`
- 원문 사용자 요청: "git clone mirror 한 결과를 새로운 원격 저장소로 push하여 그대로 반영하는 방법 가이드"
- 대상 경로: `git/git_clone_mirror.md`, `git/clone.md`, `git/push.md`, `git/lfs.md`
- 시작일: 2026-07-30
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `verify + commit`

## 1. Request Normalization

- goal: mirror clone에서 새 원격으로 Git 데이터를 이전할 때 복사 범위, 삭제 위험, 호스팅 서비스 예외, LFS와 검증을 한 문서에서 재현 가능하게 설명한다.
- scope: 일회성 이전을 기본 경로로 삼고, GitLab/GitHub hidden ref와 주기 동기화를 보조 경로로 다룬다.
- must_keep: 새 대상이 비어 있다는 전제, source/target ref 해시 비교, LFS와 호스팅 메타데이터 경계를 명시한다.
- non-goals: 특정 실제 원격 저장소에 push, GitLab/GitHub 프로젝트 메타데이터 자동 이전 도구 구현.

## 2. Root-First Framing

- 근본 문제: `clone --mirror`와 `push --mirror`를 "전체 복사"라는 이름만 보고 실행하면 대상 전용 ref가 삭제되고 서버 관리 hidden ref가 거부될 수 있다.
- 성공 정의: 독자가 빈 대상에 안전하게 이전하고, 원본과 대상의 ref 이름과 객체 ID를 비교하며, Git 밖의 미이전 자산을 분리해 후속 조치할 수 있다.
- PARTIAL 조건: 명령은 있으나 삭제 의미, hidden ref, LFS, 기본 브랜치, 사후 검증 중 하나라도 빠진다.
- BLOCKED 조건: 공식 근거나 실행 검증 없이 destructive 명령을 권고한다.

## 3. Reader & Internalization Contract

- 주 독자: Git 브랜치와 원격 저장소는 알지만 ref 네임스페이스와 mirror의 삭제 의미는 익숙하지 않은 개발자.
- teach-back 목표: "`--mirror`는 모든 공개 ref를 같은 이름으로 복제하고, push 시 대상 ref를 강제로 맞추므로 빈 대상과 사후 ref 비교가 필요하다."
- 첫 번째 벽돌: 바로 실행 가능한 PowerShell 명령과 "대상은 비어 있어야 한다"는 중단 조건.
- 핵심 대조쌍: 일반 clone vs mirror clone, 일반 push vs mirror push, Git 데이터 vs 호스팅 메타데이터, 전체 mirror vs 브랜치/태그 선택 push.
- 특히 막아야 하는 오해:
    - mirror push가 기존 대상 데이터를 보존하면서 추가만 한다.
    - 성공 종료만 확인하면 원본과 대상이 같다.
    - Git LFS와 MR/PR도 Git ref와 함께 모두 이전된다.
    - 대상의 기본 브랜치까지 `--mirror`가 설정한다.
- primary exemplar: `git/git_rebase.md`.
    - 참고 원리: 실제 명령과 상태 변화를 연결하고, ref 역할을 명령 인자 수준에서 설명한다.
    - 따라 하지 않을 visible trait: 정의 중복과 누적형 사례로 문서 중심 흐름이 길어지는 부분.
- secondary exemplar: `algorithms/dynamic_programming.md`.
    - 참고 원리: 작은 실제 예에서 관측을 만든 뒤 일반 규칙으로 확장한다.
    - 따라 하지 않을 visible trait: 이번 명령 가이드에 필요하지 않은 넓은 taxonomy.
- exemplar보다 강화할 축: destructive 범위, 중단 조건, rollback, 실제 ref 해시 비교, 호스팅 경계를 한 실행 흐름에 결합한다.

## 4. Depth Decision

- `full`.
- 개념 문서 신설이며 잘못된 명령은 새 원격의 ref를 강제 갱신하거나 삭제한다. 공식 문서, 로컬 실험, 실패 경계, 검증 경로가 모두 필요하다.

## 6-7. Topic Analysis & Critique

- Challenge 1: 가장 짧은 두 명령만 제시하면 대상 삭제 범위를 숨긴다.
    - 보강: 빈 대상 확인, dry-run, ref별 출력 해석, 사후 비교를 기본 경로에 포함한다.
- Challenge 2: "그대로"를 프로젝트 전체 이전으로 오해할 수 있다.
    - 보강: Git ref/객체, LFS, 서브모듈, symbolic `HEAD`, 호스팅 메타데이터를 표로 분리한다.
- Challenge 3: GitLab/GitHub의 서버 관리 ref가 전체 mirror를 막는다.
    - 보강: hidden ref를 사전 탐지하고 브랜치/태그 명시 refspec으로 전환하는 보조 경로를 제공한다.

## 8. Scope Expansion & Impact Sync

- 시작 키워드: `git clone --mirror`, `git push --mirror`, 새 원격 저장소.
- 확장 키워드: bare, refspec, prune, forced update, hidden ref, `refs/merge-requests`, `refs/pull`, LFS, symbolic `HEAD`, default branch, atomic push.
- 점검 자산: `git/clone.md`, `git/push.md`, `git/remote.md`, `git/lfs.md`, `git/git.md`, `git/git_rebase.md`.
- 동기화 표면: 새 가이드를 `clone.md`, `push.md`, `lfs.md`에서 찾을 수 있게 연결한다.
- 제외 표면: 루트 README는 개별 Git 문서를 열거하는 허브가 아니므로 수정하지 않는다.

## 9. Evidence Ledger

- E-01
    - 주장: `git clone --mirror`는 `--bare`를 포함하고 모든 공개 ref를 같은 이름으로 매핑한다.
    - 근거: Git 공식 `git-clone` 문서와 로컬 config 관측.
- E-02
    - 주장: `git push --mirror`는 ref를 강제 갱신하고 대상 전용 ref를 삭제한다.
    - 근거: Git 공식 `git-push` 문서와 로컬 divergent `main`/`obsolete` 실험.
- E-03
    - 주장: GitLab 내부 ref는 전체 mirror push를 거부할 수 있다.
    - 근거: GitLab Gitaly 공식 문서와 과거 실제 `refs/merge-requests/*` 거부 관측.
- E-04
    - 주장: LFS 실제 파일은 별도 fetch/push가 필요하다.
    - 근거: GitHub 공식 저장소 복제 가이드.
- E-05
    - 주장: ref 이름과 객체 ID의 사후 비교가 가능하다.
    - 근거: Git 공식 `git-ls-remote` 출력 계약과 로컬 실험.

## 10. Evidence Critique + Repair

- 공식 Git 문서는 일반 ref 동작을 설명하지만 호스팅 서비스 정책을 보장하지 않는다.
- GitLab 공식 문서와 GitHub 공식 문서를 추가해 hidden ref와 LFS 경계를 보완한다.
- 로컬 실험은 Windows Git 2.54.0 bare 저장소 사이의 동작만 증명한다. 원격 hook, quota, 권한은 문서에서 별도 미확인 경계로 둔다.

## 11-12. Design & Critique

- 채택 구조: 직답 -> 복사 경계 -> destructive 의미 -> 안전한 시간 순서 -> hidden ref 분기 -> 주기 동기화 -> 실패와 rollback -> 검증 근거.
- 대안 1: `clone --mirror`와 `push --mirror` 두 명령만 `clone.md`에 추가.
    - 폐기 이유: 삭제 위험과 프로젝트 이전 경계를 설명하지 못한다.
- 대안 2: 기존 `clone.md` 전체를 장문으로 재작성.
    - 폐기 이유: 특정 브랜치 clone 메모와 destructive migration runbook의 역할이 달라 별도 문서가 더 응집적이다.
- 트레이드오프: 문서가 길어지지만 실제 실행자가 다른 문서를 찾아 조립하지 않아도 된다.

## 13-16. Plan

1. 공식 Git/GitLab/GitHub 문서에서 mirror, hidden ref, LFS, ref 조회 계약을 확인한다.
2. 로컬 bare 저장소에서 강제 갱신, 대상 ref 삭제, notes/tag/custom ref 이전을 재현한다.
3. `git/git_clone_mirror.md`를 신설한다.
4. `clone.md`, `push.md`, `lfs.md`에 역할이 분명한 교차 링크를 추가한다.
5. 링크, Markdown, 명령 구문, git diff를 검증한다.
6. 최종 감사 뒤 관련 파일만 커밋한다.

### Success / Failure / Regression Cases

- S1: 빈 대상에 전체 mirror push 후 source/target의 공개 ref 해시가 모두 같다.
- S2: GitLab hidden ref가 있으면 전체 push를 중단하고 브랜치/태그 refspec으로 전환할 수 있다.
- S3: LFS, 기본 브랜치, 서브모듈, 호스팅 메타데이터의 별도 검증 경로가 있다.
- F1: 대상이 비어 있지 않은데 경고 없이 `--mirror`를 실행한다.
- F2: push 종료 코드만 보고 완료하며 ref 해시를 비교하지 않는다.
- F3: Git LFS와 MR/PR까지 자동으로 이전된다고 단정한다.
- 회귀 위험: 기존 짧은 `clone.md`, `push.md`, `lfs.md`의 역할을 무리하게 확장하거나 기존 내용을 덮어쓰는 일.

## 17. Frozen Checklist

- C-01 (사용자): 새 원격에 mirror 결과를 반영하는 실행 가이드가 있다.
    - PASS: clone, preflight, push, verification 명령이 한 문서에 있다.
- C-02 (AI): destructive 의미와 중단 조건이 있다.
    - PASS: forced update와 delete를 설명하고 빈 대상 및 dry-run을 요구한다.
- C-03 (AI): "그대로"의 범위를 정확히 제한한다.
    - PASS: Git ref/객체와 LFS, 기본 브랜치, 서브모듈, 호스팅 메타데이터를 구분한다.
- C-04 (AI): GitLab/GitHub hidden ref 분기가 있다.
    - PASS: 오류 원인, 탐지 ref, 브랜치/태그 대체 명령, 선택 범위 검증이 있다.
- C-05 (AI): 실제 검증과 공식 근거가 있다.
    - PASS: 로컬 mutation 실험, 링크 검사, Markdown 검사, 공식 문서 링크가 남는다.
- C-06 (AI): 저장소 closure를 지킨다.
    - PASS: 관련 파일만 stage하고 commit하며 push하지 않는다.

## 18. Execution Log

- 기존 Git 문서와 exemplar, WORK 템플릿, Markdown 규칙을 확인했다.
- 공식 Git `clone`, `push`, `ls-remote`, GitLab Gitaly, GitHub repository duplication 문서를 확인했다.
- 로컬 bare 저장소 실험을 실행했다. 계산된 임시 경로의 재귀 정리를 포함한 첫 명령은 정책에서 실행 전 차단됐고 파일은 생성되지 않았다.
- 고정된 `.tmp/git-mirror-experiment-20260730` 경로에서 재실행해 ref 동작을 확인했다.
- 실험 디렉터리는 경로를 확인한 뒤 휴지통으로 이동했다.
- Git plumbing으로 만든 두 번째 최소 저장소 실험에서 `git push --mirror` 종료 코드 0, source/target ref 차이 0건, 대상 전용 `obsolete` ref 삭제를 다시 확인했다.
- 두 번째 실험 디렉터리도 경로를 확인한 뒤 휴지통으로 이동했다.
- 본문과 관련 문서 교차 링크를 작성했다.

## 19. Verification

- `npx --yes markdownlint-cli2`로 대상 Markdown 5개 검사: PASS, 오류 0건.
- 상대 Markdown 링크의 파일 경로 검사: PASS, 누락 0건.
- `git diff --check`: PASS, whitespace 오류 0건.
- PowerShell 배열 refspec과 `Compare-Object` 예제 실행: PASS, 비교 차이 0건.
- bare mirror 설정 관측: `remote.origin.fetch=+refs/*:refs/*`, `remote.origin.mirror=true`.
- mutation 실험:
    - 갈라진 대상 `main`: forced update 확인.
    - 대상 전용 `obsolete`: 삭제 확인.
    - tag, notes, 사용자 정의 ref: 생성 확인.
    - source/target ref 이름과 객체 ID 차이: 0건.
    - 두 번째 최소 실험의 실제 push 종료 코드: 0.
- `git fsck --full`: 객체 손상 오류 0건. 강제 갱신 전 대상 커밋의 `dangling commit`만 관측.
- 공식 링크: Git `clone`, `push`, `ls-remote`, GitLab Gitaly, GitHub repository duplication 문서를 2026-07-30에 열어 본문 주장과 대조.
- 실행하지 못한 검증: 실제 GitLab/GitHub/사내 원격에는 push하지 않았다. 권한, hook, quota, 보호 브랜치 동작은 대상별 preflight 항목으로 남겼다.

## 20. Explanation Quality Review

- 결론과 최소 명령을 먼저 제시하고, 같은 자리에서 빈 대상 전제와 삭제 위험을 설명했다.
- `clone --mirror`와 `push --mirror`의 상태 변화를 ref 수준에서 연결했다.
- Git 데이터와 LFS, 기본 브랜치, 서브모듈, 호스팅 메타데이터를 표와 검증 경로로 분리했다.
- 정상 경로와 hidden ref 경로를 섞지 않고 별도 refspec과 별도 비교 범위로 설명했다.
- 각 destructive 단계 앞에 중단 조건을 두고, 종료 코드뿐 아니라 ref 이름과 객체 ID를 비교하게 했다.
- 로컬 mutation 실험의 관측과 실제 호스팅 서버에서 미검증한 정책 경계를 구분했다.
- 문서의 명령을 순서대로 실행하면 다른 문서를 찾아 조립하지 않고도 preflight, push, verification, rollback 판단을 수행할 수 있다.

## 21. Final Audit & Closure

- intent-fit: mirror clone 결과를 새 원격에 반영하는 직접 명령과 안전한 운영 순서를 모두 제공한다.
- expert-perspective: 대상 삭제, hidden ref, partial push, LFS, symbolic `HEAD`, 쓰기 중지 시점을 포함해 "성공 출력만 확인"하는 오류를 막는다.
- remaining risks: 실제 대상의 권한, hook, quota, 보호 브랜치와 메타데이터 이전 수단은 대상 서비스마다 다르다. 본문에서 미확인 경계와 중단 조건으로 명시했다.
- 관련 자산 동기화: `clone.md`, `push.md`, `lfs.md`에서 새 가이드로 연결했다.
- C-01: `PASS`
- C-02: `PASS`
- C-03: `PASS`
- C-04: `PASS`
- C-05: `PASS`
- C-06: `PASS`
- 최종 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- 커밋: 본 WORK 문서를 포함하는 작업 커밋으로 마감한다.
