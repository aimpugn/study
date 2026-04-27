# WORK_20260416_JAVA_SOCKET_CONNECTION_RESET_TROUBLESHOOTING

## 0. Meta

- 작업 제목: `java_socket_connection_reset.md` 신규 작성으로 troubleshooting proof round 수행
- WORK 파일 경로: `/Users/rody/VscodeProjects/study/WORK_20260416_JAVA_SOCKET_CONNECTION_RESET_TROUBLESHOOTING.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `analysis | design | explain | execute | refactor_docs`
- 작업 깊이: `full`
- 관련 요청:
  - 새 파일로 Java/Spring `Connection reset` troubleshooting 문서를 작성
- 원문 사용자 요청:
  - 새 파일로 진행
- 대상 경로 / 자산:
  - 신규 문서: `/Users/rody/VscodeProjects/study/jvm/java/java_socket_connection_reset.md`
- 참고 원재료: `/Users/rody/VscodeProjects/study/jvm/java/java_socket.md`
- 실행자: Codex
- 시작 일시: `2026-04-16 23:01:28 +0900`
- 종료 일시: `2026-04-16 23:29:27 +0900`
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `verify + commit`

## 1. Request Normalization

- goal:
  - broad한 `java_socket.md`를 덮어쓰지 않고, 범위가 정확히 맞는 troubleshooting 문서 `java_socket_connection_reset.md`를 새로 만든다.
- refs:
  - `/Users/rody/VscodeProjects/study/AGENTS.md`
  - `/Users/rody/VscodeProjects/study/PROJECT_INTENT.md`
  - `/Users/rody/VscodeProjects/study/USECASE.md`
  - `/Users/rody/.codex/skills/study-explanation/SKILL.md`
  - `/Users/rody/.codex/skills/study-explanation/references/deep-study-monograph.md`
  - `/Users/rody/.codex/skills/study-explanation/references/output-quality-gates.md`
  - `/Users/rody/VscodeProjects/study/jvm/java/java_socket.md`
  - `/Users/rody/VscodeProjects/study/jvm/java/java_http.md`
  - `/Users/rody/VscodeProjects/study/linux/errors/errno.md`
  - `/Users/rody/VscodeProjects/study/linux/commands/curl/curl_errors.md`
- scope:
  - 신규 Markdown 문서 작성
  - troubleshooting route 기준으로 문서 구조 설계
  - 관련 근거와 replay path 포함
  - self-audit
  - commit
- must_keep:
  - 기존 `java_socket.md`는 사용자 원재료로 보존
  - 새 문서는 제목, opening, 실제 범위가 정확히 맞아야 한다
  - assistant-presence scaffolding 없이 직접 진술형으로 연다

## 2. Root-First Framing

- 근본 문제:
  - 기존 `java_socket.md`는 제목은 넓지만 실제 내용은 `Connection reset` troubleshooting에 치우쳐 있어 scope drift가 있다.
- 왜 지금 중요한가:
  - 이번 proof round는 새로 추가한 troubleshooting route와 title/opening/scope alignment gate가 실제 문서 작성에 유효한지 검증해야 한다.
- 작업 목표:
  - 독자가 `java.net.SocketException: Connection reset`를 봤을 때 "무슨 일이 일어났는가"와 "어디서부터 의심해야 하는가"를 재설명할 수 있는 문서를 만든다.
- 기대 이점:
  - broad concept note와 fix memo 사이에서 흔들리지 않는 troubleshooting exemplar 후보를 만든다.
- 성공 정의:
  - 새 파일이 troubleshooting route를 충족한다
  - 직접 원인과 근본 원인이 분리된다
  - 수정이 통하는 이유와 replay path가 보인다
  - final review와 commit까지 닫힌다

## 3. Reader & Internalization Contract

- 주 독자:
  - Spring/Java 애플리케이션에서 외부 HTTP 호출 중 `Connection reset`를 만난 개발자
- 나중에 스스로 설명할 수 있어야 하는 것:
  - `Connection reset`이 Java 예외명보다 먼저 TCP의 `RST` 사건이라는 점
  - 왜 stale keep-alive, TLS 불일치, middlebox, 서버 크래시, 클라이언트 abort가 같은 예외로 보일 수 있는지
  - 어떤 관측이 direct cause이고, 무엇이 root cause 후보인지
  - 어떤 순서로 확인하면 되는지
- 특히 막아야 하는 오해:
  - `Connection reset == 서버가 무조건 죽었다`
  - `SocketException == Java 라이브러리 버그`
  - `timeout`, `connection refused`, `reset`이 같은 종류의 오류
- 목차 필요 여부와 이유:
  - 필요. 증상/메커니즘/원인/확인 순서를 독자가 빠르게 찾을 수 있어야 한다.
- 문서를 어떤 직접 진술형 문장으로 열 것인가:
  - `Java에서 Connection reset이 보이면, 먼저 Java가 아니라 TCP에서 누가 RST를 보냈는지를 의심해야 합니다.`
- 초반에 먼저 고정할 공통 메타데이터 / 구조 / artifact:
  - TCP 연결의 정상 종료(FIN)와 강제 종료(RST)
  - Spring/Java에서 표면화되는 예외 모양
  - 간단한 시퀀스 흐름
- 본격적인 실행 경로 / 동작 경로 추적이 시작되는 지점:
  - `클라이언트 요청 -> TCP/TLS/HTTP -> RST 발생 시 Java 예외 전파`

## 4. Scope Expansion & Impact Sync

- 조사한 경로:
  - `/Users/rody/VscodeProjects/study/jvm/java`
  - `/Users/rody/VscodeProjects/study/linux/errors`
  - `/Users/rody/VscodeProjects/study/linux/commands/curl`
  - `/Users/rody/VscodeProjects/study/computer_architecture/network`
- 함께 점검한 자산:
  - `java_socket.md`
  - `java_http.md`
  - `errno.md`
  - `curl_errors.md`
- 제외 표면과 근거:
  - 기존 `java_socket.md` 수정
    - 현재는 사용자 원재료이자 untracked scratch로 보이며, 이번 proof round의 핵심은 새 file scope 정합성과 troubleshooting route 검증이다.
  - hub/index update
    - 현재 검색 기준 새 파일을 참조해야 하는 existing index/hub link는 발견하지 못했다.

## 5. Evidence Ledger

- E-01
  - 주장:
    - `Connection reset`의 첫 의미는 Java보다 TCP `RST`다.
  - 근거 유형:
    - `repo evidence`, `official doc`
  - 자료:
    - `jvm/java/java_http.md`의 `RST vs FIN`
    - RFC 9293
- E-02
  - 주장:
    - Linux에서는 같은 사건이 `ECONNRESET`로 보일 수 있다.
  - 근거 유형:
    - `repo evidence`
  - 자료:
    - `linux/errors/errno.md`
    - `linux/commands/curl/curl_errors.md`
- E-03
  - 주장:
    - stale keep-alive와 커넥션 풀 청소는 Java/Spring troubleshooting에서 중요한 원인 축이다.
  - 근거 유형:
    - `repo evidence`
  - 자료:
    - `jvm/java/java_http.md`
    - 기존 `jvm/java/java_socket.md`

## 6. Design Decision

- 선택한 접근:
  - `troubleshooting route`를 따른 새 장문 문서 작성
- 고려한 대안:
  - 기존 `java_socket.md` 덮어쓰기
  - 기존 파일 상단에 일부 수정만 추가
- 대안을 채택하지 않은 이유:
  - 기존 파일은 제목/범위 mismatch가 크고, broad title을 유지한 채 수정하면 이번 proof round 목적이 흐려진다.
- 문서 구조:
  - 질문/직답
  - 먼저 알아야 할 `RST vs FIN`
  - Java/Spring에서 표면화되는 위치
  - direct cause vs root cause
  - 가장 흔한 원인군
  - 실제 확인 순서
  - 수정이 통하는 이유
  - 혼동쌍 비교
  - replay / verification
  - 관련 문서 / 참고 자료

## 7. Frozen Checklist

- [x] C-01 새 파일명이 제목과 실제 범위에 맞는다
- [x] C-02 opening이 assistant scaffolding 없이 직접 진술형이다
- [x] C-03 troubleshooting route의 `증상 -> 직접 원인 -> 근본 원인 -> 수정 이유 -> replay path`가 보인다
- [x] C-04 `RST`, `FIN`, `timeout`, `connection refused` 혼동이 줄어든다
- [x] C-05 Java/Spring 표면 증상과 TCP 사건 사이 메커니즘이 닫힌다
- [x] C-06 실무 확인 순서와 PASS/FAIL 관측이 보인다
- [x] C-07 기존 `java_socket.md`는 건드리지 않고 새 파일만 추가한다
- [x] C-08 final review + commit

## 8. Verification Plan

- target file self-audit against troubleshooting route gates
- `rg`로 assistant scaffolding, title/scope mismatch 여부 점검
- related path impact re-check
- commit only new file + WORK ledger

## 9. Final Audit

- self-audit 결과:
  - opening은 direct statement로 시작하고 assistant-presence scaffolding가 없다
  - scope는 `Java Socket 전체`가 아니라 `Connection reset troubleshooting`으로 명확히 좁아졌다
  - `RST -> 로컬 커널 오류 -> JDK SocketException -> Spring wrapper` 흐름이 문서 안에서 닫혔다
  - direct cause와 root cause를 분리했고, stale keep-alive / TLS / middlebox / server abort / client abort를 주요 원인군으로 정리했다
  - replay path에는 idle reuse 실험, `curl -v`, SSL debug, `tcpdump`가 포함된다
- verification 로그:
  - `rg`로 assistant scaffolding pattern 미검출
  - related path search 기준 새 파일을 반드시 함께 갱신해야 하는 hub/index는 발견하지 못함
  - 기존 `java_socket.md`는 수정하지 않음
- 남은 리스크:
  - 이번 라운드는 새 troubleshooting 문서 작성까지 닫았지만, 이 문서를 route-local benchmark candidate로 승격할지는 다음 benchmark 비교 라운드에서 따로 판단해야 한다
