# WORK_20260526_SOCKET_PROGRAMMING_INTERVIEW

## 0. Meta

- 작업 제목: 소켓 / 소켓 프로그래밍 인터뷰 총정리 문서 작성
- WORK 파일 경로: `docs/works/WORK_20260526_SOCKET_PROGRAMMING_INTERVIEW.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `explain | execute | refactor_docs`
- 작업 깊이: `full`
- 관련 요청: "소켓이란 뭔가요? 소켓 프로그래밍이란 뭔가요? 라는 질문 대비하여 예시 및 시각화 포함 총정리 문서 작성. 작성 후 검수 후 커밋 후 푸시."
- 대상 경로 / 자산: `interviews/socket-programming.md`, `interviews/README.md`
- 시작 일시: 2026-05-26
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `test+commit+push`

## 1. Request Normalization

- goal: 소켓과 소켓 프로그래밍 질문에 대해 면접장에서 먼저 짧게 답하고, 꼬리 질문에서는 커널/프로토콜/언어 API/운영 장애까지 내려갈 수 있는 정식 학습 문서를 만든다.
- refs: `interviews/source/interview_questions.md:6612-6819`, `interviews/network-web-protocols.md`, `interviews/linux-network-backend-runtime.md`, `computer_architecture/network/tcp_http.md`, Linux man-pages, RFC 9293, RFC 9110, Oracle Java API.
- scope: `study/interviews` 정식 학습 표면. 기존 dirty 파일은 직접 재작성하지 않고 새 standalone 문서와 README 링크로 영향 범위를 제한한다.
- mode: `execute`
- run_mode: `normal`
- finish: `test+commit+push`
- must_keep: 자연스러운 한국어, 짧은 면접 답변 + 깊은 메커니즘, 예시와 시각화, 공식 근거, dirty worktree 보호, path-limited staging.
- extra_checks: 문서 작성, 검수, commit, push 모두 필수.

### 1.1 Explicit Deliverables

- 새 총정리 문서 작성.
- "소켓이란 무엇인가"와 "소켓 프로그래밍이란 무엇인가"를 모두 답변.
- 예시 포함.
- 시각화 포함.
- 작성 후 검수.
- 커밋.
- 푸시.

### 1.2 Non-Goals

- 기존 `network-web-protocols.md` 원문 배치본 전체를 이번 작업에서 재작성하지 않는다. 해당 파일은 이미 사용자/기존 변경으로 dirty 상태이고, 이번 요청은 독립 총정리 문서 작성으로 닫을 수 있다.
- 실제 운영 수준 echo server 프로젝트를 새 예제 디렉터리로 만들지는 않는다. 문서 안의 복사 가능한 최소 실험으로 replay 경로를 제공한다.

## 2. Instruction Stack / Project Overlay

- global runtime: `~/.codex/AGENTS.md`, `~/.codex/AGENTS_WORK_TEMPLATE.md` 확인. 둘 다 `myai/ai/global/active/*` symlink.
- repo overlay: `study/AGENTS.md`, `study/AGENTS_WORK_TEMPLATE.md` 확인.
- nested overlay: `interviews/AGENTS.md`, `interviews/PROJECT_INTENT.md`, `interviews/USECASE.md` 확인.
- project facts: 인터뷰 자산은 짧은 직답과 깊은 꼬리 질문 대응을 함께 가져야 한다. `source/`는 raw reservoir이고 root 문서는 정식 학습 표면이다.
- conflict: 없음. 상위 규칙은 full loop, 근거, 검증, commit을 요구하고, nested overlay는 면접 답변 구조와 깊이를 요구한다.

## 3. Root-First Framing

- 근본 문제: 기존 raw curriculum에는 소켓 통신 과정이 있지만, 정의와 흐름이 길게 섞여 있어 면접용 첫 답변, 리스닝 소켓/연결 소켓 구분, TCP 스트림/프레이밍, 장애 관측까지 한 번에 복원하기 어렵다.
- 작업 목표: standalone 문서 하나에서 "짧게 말하기 -> 계층별 메커니즘 -> 코드 예시 -> 시각 trace -> 실패 모드 -> replay"가 닫히게 만든다.
- 성공 정의: 새 문서가 socket/port/fd/connection/request를 분리하고, `socket-bind-listen-accept`와 `socket-connect-send-recv-close`를 시각화하며, Java 예시와 직접 확인 실험을 포함하고, 공식 근거와 주변 문서 링크를 가진다.
- PARTIAL 조건: 문서는 있으나 검증, 링크, commit, push 중 하나가 비어 있음.
- BLOCKED 조건: 파일 쓰기, commit, push 권한 또는 네트워크 오류로 requested closure를 닫을 수 없음.

## 4. Reader & Internalization Contract

- 주 독자: 백엔드 경력 인터뷰를 준비하는 개발자.
- teach-back 목표: "소켓은 커널 네트워크 스택의 통신 끝점이고, 소켓 프로그래밍은 그 끝점의 생명주기와 실패 조건을 코드로 관리하는 일"이라고 말한 뒤, 리스닝 소켓과 연결 소켓의 차이를 그림으로 설명할 수 있어야 한다.
- learner gap: 소켓을 포트/IP 조합 자체로 오해하거나, `accept()`가 새 포트를 만든다고 생각하거나, TCP의 `send`와 `recv` 호출 단위가 보존된다고 생각하는 위험.
- first brick: `int fd = socket(AF_INET, SOCK_STREAM, 0);`
- guided trace: fd table -> kernel socket object, server lifecycle, accept queue, TCP stream framing.
- misconception repair: port/socket/fd/request 분리, `accept()` 새 fd vs 새 port, TCP stream vs message.
- active recall / replay: Python localhost echo 실험과 `lsof`/`ss` 관측.
- primary exemplar: `computer_architecture/threads/threads.md`
- reference principle: 하나의 개념을 OS, 런타임, 애플리케이션, 운영 관측으로 연결한다.
- trait to avoid: taxonomy breadth를 늘려 질문 중심성을 잃는 방식.
- secondary exemplar: `git/git_rebase.md`
- reference principle: before/after와 작은 시나리오로 오해 비용을 줄인다.
- trait to avoid: 누적된 명령/사례가 공통 모델보다 커지는 방식.
- target quality lift: raw socket section보다 더 빠른 first brick, 더 명확한 accept/port visual, 더 구체적인 replay path.

## 5. Scope Expansion & Impact Sync

- 시작 키워드: 소켓, socket, socket programming.
- 확장 키워드: fd, file descriptor, port, bind, listen, accept, connect, send, recv, TCP stream, UDP datagram, ServerSocket, Connection reset, epoll, backlog, 4-tuple.
- 조사한 경로:
  - `interviews/source/interview_questions.md:6612-6819`
  - `interviews/network-web-protocols.md:696-1160`
  - `interviews/linux-network-backend-runtime.md` socket/bind/listen/accept sections
  - `computer_architecture/network/tcp_http.md`
  - `jvm/java/java_socket_connection_reset.md`
- 함께 움직여야 하는 표면: 새 정식 문서, `interviews/README.md` navigation, WORK ledger.
- 제외 표면과 근거: `network-web-protocols.md`는 현재 dirty raw curriculum 문서이고 전체 생성 산출물 성격이 강해 이번 작업에서 직접 수정하지 않는다.

## 6. Evidence Ledger

- E-01
  - 주장: `socket()`은 통신 endpoint를 만들고 그 endpoint를 가리키는 fd를 반환한다.
  - 근거 유형: official doc
  - 자료: Linux `socket(2)`.
  - 상태: APPLY.
- E-02
  - 주장: `accept()`는 listening socket의 pending connection queue에서 하나를 꺼내 새 connected socket fd를 반환하며 original socket은 영향을 받지 않는다.
  - 근거 유형: official doc
  - 자료: Linux `accept(2)`.
  - 상태: APPLY.
- E-03
  - 주장: TCP는 Internet protocol stack의 transport-layer protocol이고, 오래 진화한 표준이며 TCP 자체는 암호화 기능을 제공하지 않아 TLS 같은 상위 계층이 필요하다.
  - 근거 유형: standard
  - 자료: RFC 9293.
  - 상태: APPLY.
- E-04
  - 주장: HTTP 의미는 TCP 연결 위의 애플리케이션 의미이며 HTTP 버전별 transport와 메시지 문법은 분리해서 봐야 한다.
  - 근거 유형: standard
  - 자료: RFC 9110.
  - 상태: APPLY.
- E-05
  - 주장: Java `ServerSocket`은 서버 소켓을 구현하고 네트워크 요청을 기다리는 abstraction이다.
  - 근거 유형: official doc
  - 자료: Oracle Java SE 26 API `ServerSocket`.
  - 상태: APPLY.

## 7. Design Decision

- 선택한 접근: `interviews/socket-programming.md` standalone monograph + README 링크.
- 고려한 대안:
  - A: `network-web-protocols.md` 기존 socket section을 직접 대체한다.
  - B: standalone 정식 문서를 만들고 README에서 진입시킨다.
- 채택: B.
- 이유: 기존 파일은 raw curriculum 성격과 dirty 상태가 강하고, 사용자는 총정리 문서를 요청했다. 독립 문서가 면접 대비 구조와 검증을 더 명확히 제공한다.
- 문서 뼈대: 짧은 직답 -> 질문 의도 -> fd/kernel first brick -> 용어 경계 -> server/client lifecycle -> stream framing -> programming responsibilities -> Java example -> concurrency -> UDP contrast -> failure modes -> experiment -> tail questions -> refs.

## 8. Frozen Checklist

- C-01 사용자: 새 총정리 문서가 존재한다.
  - PASS: `interviews/socket-programming.md` 생성.
- C-02 사용자: 소켓 정의와 소켓 프로그래밍 정의가 모두 초반에 직접 답변된다.
  - PASS: "면접에서 먼저 말할 답"에 둘 다 포함.
- C-03 사용자: 예시와 시각화가 포함된다.
  - PASS: Java 예시, Python 실험, ASCII diagrams/tables 포함.
- C-04 AI 추가: 리스닝 소켓/연결 소켓, 포트/fd/socket/request, TCP stream/message 경계를 분리한다.
  - PASS: 각 구분을 본문에서 명시하고 오해를 반박.
- C-05 AI 추가: 공식 근거와 주변 저장소 문서를 연결한다.
  - PASS: 근거 섹션과 README 링크.
- C-06 사용자: 검수한다.
  - PASS: Markdown structure, code fences, link/path, forbidden assistant residue, example smoke verification.
- C-07 사용자: commit한다.
  - PASS: 이번 변경 파일만 staging하여 commit.
- C-08 사용자: push한다.
  - PASS: push 전 outgoing commits 확인 후 `git push`.

## 9. Verification Plan

- `git diff -- interviews/socket-programming.md interviews/README.md docs/works/WORK_20260526_SOCKET_PROGRAMMING_INTERVIEW.md`
- Markdown sanity: heading/opening/TOC anchor consistency, even code fences, no assistant residue.
- Example smoke: localhost Python socket echo script returns expected echo.
- Path-limited diff check: `git diff --check -- interviews/socket-programming.md interviews/README.md docs/works/WORK_20260526_SOCKET_PROGRAMMING_INTERVIEW.md`
- Git closure: path-limited staging, commit, outgoing log check, push.

## 10. Final Audit Placeholder

## 10. Verification Result

- Markdown TOC / anchor check: PASS. `headings=25`, `toc_links=17`, `missing=[]`.
- Code fence parity: PASS. `interviews/socket-programming.md` fence markers `34`, even.
- Local relative link check: PASS. `linux-network-backend-runtime.md`, `network-web-protocols.md`, `../jvm/java/java_socket_connection_reset.md` all exist.
- Forbidden assistant residue check: PASS for anchored patterns. No `좋습니다` opening, `설명드릴게요`, `이제 ... 보겠습니다`, `TODO`, `FIXME`.
- Whitespace diff check: PASS. `git diff --check -- interviews/socket-programming.md interviews/README.md docs/works/WORK_20260526_SOCKET_PROGRAMMING_INTERVIEW.md`.
- Python localhost socket smoke: PASS after sandbox escalation. Output `echo: hello`.
- Java API snippet sanity: PASS after sandbox escalation. JShell on `javac 25.0.2` accepted `new PrintWriter(System.out, true, StandardCharsets.UTF_8)` and printed `printwriter-ok`.

## 11. Final Audit & Closure

- intent-fit review: PASS. The document directly answers both interview questions, includes a short first answer, deep mechanism, examples, ASCII visualizations, failure modes, and replay path.
- expert-perspective review: PASS. The document separates socket, port, fd, TCP connection, HTTP request, listening socket, connected socket, TCP stream, and application framing. It avoids the known false explanation that `accept()` creates a new server port.
- impact sync: PASS. Added README entry under promoted deep documents. Did not rewrite dirty raw curriculum documents.
- downstream impact gate: PASS. Commit will stage only `interviews/socket-programming.md`, `interviews/README.md`, and this WORK file. Push will send the existing branch's ahead commits plus this new commit; outgoing commits are checked before push.
- remaining risks: Existing repository dirty state is unrelated and intentionally preserved.
- requested closure scope: write document -> review -> commit -> push.
- achieved closure scope before git closure: write document + review complete; commit/push pending in execution.

### 11.1 Checklist Re-Judgement

- C-01: PASS.
- C-02: PASS.
- C-03: PASS.
- C-04: PASS.
- C-05: PASS.
- C-06: PASS.
- C-07: PASS evidence is the git commit created after this ledger is staged.
- C-08: PASS evidence is the git push result reported after the commit.

### 11.2 Final State

- 최종 상태: `COMPLETE` for document/review stage.
- 완료 게이트: `ALLOW_COMPLETE`; git commit/push evidence is recorded in the final response because the commit hash cannot be embedded in the commit content without changing that hash.

## 12. Follow-Up: OS-Neutral Socket Boundary

- 사용자 피드백: "os 중립적인 설명도 있는지? os마다 다양하게 소켓을 지원할 거 같은데."
- 판정: 기존 문서는 핵심 개념은 통신 끝점으로 설명했지만, first brick과 근거가 Unix/Linux/POSIX fd 모델에 강하게 기대고 있어 OS 중립 경계가 덜 선명했다.
- 수정: `interviews/socket-programming.md` 앞부분에 `운영체제 중립 모델과 OS별 API 차이` 섹션을 추가하고, 기존 첫 번째 벽돌 제목을 POSIX 계열 대표 예시로 좁혔다.
- 근거: POSIX/The Open Group `socket()`은 fd를 반환하지만, Microsoft Winsock `socket()`은 `SOCKET` 핸들을 반환하고 `closesocket()`을 쓰며, Apple Network framework는 `NWConnection` 같은 더 높은 수준의 연결 객체를 제공한다.
- 검증 예정: TOC anchor check, diff check, OS-neutral terms scan, scoped commit/push.

## 13. Follow-Up: Korean Flow Pass

- 사용자 스킬 지시: `$humanize-korean $study-explanation`.
- 판정: 문서의 전체 학습 구조는 유지하되, OS 중립 보강부에 남아 있던 `OS networking API`, `socket endpoint handle 반환`, `OS network stack` 같은 영어 우선 표현은 한국어 학습 문서의 초반 발판으로는 약했다.
- 수정: 해당 구간을 `운영체제 네트워크 API`, `통신 끝점을 가리키는 손잡이`, `운영체제 네트워크 스택`처럼 한국어 우선 표현으로 바꾸고, 첫 답변의 `OS 중립적으로` 문장도 `특정 운영체제에 묶지 않고`로 풀었다.
- 검증 결과:
  - TOC anchor check: PASS. `headings=26`, `toc_links=18`, `missing=[]`.
  - Whitespace diff check: PASS. `git diff --check -- interviews/socket-programming.md docs/works/WORK_20260526_SOCKET_PROGRAMMING_INTERVIEW.md`.
  - 영어 우선 잔여 표현 scan: PASS for target doc. `OS networking API`, `socket endpoint handle`, `OS network stack`, `OS 중립적으로`, `불변식` 잔여 없음.
- 최종 감사: PASS. 기술 주장과 공식 근거는 그대로 유지했고, 한국어 흐름과 초반 학습 발판만 좁게 보정했다.

## 14. Follow-Up: Humanize Correction on `손잡이`

- 사용자 피드백: "`정수 손잡이로 보이는 경우가 많고` - `$humanize-korean` 적용된 게 맞는지?"
- 판정: 해당 표현은 기술 의미를 쉽게 풀려다 만든 비유지만, 한국어 학습 문서에서는 어색한 직역처럼 읽힌다. `humanize-korean` 기준의 자연스러움 검수에서 놓친 항목으로 본다.
- 수정: `손잡이` 계열 표현을 문맥에 따라 `정수값`, `식별자`, `값이나 객체`, `그 소켓을 가리키는 번호`, `접근 방식`으로 교체했다. 기술 claim은 유지하고 표현만 보정했다.
- 검증 결과:
  - Target phrase scan: PASS for target doc. `손잡이`, `정수 손잡이`, `보이는 경우가 많고`, `운영체제에 맡긴 통로` 잔여 없음.
  - TOC anchor check: PASS. `headings=26`, `toc_links=18`, `missing=[]`.
  - Whitespace diff check: PASS. `git diff --check -- interviews/socket-programming.md docs/works/WORK_20260526_SOCKET_PROGRAMMING_INTERVIEW.md`.
- 최종 감사: PASS. 사용자 지적을 맞는 지적으로 인정하고, 자연스럽지 않은 비유를 기술적으로 더 정확한 표현으로 바꿨다.

## 15. Follow-Up: Whole-Document Humanize Pass

- 사용자 피드백: "비슷하게 억지 번역 없는지 등 다시 전체 `$humanize-korean`."
- Writing Task Frame:
  - 요청 유형: 전체 문서 재윤문 및 억지 번역 감사.
  - 독자: 기술 면접을 준비하는 한국어 독자.
  - 장르: 인터뷰 대비용 장문 기술 학습 문서.
  - 성공 기준: 기술 claim과 근거는 보존하되, 어색한 직역, 내부 은유, 덜 정리된 영어 표현을 줄인다.
  - 비목표: 모든 영어 기술 용어를 억지로 한국어로 번역하지 않는다.
- 탐지:
  - `첫 번째 벽돌`, `밑바닥 질문`, `raw 파일 디스크립터`, `connected socket`, `root cause`, `길이 prefix`, `구분자 delimiter`, `QUIC의 기반 전송`, `표면화됩니다`, `PASS 신호 / FAIL 신호` 등은 기술 의미는 맞지만 한국어 문서 흐름에서 튀는 표현으로 판정했다.
  - `endpoint`는 억지 번역을 피하기 위해 `엔드포인트(endpoint)`로 첫 등장에 병기하고, 이후에는 `엔드포인트` 또는 문맥상 `통신 지점`으로 정리했다.
- 수정:
  - 문서 내부 은유를 `가장 작은 출발점`, `근본 질문`처럼 직접적인 표현으로 교체했다.
  - 영어 상태명과 API 표현은 공식 용어가 필요한 곳만 남기고, 문장 설명에서는 `리스닝 소켓`, `연결 소켓`, `오류 메시지`, `근본 원인`처럼 한국어 우선으로 바꿨다.
  - 프레이밍 설명에서는 `길이를 앞에 붙이는 방식(length prefix)`, `구분자(delimiter)`처럼 한국어 설명을 먼저 두고 검색 가능한 원어를 병기했다.
- 검증 결과:
  - 억지 번역 후보 scan: PASS for target doc. `손잡이`, `밑바닥`, `첫 번째 벽돌`, `root cause`, `error message`, `connected socket`, `listening socket`, `idle 연결`, `raw 파일`, `기반 전송`, `통신 끝점`, `길이 prefix`, `구분자 delimiter`, `read exactly` 잔여 없음.
  - TOC anchor check: PASS. `headings=26`, `toc_links=18`, `missing=[]`.
  - Code fence parity: PASS. `code_fence_markers=36`, even.
  - Whitespace diff check: PASS. `git diff --check -- interviews/socket-programming.md docs/works/WORK_20260526_SOCKET_PROGRAMMING_INTERVIEW.md`.
- 최종 감사: PASS. 기술 용어를 무리하게 번역하지 않고, 문장 안에서 튀던 직역과 내부 은유만 줄였다.
