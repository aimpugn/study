# WORK_20260626_MAVEN_LOCAL_SHIM_NUSHELL_RUNBOOK

> 템플릿: [`AGENTS_WORK_TEMPLATE.md`](../../AGENTS_WORK_TEMPLATE.md). 이번 작업은 기존 Maven 문서 보강 + troubleshooting 진입 링크 추가.

## 0. Meta

- 작업 제목: Maven local shim 저장소와 Windows/Nushell offline 실행 런북 정리
- 작업 유형: `explain + execute`
- 작업 깊이: `standard` (기존 중심 문서 보강, 관련 진입 문서 2개 동기화)
- 원문 사용자 요청: 현재 로컬에서 public repository 없이도 `mvn` 사용 가능한 방법을, Maven 동작 방식·Codex setup shell script·local shim 개념까지 연결해 `C:\Users\rody\WorkspacePrivate\study`에 체계적으로 정리
- 대상:
  - `jvm/maven_local_shim_repo.md`
  - `troubleshooting/windows.md`
  - `troubleshooting/maven.md`
- 시작/종료: 2026-06-26
- 현재 상태: `COMPLETE`

## 1. Scope

- 기존 중심 문서 `jvm/maven_local_shim_repo.md`에 새 13절을 추가했다.
- 새 13절은 `JAVA_HOME`, mvnd 번들 Maven, `.mvn/maven.config`, `.codex-m2`, `mvn -o`, `codex-setup.sh`, CRLF shell script 실패를 한 흐름으로 설명한다.
- `troubleshooting/windows.md`와 `troubleshooting/maven.md`에는 증상에서 중심 문서로 이동할 수 있는 링크를 추가했다.

## 2. Evidence

- `kcf-firmbanking-service`에서 직접 확인한 값:
  - Maven bin: `C:\dev\mvnd\1.0.6\mvn\bin`
  - JDK 21: `C:\dev\jdk\21.0.11+10`
  - `.mvn/maven.config`: `-Dmaven.repo.local=.codex-m2`
  - `.codex-m2`: 존재
  - `mvn -o -q -DskipTests test`: exit 0
  - focused JUnit 5개 클래스: `Tests run: 67, Failures: 0, Errors: 0, Skipped: 0`, `BUILD SUCCESS`
- 제약:
  - 현재 Codex PowerShell 세션에서는 `nu` 실행 파일이 PATH에 없어 Nushell 자체 실행은 직접 검증하지 못했다.
  - Nushell 명령은 Windows Nushell의 `$env.NAME` 및 `$env.PATH` 리스트 조작 문법으로 작성했다.

## 3. Decisions

- 새 문서를 만들지 않고 `jvm/maven_local_shim_repo.md`를 중심 문서로 확장했다.
  같은 주제를 설명하는 문서가 이미 있었고, 새 문서를 만들면 local shim 원리와 실제 Windows 런북이 갈라질 위험이 있었다.
- troubleshooting 문서에는 자세한 원리를 중복하지 않았다.
  이 파일들은 증상별 입구 역할이므로, 중심 문서 링크만 추가하는 편이 유지보수 비용이 낮다.

## 4. Verification

- 실행한 검증:
  - `git diff --check` -> PASS. trailing whitespace 5건을 수정한 뒤 통과.
  - `rg -n "실제 Windows/Nushell|관련 원리 문서|maven_local_shim_repo.md#13|WORK_20260626" ...` -> 새 목차, 실제 13절 제목, troubleshooting 링크, WORK 문서 위치 확인.
  - `git status --short` -> 의도한 4개 파일만 변경.
- 실행하지 못한 검증:
  - `markdownlint-cli2`/`markdownlint`는 PATH에 없어 실행하지 못했다.
  - `nu` 실행 파일도 현재 Codex PowerShell PATH에 없어 Nushell 명령을 직접 실행하지 못했다.

## 5. Closure

- 사용자 요구: PASS.
- 남은 리스크:
  - Nushell 명령은 현재 Codex 세션에서 직접 실행하지 못했으므로, 사용자의 실제 Nu 세션에서 한 번 replay하면 더 강해진다.
