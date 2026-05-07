# WORK_20260507_CORE_INTERVIEW_WEB_REQUEST_SOCKET_FLOW

## 0. Meta

- 작업 제목: core interview guide 웹 요청 소켓 흐름 보강
- WORK 파일 경로: `docs/works/WORK_20260507_CORE_INTERVIEW_WEB_REQUEST_SOCKET_FLOW.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `refactor_docs | explain`
- 작업 깊이: `full`
- 관련 요청: `interviews/core-interview-guide.md` 1번 웹 요청 부분에서 full-duplex, NIC, 커널, Nginx, Tomcat, socket/accept/handshake 흐름을 질문이 남지 않게 정리
- 대상 경로 / 자산: `interviews/core-interview-guide.md`
- 실행자: Codex
- 시작 일시: 2026-05-07 21:06 KST
- 종료 일시: 2026-05-07 21:10 KST
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `test+commit`

## 1. Request Normalization

- goal: 웹 요청 1번 섹션에서 서버 수신 경로를 낮은 계층부터 사용자 공간 서버까지 순서대로 설명한다.
- refs: `core-interview-guide.md` 1번 섹션, TCP/RFC, Linux socket API, Nginx reverse proxy, Tomcat request flow.
- scope: `interviews/core-interview-guide.md` 1번 질문 섹션과 이 WORK ledger.
- mode: `execute`
- run_mode: `normal`
- finish: 문서 편집, 검증, 최종 감사, scoped commit.
- must_keep:
  - 인터뷰 답변용 30초 직답과 이어 말할 순서 구조를 유지한다.
  - 사용자가 궁금해한 `full-duplex`, `Content-Length`, `NIC`, `kernel`, `accept`, `ephemeral port`, `Nginx`, `Tomcat` 흐름을 빠뜨리지 않는다.
  - "소켓을 연다"라는 표현을 리스닝 소켓, 연결별 소켓, upstream outbound socket으로 나누어 설명한다.
- extra_checks:
  - 문서가 "NIC가 handshake를 한다"거나 "서버가 무작위 포트로 새 소켓을 연다"는 오해를 만들지 않는지 확인한다.
  - 관련 로컬 자산과 공식 근거를 구분한다.

## 2. Root-First Framing

- 근본 문제: 기존 1번 섹션은 TCP/TLS/DNS 설명은 강하지만, 서버 수신 쪽에서 누가 소켓을 만들고 누가 핸드셰이크를 수행하는지의 ownership이 더 선명해야 한다.
- 왜 중요한가: 면접에서는 "웹 요청 하나" 질문이 곧 OS, 커널 네트워킹, 프록시, WAS, Spring 실행 경로 꼬리 질문으로 이어진다.
- 작업 목표: 한 번 읽으면 `클라이언트 요청 byte -> NIC -> 커널 TCP/IP stack -> listening socket / connected socket -> Nginx -> upstream socket -> Tomcat -> Spring` 흐름을 자기 말로 다시 설명할 수 있게 한다.
- 성공 정의:
  - 질문에 나온 모든 의문이 본문 또는 꼬리 질문 지도에서 닫힌다.
  - 공식 근거 기반의 load-bearing claim만 확정형으로 쓴다.
  - Markdown 구조와 링크가 깨지지 않는다.
- PARTIAL 조건: 문서 보강은 되었지만 근거나 검증이 비어 있음.
- BLOCKED 조건: 대상 파일 변경이나 검증을 수행할 수 없음.

## 3. Reader & Internalization Contract

- 주 독자: 기술 인터뷰에서 웹 요청 경로를 계층별로 설명해야 하는 백엔드 개발자.
- 독자가 다시 설명할 수 있어야 하는 것:
  - TCP 연결의 당사자는 애플리케이션 이름이 아니라 양쪽 커널 TCP 구현이라는 점.
  - NIC는 frame 이동과 queue/offload를 돕지만 HTTP/TLS/Tomcat 요청 판단을 하지 않는다는 점.
  - `accept()`는 리스닝 소켓을 바꾸는 호출이 아니라 연결별 file descriptor를 받는 호출이라는 점.
  - 서버의 서비스 포트는 유지되고, 무작위처럼 보이는 port는 active opener의 ephemeral source port라는 점.
- 막아야 하는 오해:
  - "Content-Length만큼 커널이 먼저 전부 받아 놓은 뒤 서버가 처리한다."
  - "Nginx나 Tomcat이 TCP 3-way handshake packet을 직접 주고받는다."
  - "accept할 때 서버가 새 무작위 포트를 연다."
- primary exemplar: `computer_architecture/threads/threads.md`
- reference principle: 한 개념을 hardware -> kernel -> runtime/server -> framework로 연결한다.
- trait to avoid: 계층 taxonomy만 늘리고 실제 byte 흐름과 ownership을 흐리게 만드는 방식.
- target quality lift: 사용자의 구체 의문을 꼬리 질문 없이도 본문 순서만 따라가면 해소되게 만든다.

## 4. Project Overlay

- 적용한 로컬 AGENTS 경로: `/Users/rody/VscodeProjects/study/AGENTS.md`, `/Users/rody/VscodeProjects/study/interviews/PROJECT_INTENT.md`, `/Users/rody/VscodeProjects/study/interviews/USECASE.md`
- 활성화한 프로젝트 규칙:
  - 현재 평균이 아니라 exemplar 수준으로 설명한다.
  - 인터뷰 답변은 짧은 직답과 깊은 메커니즘을 함께 가져야 한다.
  - 근거, 메커니즘, 검증 경로를 닫는다.
- 전역 규칙과의 충돌 여부 / 해소: 없음. 런타임상 delegation은 사용자 명시 요청이 없어 사용하지 않고, single-agent separated audit lane으로 보완한다.

## 5. Scope Expansion & Impact Sync

- 시작 키워드: 웹 요청, full-duplex, socket, accept, NIC, kernel, Nginx, Tomcat.
- 확장 키워드: TCP 3-way handshake, listening socket, connected socket, ephemeral port, Content-Length, receive buffer, accept queue, SYN backlog, proxy_pass, Coyote adapter.
- 조사한 경로:
  - `interviews/core-interview-guide.md`
  - `interviews/network-web-protocols.md`
  - `web/massive_connections.md`
  - `web/http_connection_flow.md`
  - `network/network_tpcip.md`
- 함께 움직여야 하는 표면: 이번 요청은 실전 guide의 1번 섹션 보강이므로 세부 deep doc의 전면 재작성은 제외한다.
- 제외 표면과 근거: 다른 인터뷰 대주제 문서는 1번 섹션 링크로 충분하며, 이번 질문의 직접 결함은 core guide 안의 수신 경로 설명 밀도다.

## 6. Evidence Ledger

- E-01
  - 주장: TCP는 신뢰성 있는 순서 보장 byte stream이고, 데이터 흐름은 양방향으로 지원된다.
  - 근거 유형: `standard`
  - 자료: RFC 9293.
  - 상태: `APPLY`
- E-02
  - 주장: 서버가 연결을 받으려면 socket 생성, bind, listen, accept 흐름이 필요하며, Linux에서 accept는 pending queue의 연결 요청을 꺼내 새 connected socket fd를 돌려준다.
  - 근거 유형: `official API`
  - 자료: Linux man-pages `socket(2)`, `listen(2)`, `accept(2)`.
  - 상태: `APPLY`
- E-03
  - 주장: HTTP/1.1 메시지는 start-line, header, 빈 줄, optional body로 구성되며, body가 있으면 정해진 길이만큼 stream에서 읽는다.
  - 근거 유형: `standard`
  - 자료: RFC 9112.
  - 상태: `APPLY`
- E-04
  - 주장: Nginx는 `proxy_pass`로 처리한 요청을 지정한 upstream HTTP 서버 주소와 port로 전달할 수 있다.
  - 근거 유형: `official doc`
  - 자료: NGINX reverse proxy documentation.
  - 상태: `APPLY`
- E-05
  - 주장: Tomcat 요청은 endpoint에서 시작해 protocol, Coyote adapter를 거쳐 servlet request processing으로 들어간다.
  - 근거 유형: `official doc`
  - 자료: Apache Tomcat request process flow.
  - 상태: `APPLY`

## 7. Frozen Checklist

- [x] `Content-Length`와 TCP byte stream 관계를 설명한다.
- [x] 리스닝 소켓, 연결별 소켓, upstream outbound socket을 구분한다.
- [x] TCP 3-way handshake 당사자를 kernel TCP stack 중심으로 설명한다.
- [x] NIC와 kernel TCP/IP stack의 책임을 분리한다.
- [x] Nginx와 Tomcat 사이에 별도 TCP 연결이 생길 수 있음을 설명한다.
- [x] Tomcat connector에서 servlet/Spring으로 넘어가는 흐름을 보강한다.
- [x] 꼬리 질문 지도에 사용자의 의문을 직접 닫는 항목을 추가한다.
- [x] 공식 근거와 확인 명령을 유지하거나 보강한다.
- [x] Markdown 구조와 내부 링크를 검증한다.
- [x] scoped diff만 commit한다.

## 8. Claim / Reasoning Ledger

- Claim: 서버가 accept할 때 새 무작위 서버 port를 여는 것이 아니다.
  - Reason: 서비스 port에 bind/listen한 리스닝 소켓이 있고, 연결별 socket은 같은 local server port와 client IP/port를 포함한 4-tuple로 구분된다. 무작위 port는 active opener의 source port에서 주로 나타난다.
  - Evidence: Linux `listen(2)`/`accept(2)`, RFC 9293 port multiplexing.
  - Counterexample check: Nginx가 upstream에 연결할 때는 Nginx가 active opener가 되어 ephemeral source port를 쓴다. 이는 client-facing accepted connection과 다른 별도 connection이다.
  - Support tier: T1
  - Admission lane: APPLY
- Claim: TCP 3-way handshake packet 처리의 직접 주체는 보통 양쪽 커널 TCP/IP stack이다.
  - Reason: 애플리케이션은 `connect` 또는 `listen/accept` API로 의도를 표현하고, TCP state machine과 SYN/SYN-ACK/ACK 처리는 커널이 맡는다.
  - Evidence: RFC 9293 connection establishment, Linux socket API.
  - Counterexample check: user-space TCP stack/DPDK 같은 특수 구조는 예외지만 일반 Nginx/Tomcat 서버 설명의 기본값은 커널 TCP/IP stack이다.
  - Support tier: T1
  - Admission lane: APPLY
- Claim: `Content-Length`는 HTTP parser가 message body 경계를 판단하는 정보이지, TCP가 message 전체를 한 번에 전달한다는 뜻이 아니다.
  - Reason: TCP는 byte stream이고 HTTP parser가 header 뒤 body length를 판단해 stream에서 필요한 만큼 읽는다.
  - Evidence: RFC 9293, RFC 9112.
  - Counterexample check: chunked transfer, HTTP/2 frame, streaming upload에서는 body boundary 표현이 달라질 수 있다.
  - Support tier: T1
  - Admission lane: APPLY

## 9. Verification Plan

- Markdown heading anchor와 중복 heading 구조가 깨지지 않았는지 확인한다.
- `rg`로 핵심 용어가 새 설명 안에 모두 들어갔는지 확인한다.
- `git diff --check -- interviews/core-interview-guide.md docs/works/WORK_20260507_CORE_INTERVIEW_WEB_REQUEST_SOCKET_FLOW.md`를 실행한다.
- 최종 감사에서 사용자 질문별 답변 여부를 재판정한다.

## 10. Verification Log

- `git diff --check -- interviews/core-interview-guide.md docs/works/WORK_20260507_CORE_INTERVIEW_WEB_REQUEST_SOCKET_FLOW.md`: PASS.
- `rg` 핵심 용어 확인: `Content-Length`, `accept()`, `ephemeral port`, `TCP 3-way handshake`, `NIC`, `Coyote adapter`, `backpressure`, `upstream keep-alive` 모두 대상 문서 또는 WORK에 존재.
- 변경 범위 확인: `interviews/core-interview-guide.md`, `docs/works/WORK_20260507_CORE_INTERVIEW_WEB_REQUEST_SOCKET_FLOW.md`만 이번 작업 범위로 확인.

## 11. Final Audit

- 사용자 질문 재판정:
  - full-duplex와 요청/응답 byte stream 관계: PASS.
  - `Content-Length`만큼 모든 byte를 커널이 먼저 받는지 여부: PASS.
  - NIC, driver, kernel TCP/IP stack, Nginx/Tomcat 책임 경계: PASS.
  - socket 생성 시점, listening socket, connected socket, accept 시점: PASS.
  - 무작위 port처럼 보이는 ephemeral source port 설명: PASS.
  - Nginx upstream 연결과 Tomcat connector/Coyote/Spring MVC 흐름: PASS.
- Critic lane:
  - 반론: "너무 낮은 계층 설명이 늘어나면 인터뷰 guide의 빠른 복원 역할을 해칠 수 있다."
  - 수리: `첫 30초 답변` 구조는 유지하고, 상세 흐름은 1번 섹션의 `이어 말할 순서` 안에 배치했다. 꼬리 질문 지도에는 사용자의 구체 의문만 직접 추가했다.
- Residual risk:
  - 세부 OS 구현은 Linux 중심으로 설명했다. 일반 사용자 공간 TCP stack, DPDK, L4 proxy passthrough, TLS passthrough 같은 특수 구성은 기본 경로의 예외로 남는다.
- Whole-request verdict: `WHOLE_COMPLETE`.
- Remaining open child work in requested scope: none.
- Unrelated open work disclosure: 저장소에는 기존 다른 파일 변경과 미추적 파일이 많지만, 이번 scoped commit 대상은 위 두 파일뿐이다.
