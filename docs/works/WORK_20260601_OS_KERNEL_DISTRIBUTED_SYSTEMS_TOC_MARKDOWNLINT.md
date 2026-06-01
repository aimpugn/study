# OS Kernel Distributed Systems ToC Markdownlint

## 0. Meta

- 작업 제목: OS/kernel/distributed systems deep dive 문서 ToC 생성과 markdownlint 정렬
- WORK 파일 경로: `docs/works/WORK_20260601_OS_KERNEL_DISTRIBUTED_SYSTEMS_TOC_MARKDOWNLINT.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `execute | refactor_docs`
- 작업 깊이: `standard`
- 관련 요청: "모든 문서에 ToC가 없습니다. markdownlint에 따라서 ToC 생성하고 문서들 린트 맞춰 주세요."
- 대상 경로 / 자산: `interviews/os-kernel-distributed-systems-deep-dive/**/*.md`
- 현재 상태: `IN_PROGRESS`
- 완료 게이트: `PENDING`
- finish: `test+commit`

## 1. Request Normalization

- goal: 대상 corpus의 모든 Markdown에 링크형 목차를 추가하고 markdownlint 위반을 제거한다.
- scope: git에 추적되는 `os-kernel-distributed-systems-deep-dive` Markdown 18개와 이 WORK 기록.
- mode: `execute`
- run_mode: `normal`
- finish: markdownlint 통과, diff 검수, commit.
- must_keep: 문서 의미와 본문 구조를 보존하고, 기존 unrelated dirty worktree는 건드리지 않는다.
- extra_checks:
    - 모든 대상 Markdown에 `## 목차`가 있어야 한다.
    - `npx --yes markdownlint-cli2 --config .markdownlint.json "os-kernel-distributed-systems-deep-dive/**/*.md"`가 통과해야 한다.
    - 변경 범위는 대상 corpus와 WORK 파일로 제한되어야 한다.

## 2. Checklist Freeze

- [x] 대상 Markdown 18개 확인
- [x] 각 문서에 markdownlint 친화적인 ToC 삽입
- [x] baseline lint 위반(`MD040`, `MD060`, `MD034` 등) 수리
- [x] markdownlint 재실행 통과
- [x] final diff review와 whitespace check 통과
- [x] 관련 변경만 staged 후 commit

## 3. Evidence Ledger

- E-01: `git ls-files -- os-kernel-distributed-systems-deep-dive` 결과 대상 Markdown 18개를 확인했다.
- E-02: baseline markdownlint 결과 18개 파일에서 440개 오류가 확인되었고, 주된 유형은 fenced code language, table column style, bare URL이다.

## 4. Closure

- 최종 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- 검증 로그:
    - `rg -n "^## 목차$" os-kernel-distributed-systems-deep-dive`: 18개 대상 문서에서 ToC 확인.
    - `npx --yes markdownlint-cli2 --config .markdownlint.json "os-kernel-distributed-systems-deep-dive/**/*.md"`: 18개 파일, 0 errors.
    - `git diff --check -- os-kernel-distributed-systems-deep-dive ../docs/works/WORK_20260601_OS_KERNEL_DISTRIBUTED_SYSTEMS_TOC_MARKDOWNLINT.md`: PASS.
    - `npx --yes markdownlint-cli2 --config .markdownlint.json "../docs/works/WORK_20260601_OS_KERNEL_DISTRIBUTED_SYSTEMS_TOC_MARKDOWNLINT.md"`: WORK 문서 lint PASS.
